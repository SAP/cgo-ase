// SPDX-FileCopyrightText: 2020 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/SAP/cgo-ase"
	"github.com/SAP/go-dblib/term"
)

func main() {
	if err := doMain(); err != nil {
		log.Fatalf("cgoase failed: %v", err)
	}
}

func doMain() error {
	ase.GlobalServerMessageBroker.RegisterHandler(handleMessage)
	ase.GlobalClientMessageBroker.RegisterHandler(handleMessage)

	dsn, err := term.Dsn()
	if err != nil {
		return fmt.Errorf("error parsing DSN: %w", err)
	}

	db, err := sql.Open("ase", dsn.AsSimple())
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	return term.Entrypoint(db)
}

func handleMessage(msg ase.Message) {
	if msg.MessageSeverity() == 10 {
		return
	}

	log.Printf("Msg %d: %s", msg.MessageNumber(), msg.Content())
}
