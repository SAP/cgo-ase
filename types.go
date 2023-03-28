// SPDX-FileCopyrightText: 2020 SAP SE
// SPDX-FileCopyrightText: 2021 SAP SE
// SPDX-FileCopyrightText: 2022 SAP SE
// SPDX-FileCopyrightText: 2023 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

package ase

import (
	"github.com/SAP/go-dblib/asetypes"
)

// Generate ASETypes and type2string-function.
//go:generate go run ./gen_types.go

// ASEType is the byte-representation of an ASE-datatype.
type ASEType byte

// String returns the ASEType as string and satisfies the
// stringer interface.
func (t ASEType) String() string {
	s, ok := type2string[t]
	if !ok {
		return ""
	}
	return s
}

// ToDataType returns the equivalent asetypes.DataType for an ASEType.
func (t ASEType) ToDataType() asetypes.DataType {
	switch t {
	case BIGDATETIME:
		return asetypes.BIGDATETIMEN
	case BIGINT:
		return asetypes.INT8
	case BIGTIME:
		return asetypes.BIGTIMEN
	case BINARY:
		return asetypes.BINARY
	case BIT:
		return asetypes.BIT
	case BLOB:
		return asetypes.BLOB
	case BOUNDARY:
		return asetypes.BOUNDARY
	case CHAR:
		return asetypes.CHAR
	case DATE:
		return asetypes.DATE
	case DATETIME:
		return asetypes.DATETIME
	case DATETIME4:
		return asetypes.SHORTDATE
	case DECIMAL:
		return asetypes.DECN
	case FLOAT:
		return asetypes.FLT8
	case IMAGE:
		return asetypes.IMAGE
	case IMAGELOCATOR:
		// TODO
		return 0
	case INT:
		return asetypes.INT4
	case LONG:
		return asetypes.INT8
	case LONGBINARY:
		return asetypes.LONGBINARY
	case LONGCHAR:
		return asetypes.LONGCHAR
	case MONEY:
		return asetypes.MONEY
	case MONEY4:
		return asetypes.SHORTMONEY
	case NUMERIC:
		return asetypes.NUMN
	case REAL:
		return asetypes.FLT4
	case SENSITIVITY:
		return asetypes.SENSITIVITY
	case SMALLINT:
		return asetypes.INT2
	case TEXT:
		return asetypes.TEXT
	case TEXTLOCATOR:
		// TODO
		return 0
	case TIME:
		return asetypes.TIME
	case TINYINT:
		return asetypes.INT1
	case UBIGINT:
		return asetypes.UINT8
	case UINT:
		return asetypes.UINT4
	case UNICHAR, UNITEXT:
		return asetypes.UNITEXT
	case UNITEXTLOCATOR:
		// TODO
		return 0
	case USER:
		// TODO
		return 0
	case USHORT, USMALLINT:
		return asetypes.UINT2
	case VARBINARY:
		return asetypes.VARBINARY
	case VARCHAR:
		return asetypes.VARCHAR
	case VOID:
		return asetypes.VOID
	case XML:
		return asetypes.XML
	default:
		return 0
	}
}
