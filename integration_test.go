// SPDX-FileCopyrightText: 2020 - 2025 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

// +build integration

package ase

import (
	"database/sql/driver"
	"fmt"
	"log"
	"testing"

	"github.com/SAP/go-dblib/integration"
)

// TestMain starts the integration tests by using the imported package
// 'go-dblib/integration'.
func TestMain(m *testing.M) {
	if err := testMain(m); err != nil {
		log.Fatal(err)
	}
}

func testMain(m *testing.M) error {
	GlobalServerMessageBroker.RegisterHandler(genMessageHandler())
	GlobalClientMessageBroker.RegisterHandler(genMessageHandler())

	// Setup test for username/password
	teardown, err := setup("username password", func(info *Info) {
		info.Userstorekey = ""
	})
	if err != nil {
		return err
	}
	defer teardown()

	// Setup test with userkeystore
	teardown, err = setup("userstorekey", func(info *Info) {
		info.Username = ""
		info.Password = ""
	})
	if err != nil {
		return err
	}
	defer teardown()

	if rc := m.Run(); rc != 0 {
		return fmt.Errorf("tests failed with %d", rc)
	}

	return nil
}

// newConnector statisfies the integration.NewConnectorFn interface.
func newConnector(info interface{}) (driver.Connector, error) {
	return NewConnector(info.(*Info))
}

func setup(name string, infoMod func(*Info)) (func(), error) {
	info, err := NewInfoWithEnv()
	if err != nil {
		return nil, err
	}

	infoMod(info)

	if err := integration.SetupDB(info); err != nil {
		return nil, err
	}

	deferFn := func() {
		if err := integration.TeardownDB(info); err != nil {
			log.Printf("error dropping database %q: %v", info.Database, err)
		}
	}

	if err := integration.RegisterDSN(name, info, newConnector); err != nil {
		return nil, fmt.Errorf("error setting up userstore database: %w", err)
	}

	return deferFn, nil
}

// Exact numeric integer
func TestBigInt(t *testing.T)           { integration.DoTestBigInt(t) }
func TestInt(t *testing.T)              { integration.DoTestInt(t) }
func TestSmallInt(t *testing.T)         { integration.DoTestSmallInt(t) }
func TestTinyInt(t *testing.T)          { integration.DoTestTinyInt(t) }
func TestUnsignedBigInt(t *testing.T)   { integration.DoTestUnsignedBigInt(t) }
func TestUnsignedInt(t *testing.T)      { integration.DoTestUnsignedInt(t) }
func TestUnsignedSmallInt(t *testing.T) { integration.DoTestUnsignedSmallInt(t) }

// Exact numeric decimal
func TestDecimal(t *testing.T)     { integration.DoTestDecimal(t) }
func TestDecimal10(t *testing.T)   { integration.DoTestDecimal10(t) }
func TestDecimal380(t *testing.T)  { integration.DoTestDecimal380(t) }
func TestDecimal3838(t *testing.T) { integration.DoTestDecimal3838(t) }

// Approximate numeric
func TestFloat(t *testing.T) { integration.DoTestFloat(t) }
func TestReal(t *testing.T)  { integration.DoTestReal(t) }

// Money
func TestMoney(t *testing.T)  { integration.DoTestMoney(t) }
func TestMoney4(t *testing.T) { integration.DoTestMoney4(t) }

// Date and time
func TestDate(t *testing.T)          { integration.DoTestDate(t) }
func TestTime(t *testing.T)          { integration.DoTestTime(t) }
func TestSmallDateTime(t *testing.T) { integration.DoTestSmallDateTime(t) }
func TestDateTime(t *testing.T)      { integration.DoTestDateTime(t) }
func TestBigDateTime(t *testing.T)   { integration.DoTestBigDateTime(t) }
func TestBigTime(t *testing.T)       { integration.DoTestBigTime(t) }

// Character
func TestVarChar(t *testing.T)  { integration.DoTestVarChar(t) }
func TestChar(t *testing.T)     { integration.DoTestChar(t) }
func TestNChar(t *testing.T)    { integration.DoTestNChar(t) }
func TestNVarChar(t *testing.T) { integration.DoTestNVarChar(t) }
func TestText(t *testing.T)     { integration.DoTestText(t) }
func TestUniChar(t *testing.T)  { integration.DoTestUniChar(t) }
func TestUniText(t *testing.T)  { integration.DoTestUniText(t) }

// Binary
func TestBinary(t *testing.T)    { integration.DoTestBinary(t) }
func TestVarBinary(t *testing.T) { integration.DoTestVarBinary(t) }

// Bit
func TestBit(t *testing.T) { integration.DoTestBit(t) }

// Image
func TestImage(t *testing.T) { integration.DoTestImage(t) }

// Routines
func TestSQLTx(t *testing.T)       { integration.DoTestSQLTx(t) }
func TestSQLExec(t *testing.T)     { integration.DoTestSQLExec(t) }
func TestSQLQueryRow(t *testing.T) { integration.DoTestSQLQueryRow(t) }
