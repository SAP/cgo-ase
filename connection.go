// SPDX-FileCopyrightText: 2020 SAP SE
// SPDX-FileCopyrightText: 2021 SAP SE
// SPDX-FileCopyrightText: 2022 SAP SE
// SPDX-FileCopyrightText: 2023 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

package ase

//#include <stdlib.h>
//#include "ctlib.h"
import "C"
import (
	"context"
	"database/sql/driver"
	"fmt"
	"io"
	"unsafe"

	"github.com/SAP/go-dblib"
	"github.com/SAP/go-dblib/asetypes"
)

// Interface satisfaction checks.
var (
	_ driver.Conn               = (*Connection)(nil)
	_ driver.ConnBeginTx        = (*Connection)(nil)
	_ driver.ConnPrepareContext = (*Connection)(nil)
	_ driver.Execer             = (*Connection)(nil)
	_ driver.ExecerContext      = (*Connection)(nil)
	_ driver.Pinger             = (*Connection)(nil)
	_ driver.Queryer            = (*Connection)(nil)
	_ driver.QueryerContext     = (*Connection)(nil)
	_ driver.NamedValueChecker  = (*Connection)(nil)
)

// Connection implements the driver.Conn interface.
type Connection struct {
	conn      *C.CS_CONNECTION
	driverCtx *csContext
}

// NewConnection allocates a new connection based on the
// options in the dsn.
//
// If driverCtx is nil a new csContext will be initialized.
func NewConnection(driverCtx *csContext, info *Info) (*Connection, error) {
	if driverCtx == nil {
		var err error
		driverCtx, err = newCsContext(info)
		if err != nil {
			return nil, fmt.Errorf("Failed to initialize context for conn: %w", err)
		}
	}

	if err := driverCtx.newConn(); err != nil {
		return nil, fmt.Errorf("Failed to ensure context: %w", err)
	}

	conn := &Connection{
		driverCtx: driverCtx,
	}

	if retval := C.ct_con_alloc(driverCtx.ctx, &conn.conn); retval != C.CS_SUCCEED {
		conn.Close()
		return nil, makeError(retval, "C.ct_con_alloc failed")
	}

	// Set password encryption
	cTrue := C.CS_TRUE
	if retval := C.ct_con_props(conn.conn, C.CS_SET, C.CS_SEC_EXTENDED_ENCRYPTION, unsafe.Pointer(&cTrue), C.CS_UNUSED, nil); retval != C.CS_SUCCEED {
		conn.Close()
		return nil, makeError(retval, "C.ct_con_props failed for CS_SEC_EXTENDED_ENCRYPTION")
	}

	cFalse := C.CS_FALSE
	if retval := C.ct_con_props(conn.conn, C.CS_SET, C.CS_SEC_NON_ENCRYPTION_RETRY, unsafe.Pointer(&cFalse), C.CS_UNUSED, nil); retval != C.CS_SUCCEED {
		conn.Close()
		return nil, makeError(retval, "C.ct_con_props failed for CS_SEC_NON_ENCRYPTION_RETRY")
	}

	// Give preference to the user store key
	if len(info.Userstorekey) > 0 {
		// Set userstorekey
		userstorekey := unsafe.Pointer(C.CString(info.Userstorekey))
		defer C.free(userstorekey)
		if retval := C.ct_con_props(conn.conn, C.CS_SET, C.CS_SECSTOREKEY, userstorekey, C.CS_NULLTERM, nil); retval != C.CS_SUCCEED {
			conn.Close()
			return nil, makeError(retval, "C.ct_con_props failed for C.CS_SECSTOREKEY")
		}
	} else {
		// Set username.
		username := unsafe.Pointer(C.CString(info.Username))
		defer C.free(username)
		if retval := C.ct_con_props(conn.conn, C.CS_SET, C.CS_USERNAME, username, C.CS_NULLTERM, nil); retval != C.CS_SUCCEED {
			conn.Close()
			return nil, makeError(retval, "C.ct_con_props failed for CS_USERNAME")
		}

		// Set password.
		password := unsafe.Pointer(C.CString(info.Password))
		defer C.free(password)
		if retval := C.ct_con_props(conn.conn, C.CS_SET, C.CS_PASSWORD, password, C.CS_NULLTERM, nil); retval != C.CS_SUCCEED {
			conn.Close()
			return nil, makeError(retval, "C.ct_con_props failed for CS_PASSWORD")
		}
	}

	if info.Host != "" && info.Port != "" {
		// Set hostname and port as string, since it is modified if
		// '-o ssl' is set.
		strHostport := info.Host + " " + info.Port
		// If '-o ssl='-option is set, add it to strHostport
		if info.TLSHostname != "" {
			strHostport += fmt.Sprintf("ssl=\"%s\"", info.TLSHostname)
		}
		// Create pointer
		ptrHostport := unsafe.Pointer(C.CString(strHostport))
		defer C.free(ptrHostport)

		if retval := C.ct_con_props(conn.conn, C.CS_SET, C.CS_SERVERADDR, ptrHostport, C.CS_NULLTERM, nil); retval != C.CS_SUCCEED {
			conn.Close()
			return nil, makeError(retval, "C.ct_con_props failed for CS_SERVERADDR")
		}
	}

	if info.AppName != "" {
		ptrAppName := unsafe.Pointer(C.CString(info.AppName))
		defer C.free(ptrAppName)

		if retval := C.ct_con_props(conn.conn, C.CS_SET, C.CS_APPNAME, ptrAppName, C.CS_NULLTERM, nil); retval != C.CS_SUCCEED {
			return nil, makeError(retval, "C.ct_con_props failed for CS_APPNAME")
		}
	}

	if retval := C.ct_connect(conn.conn, nil, 0); retval != C.CS_SUCCEED {
		conn.Close()
		return nil, makeError(retval, "C.ct_connect failed")
	}

	// Set database
	if info.Database != "" {
		if _, err := conn.Exec("use "+info.Database, nil); err != nil {
			conn.Close()
			return nil, fmt.Errorf("Failed to connect to database %s: %w", info.Database, err)
		}
	}

	return conn, nil
}

// Close implements the driver.Conn interface. It closes and deallocates
// a connection.
func (conn *Connection) Close() error {
	// Call context.drop when exiting this function to decrease the
	// connection counter and potentially deallocate the context.
	defer conn.driverCtx.dropConn()

	retval := C.ct_close(conn.conn, C.CS_UNUSED)
	if retval != C.CS_SUCCEED {
		return makeError(retval, "C.ct_close failed, connection has results pending")
	}

	retval = C.ct_con_drop(conn.conn)
	if retval != C.CS_SUCCEED {
		return makeError(retval, "C.ct_con_drop failed")
	}

	conn.conn = nil
	return nil
}

// Ping implements the driver.Pinger interface.
func (conn *Connection) Ping(ctx context.Context) error {
	rows, err := conn.QueryContext(ctx, "SELECT 'PING'", nil)
	if err != nil {
		return driver.ErrBadConn
	}
	defer rows.Close()

	cols := rows.Columns()
	cellRefs := make([]driver.Value, len(cols))

	for {
		err := rows.Next(cellRefs)
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("Error occurred while exhausting result set: %w", err)
		}
	}

	return nil
}

// Exec implements the driver.Execer interface.
func (conn *Connection) Exec(query string, args []driver.Value) (driver.Result, error) {
	return conn.ExecContext(context.Background(), query, dblib.ValuesToNamedValues(args))
}

// ExecContext implements the driver.ExecerContext interface.
func (conn *Connection) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	_, result, err := conn.GenericExec(ctx, query, args)
	return result, err
}

// Query implements the driver.Queryer interface.
func (conn *Connection) Query(query string, args []driver.Value) (driver.Rows, error) {
	return conn.QueryContext(context.Background(), query, dblib.ValuesToNamedValues(args))
}

// QueryContext implements the driver.QueryerContext interface.
func (conn *Connection) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	rows, _, err := conn.GenericExec(ctx, query, args)
	return rows, err
}

// CheckNamedValue implements the driver.NamedValueChecker interface.
func (conn *Connection) CheckNamedValue(nv *driver.NamedValue) error {
	v, err := asetypes.DefaultValueConverter.ConvertValue(nv.Value)
	if err != nil {
		return err
	}

	nv.Value = v
	return nil
}
