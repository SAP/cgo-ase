// SPDX-FileCopyrightText: 2020 - 2025 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

package ase

//#include <stdlib.h>
//#include "ctlib.h"
import "C"
import (
	"context"
	"database/sql/driver"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"strings"
	"sync"
	"unsafe"

	"github.com/SAP/go-dblib"
	"github.com/SAP/go-dblib/asetypes"
)

// Interface satisfaction checks.
var (
	_ driver.Stmt              = (*statement)(nil)
	_ driver.StmtExecContext   = (*statement)(nil)
	_ driver.StmtQueryContext  = (*statement)(nil)
	_ driver.NamedValueChecker = (*statement)(nil)
)

// statement implements the driver.Stmt interface.
type statement struct {
	name        string
	argCount    int
	cmd         *Command
	columnTypes []ASEType
}

var (
	statementCounter  uint
	statementCounterM = sync.Mutex{}
)

// Prepare implements the driver.Conn interface.
func (conn *Connection) Prepare(query string) (driver.Stmt, error) {
	return conn.PrepareContext(context.Background(), query)
}

// PrepareContext implements the driver.ConnPrepareContext interface.
func (conn *Connection) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	return conn.prepare(ctx, query)
}

// TODO: Add doc
func (conn *Connection) prepare(ctx context.Context, query string) (*statement, error) {
	stmt := &statement{}

	stmt.argCount = strings.Count(query, "?")

	statementCounterM.Lock()
	statementCounter++
	stmt.name = fmt.Sprintf("stmt%d", statementCounter)
	statementCounterM.Unlock()

	cmd, err := conn.dynamic(stmt.name, query)
	if err != nil {
		stmt.Close()
		return nil, err
	}

	rows, _, err := cmd.ConsumeResponse(ctx)
	if err != nil {
		stmt.Close()
		cmd.Cancel()
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	if rows != nil {
		rows.Close()
		return nil, fmt.Errorf("received rows when creating prepared statement")
	}

	stmt.cmd = cmd

	if err := stmt.fillColumnTypes(); err != nil {
		stmt.Close()
		return nil, fmt.Errorf("Failed to retrieve argument types: %w", err)
	}

	return stmt, nil
}

// Close implements the driver.Stmt interface.
func (stmt *statement) Close() error {
	if stmt.cmd != nil {
		name := C.CString(stmt.name)
		defer C.free(unsafe.Pointer(name))

		retval := C.ct_dynamic(stmt.cmd.cmd, C.CS_DEALLOC, name, C.CS_NULLTERM, nil, C.CS_UNUSED)
		if retval != C.CS_SUCCEED {
			return makeError(retval, "C.ct_dynamic with C.CS_DEALLOC failed")
		}

		retval = C.ct_send(stmt.cmd.cmd)
		if retval != C.CS_SUCCEED {
			return makeError(retval, "C.ct_send failed")
		}

		var err error
		for err = nil; err != io.EOF; _, _, _, err = stmt.cmd.Response() {
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// NumInput implements the driver.Stmt interface.
func (stmt *statement) NumInput() int {
	return stmt.argCount
}

// TODO: Add doc
func (stmt *statement) exec(ctx context.Context, args []driver.NamedValue) (*Rows, *Result, error) {
	if len(args) != stmt.argCount {
		return nil, nil, fmt.Errorf("Mismatched argument count - expected %d, got %d",
			stmt.argCount, len(args))
	}

	name := C.CString(stmt.name)
	defer C.free(unsafe.Pointer(name))

	retval := C.ct_dynamic(stmt.cmd.cmd, C.CS_EXECUTE, name, C.CS_NULLTERM, nil, C.CS_UNUSED)
	if retval != C.CS_SUCCEED {
		return nil, nil, makeError(retval, "C.ct_dynamic with CS_EXECUTE failed")
	}

	for i, arg := range args {
		// TODO place binding in function to achieve an earlier free
		datafmt := (*C.CS_DATAFMT)(C.calloc(1, C.sizeof_CS_DATAFMT))
		defer C.free(unsafe.Pointer(datafmt))
		datafmt.status = C.CS_INPUTVALUE
		datafmt.namelen = C.CS_NULLTERM

		switch stmt.columnTypes[i] {
		case IMAGE:
			datafmt.datatype = (C.CS_INT)(BINARY)
		default:
			datafmt.datatype = (C.CS_INT)(stmt.columnTypes[i])
		}

		// datalen is the length of the data in bytes.
		datalen := 0

		dataType := stmt.columnTypes[i].ToDataType()

		var ptr unsafe.Pointer
		// TODO: This entire case could be moved into a function, since
		// the values set here are always the same - the expected values
		// for ct_param.
		// This function could also check for null values early.
		switch stmt.columnTypes[i] {
		case BIGINT, INT, SMALLINT, TINYINT, UBIGINT, UINT, USMALLINT, USHORT, FLOAT, REAL:
			bs, err := dataType.Bytes(binary.LittleEndian, arg.Value)
			if err != nil {
				// TODO context
				return nil, nil, err
			}
			ptr = C.CBytes(bs)
			defer C.free(ptr)
		case DECIMAL, NUMERIC:
			bs, err := dataType.Bytes(binary.LittleEndian, arg.Value)
			if err != nil {
				return nil, nil, err
			}

			csDec := (*C.CS_DECIMAL)(C.calloc(1, C.sizeof_CS_DECIMAL))
			defer C.free(unsafe.Pointer(csDec))
			csDec.precision = (C.CS_BYTE)(arg.Value.(*asetypes.Decimal).Precision)
			csDec.scale = (C.CS_BYTE)(arg.Value.(*asetypes.Decimal).Scale)

			for i, b := range bs {
				csDec.array[i] = (C.CS_BYTE)(b)
			}

			ptr = unsafe.Pointer(csDec)
		case MONEY, MONEY4, DATE, TIME, DATETIME4, DATETIME, BIGDATETIME, BIGTIME:
			bs, err := dataType.Bytes(binary.LittleEndian, arg.Value)
			if err != nil {
				return nil, nil, err
			}

			ptr = C.CBytes(bs)
			defer C.free(ptr)
		case CHAR:
			ptr = unsafe.Pointer(C.CString(arg.Value.(string)))
			defer C.free(ptr)

			datalen = len(arg.Value.(string))
			datafmt.format = C.CS_FMT_NULLTERM
			datafmt.maxlength = C.CS_MAX_CHAR
		case TEXT, LONGCHAR:
			ptr = unsafe.Pointer(C.CString(arg.Value.(string)))
			defer C.free(ptr)

			datalen = len(arg.Value.(string))
			datafmt.format = C.CS_FMT_NULLTERM
			datafmt.maxlength = (C.CS_INT)(datalen)
		case VARCHAR:
			varchar := (*C.CS_VARCHAR)(C.calloc(1, C.sizeof_CS_VARCHAR))
			defer C.free(unsafe.Pointer(varchar))
			varchar.len = (C.CS_SMALLINT)(len(arg.Value.(string)))

			for i, chr := range arg.Value.(string) {
				varchar.str[i] = (C.CS_CHAR)(chr)
			}

			ptr = unsafe.Pointer(varchar)
		case BINARY, IMAGE:
			ptr = C.CBytes(arg.Value.([]byte))
			defer C.free(ptr)
			datalen = len(arg.Value.([]byte))

			// IMAGE does not support null padding
			if stmt.columnTypes[i] == BINARY {
				datafmt.format = C.CS_FMT_PADNULL
			}

			// The maximum length of slices is constrained by the
			// ability to address elements by integers - hence the
			// maximum length we can retrieve is MaxInt64.
			datafmt.maxlength = (C.CS_INT)(math.MaxInt32)
		case VARBINARY:
			varbin := (*C.CS_VARBINARY)(C.calloc(1, C.sizeof_CS_VARBINARY))
			defer C.free(unsafe.Pointer(varbin))
			varbin.len = (C.CS_SMALLINT)(len(arg.Value.([]byte)))

			for i, b := range arg.Value.([]byte) {
				varbin.array[i] = (C.CS_BYTE)(b)
			}

			ptr = unsafe.Pointer(varbin)
		case BIT:
			b := (C.CS_BOOL)(0)
			if arg.Value.(bool) {
				b = (C.CS_BOOL)(1)
			}
			ptr = unsafe.Pointer(&b)
			datalen = 1
		case UNICHAR, UNITEXT:
			bs, err := dataType.Bytes(binary.LittleEndian, arg.Value)
			if err != nil {
				return nil, nil, err
			}

			ptr = unsafe.Pointer(C.CBytes(bs))
			defer C.free(ptr)

			datalen = len(bs)
			datafmt.format = C.CS_FMT_NULLTERM
			datafmt.maxlength = (C.CS_INT)(datalen)
		default:
			return nil, nil, fmt.Errorf("Unhandled column type: %s", stmt.columnTypes[i])
		}

		var csDatalen C.CS_INT
		if datalen != C.CS_UNUSED {
			csDatalen = (C.CS_INT)(datalen)
		} else {
			csDatalen = C.CS_UNUSED
		}

		retval = C.ct_param(stmt.cmd.cmd, datafmt, ptr, csDatalen, 0)
		if retval != C.CS_SUCCEED {
			return nil, nil, makeError(retval, "C.ct_param on parameter %d failed with argument '%v'", i, arg)
		}
	}

	retval = C.ct_send(stmt.cmd.cmd)
	if retval != C.CS_SUCCEED {
		return nil, nil, makeError(retval, "C.ct_send failed")
	}

	return stmt.cmd.ConsumeResponse(ctx)
}

// Exec implements the driver.Stmt interface.
func (stmt *statement) Exec(args []driver.Value) (driver.Result, error) {
	return stmt.ExecContext(context.Background(), dblib.ValuesToNamedValues(args))
}

// ExecContext implements the driver.StmtExecContext interface.
func (stmt *statement) ExecContext(ctx context.Context, args []driver.NamedValue) (driver.Result, error) {
	_, result, err := stmt.exec(ctx, args)
	return result, err
}

// Query implements the driver.Stmt interface.
func (stmt *statement) Query(args []driver.Value) (driver.Rows, error) {
	return stmt.QueryContext(context.Background(), dblib.ValuesToNamedValues(args))
}

// QueryContext implements the driver.StmtQueryContext interface.
func (stmt *statement) QueryContext(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
	rows, _, err := stmt.exec(ctx, args)
	return rows, err
}

func (stmt *statement) fillColumnTypes() error {
	name := C.CString(stmt.name)
	defer C.free(unsafe.Pointer(name))

	// Instruct server to send data to descriptor
	retval := C.ct_dynamic(stmt.cmd.cmd, C.CS_DESCRIBE_INPUT, name,
		C.CS_NULLTERM, nil, C.CS_UNUSED)
	if retval != C.CS_SUCCEED {
		return makeError(retval, "Error when preparing input description")
	}

	retval = C.ct_send(stmt.cmd.cmd)
	if retval != C.CS_SUCCEED {
		return makeError(retval, "Error sending command to server")
	}

	for {
		_, _, resultType, err := stmt.cmd.Response()
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("Received error while receiving input description: %w", err)
		}

		if resultType != C.CS_DESCRIBE_RESULT {
			continue
		}

		// Receive number of arguments
		var paramCount C.CS_INT
		retval = C.ct_res_info(stmt.cmd.cmd, C.CS_NUMDATA, unsafe.Pointer(&paramCount), C.CS_UNUSED, nil)
		if retval != C.CS_SUCCEED {
			return makeError(retval, "Failed to retrieve parameter count")
		}

		stmt.argCount = int(paramCount)
		stmt.columnTypes = make([]ASEType, stmt.argCount)

		for i := 0; i < stmt.argCount; i++ {
			datafmt := (*C.CS_DATAFMT)(C.calloc(1, C.sizeof_CS_DATAFMT))

			retval = C.ct_describe(stmt.cmd.cmd, (C.CS_INT)(i+1), datafmt)
			if retval != C.CS_SUCCEED {
				return makeError(retval, "Failed to retrieve description of parameter %d", i)
			}

			stmt.columnTypes[i] = ASEType(datafmt.datatype)
		}

	}

	return nil
}

// CheckNamedValue implements the driver.NamedValueChecker interface.
func (stmt statement) CheckNamedValue(named *driver.NamedValue) error {
	index := named.Ordinal - 1
	if index > len(stmt.columnTypes) {
		return fmt.Errorf("cgo-ase: ordinal %d is larger than the number of columns %d",
			named.Ordinal, len(stmt.columnTypes))
	}

	val, err := stmt.columnTypes[index].ToDataType().ConvertValue(named.Value)
	if err != nil {
		return fmt.Errorf("cgo-ase: error converting value: %w", err)
	}

	named.Value = val
	return nil
}
