// SPDX-FileCopyrightText: 2020 SAP SE
// SPDX-FileCopyrightText: 2021 SAP SE
// SPDX-FileCopyrightText: 2022 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

package ase

import (
	"context"
	"database/sql/driver"
	"fmt"
)

// Interface satisfaction checks.
var _ driver.Connector = (*connector)(nil)

// connector implements the driver.Connector interface.
type connector struct {
	driverCtx *csContext
	info      *Info
}

// NewConnector returns a new connector with the passed configuration.
func NewConnector(info *Info) (driver.Connector, error) {
	driverCtx, err := newCsContext(info)
	if err != nil {
		return nil, fmt.Errorf("Failed to initialize context: %w", err)
	}

	c := &connector{
		driverCtx: driverCtx,
		info:      info,
	}

	conn, err := c.Connect(context.Background())
	if err != nil {
		driverCtx.drop()
		return nil, fmt.Errorf("Failed to open connection: %w", err)
	}

	defer func() {
		// In- and decrease connections count before and after closing
		// connection to prevent the context being deallocated.
		driverCtx.connections++
		conn.Close()
		driverCtx.connections--
	}()

	return c, nil
}

// Connect implements the driver.Connector interface.
func (connector *connector) Connect(ctx context.Context) (driver.Conn, error) {
	connChan := make(chan driver.Conn, 1)
	errChan := make(chan error, 1)
	go func() {
		conn, err := NewConnection(connector.driverCtx, connector.info)
		connChan <- conn
		close(connChan)
		errChan <- err
		close(errChan)
	}()

	select {
	case <-ctx.Done():
		defer func() {
			conn := <-connChan
			if conn != nil {
				conn.Close()
			}
		}()
		return nil, ctx.Err()
	case err := <-errChan:
		conn := <-connChan
		return conn, err
	}
}

// Driver implements the driver.Connector interface.
func (connector connector) Driver() driver.Driver {
	return drv
}
