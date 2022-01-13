// SPDX-FileCopyrightText: 2020 SAP SE
// SPDX-FileCopyrightText: 2021 SAP SE
// SPDX-FileCopyrightText: 2022 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

package ase

/*
#cgo CFLAGS: -I${SRCDIR}/includes
#cgo LDFLAGS: -lsybct64 -lsybct_r64 -lsybcs_r64 -lsybtcl_r64 -lsybcomn_r64 -lsybintl_r64 -lsybunic64
#cgo LDFLAGS: -Wl,-rpath,\$ORIGIN/../lib
#include <stdlib.h>
#include "ctlib.h"
#include "bridge.h"
*/
import "C"
import (
	"database/sql"
	"database/sql/driver"
	"fmt"

	"github.com/SAP/go-dblib/dsn"
)

// Interface satisfaction checks.
var (
	_   driver.Driver = (*aseDrv)(nil)
	drv               = &aseDrv{}
)

//DriverName is the driver name to use with sql.Open for ase databases.
const DriverName = "ase"

// aseDrv is the struct on which we later call Open() to get a connection.
type aseDrv struct{}

func init() {
	sql.Register(DriverName, drv)
}

// Open implements the driver.Driver interface.
func (d *aseDrv) Open(name string) (driver.Conn, error) {
	info := new(Info)
	if err := dsn.Parse(name, info); err != nil {
		return nil, fmt.Errorf("error parsing DSN: %w", err)
	}

	return NewConnection(nil, info)
}

// OpenConnector implements the driver.DriverContext interface.
func (d *aseDrv) OpenConnector(name string) (driver.Connector, error) {
	info := new(Info)
	if err := dsn.Parse(name, info); err != nil {
		return nil, fmt.Errorf("error parsing DSN: %w", err)
	}

	return NewConnector(info)
}
