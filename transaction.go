// SPDX-FileCopyrightText: 2020 SAP SE
// SPDX-FileCopyrightText: 2021 SAP SE
// SPDX-FileCopyrightText: 2022 SAP SE
// SPDX-FileCopyrightText: 2023 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

package ase

//#include "ctlib.h"
import "C"
import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"unsafe"

	"github.com/SAP/go-dblib"
)

// Interface satisfaction checks.
var _ driver.Tx = (*transaction)(nil)

// transaction implements the driver.Tx interface.
type transaction struct {
	conn *Connection
	// ASE does not support read-only transactions - the connection
	// itself however can be set as read-only.
	// readonlyPreTx signals if a connection was marked as read-only
	// before the transaction startet.
	readonlyPreTx C.CS_INT
	// readonlyNeedsReset is true if the read-only option passed to
	// BeginTx differs from the read-only property of the connection.
	readonlyNeedsReset bool
}

// Begin implements the driver.Conn interface.
func (conn *Connection) Begin() (driver.Tx, error) {
	return conn.beginTx(driver.TxOptions{Isolation: driver.IsolationLevel(sql.LevelDefault), ReadOnly: false})
}

// BeginTx implements the driver.ConnBeginTx interface.
func (conn *Connection) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	recvTx := make(chan driver.Tx, 1)
	recvErr := make(chan error, 1)
	go func() {
		tx, err := conn.beginTx(opts)
		recvTx <- tx
		close(recvTx)
		recvErr <- err
		close(recvErr)
	}()

	for {
		select {
		case <-ctx.Done():
			// Context exits early, Tx will still be created and
			// initialized; read and rollback
			go func() {
				tx := <-recvTx
				if tx != nil {
					tx.Rollback()
				}
			}()
			return nil, ctx.Err()
		case err := <-recvErr:
			tx := <-recvTx
			return tx, err
		}
	}
}

// TODO: Add doc
func (conn *Connection) beginTx(opts driver.TxOptions) (driver.Tx, error) {
	isolationLevel, err := dblib.ASEIsolationLevelFromGo(sql.IsolationLevel(opts.Isolation))
	if err != nil {
		return nil, err
	}

	tx := &transaction{conn, C.CS_FALSE, false}

	if _, err := tx.conn.Exec("BEGIN TRANSACTION", nil); err != nil {
		return nil, fmt.Errorf("Failed to start transaction: %w", err)
	}

	if _, err := tx.conn.Exec(fmt.Sprintf("SET TRANSACTION ISOLATION LEVEL %d", isolationLevel), nil); err != nil {
		return nil, fmt.Errorf("Failed to set isolation level for transaction: %w", err)
	}

	var currentReadOnly C.CS_INT = C.CS_FALSE
	if retval := C.ct_con_props(tx.conn.conn, C.CS_GET, C.CS_PROP_READONLY, unsafe.Pointer(&currentReadOnly), C.CS_UNUSED, nil); retval != C.CS_SUCCEED {
		return nil, makeError(retval, "Failed to retrieve readonly property")
	}

	var targetReadOnly C.CS_INT = C.CS_FALSE
	if opts.ReadOnly {
		targetReadOnly = C.CS_TRUE
	}

	if currentReadOnly != targetReadOnly {
		tx.readonlyPreTx = currentReadOnly
		tx.readonlyNeedsReset = true

		if err := tx.setRO(targetReadOnly); err != nil {
			return nil, err
		}
	}

	return tx, nil
}

// Commit implements the driver.Tx interface.
func (tx *transaction) Commit() error {
	if _, err := tx.conn.Exec("COMMIT TRANSACTION", nil); err != nil {
		return err
	}
	return tx.finish()
}

// Rollback implements the driver.Tx interface.
func (tx *transaction) Rollback() error {
	if _, err := tx.conn.Exec("ROLLBACK TRANSACTION", nil); err != nil {
		return err
	}
	return tx.finish()
}

// finish finishes the transaction.
func (tx *transaction) finish() error {
	if tx.readonlyNeedsReset {
		if err := tx.setRO(tx.readonlyPreTx); err != nil {
			return err
		}
	}
	tx.conn = nil
	return nil
}

// setRO sets the transaction as Read-Only.
func (tx *transaction) setRO(ro C.CS_INT) error {
	retval := C.ct_con_props(tx.conn.conn, C.CS_SET, C.CS_PROP_READONLY, unsafe.Pointer(&ro), C.CS_UNUSED, nil)
	if retval != C.CS_SUCCEED {
		return makeError(retval, "Failed to set readonly")
	}

	return nil
}
