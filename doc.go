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
package ase
