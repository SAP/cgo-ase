// SPDX-FileCopyrightText: 2020 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

// Package ase contains code of the cgo-ase driver for the database/sql
// package of Go (golang) to provide access to SAP ASE instances.
//
// By using cgo this code is also enabled to call C code and to link against
// shared object. Thus, it is a shim to satisfy the
// databse/sql/driver-interface.
//
// TODO: Some sentences about the .c and .h-files in this package
//
// Besides the cgo-implementations the package also implements the
// following database/sql/driver-interfaces:
//
// - driver.Conn
//
// - driver.ConnPrepareContext
//
// - driver.ConnBeginTx
//
// - driver.Connector
//
// - driver.Driver
//
// - driver.DriverContext
//
// - driver.Execer
//
// - driver.ExecerContext
//
// - driver.NamedValueChecker
//
// - driver.Pinger
//
// - driver.Queryer
//
// - driver.QueryerContext
//
// - driver.Result
//
// - driver.Rows
//
// - driver.RowsColumnTypeLength
//
// - driver.RowsColumnTypeNullable
//
// - driver.Stmt
//
// - driver.StmtExecContext
//
// - driver.StmtQueryContext
//
// - driver.Tx
//
package ase
