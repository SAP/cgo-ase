// SPDX-FileCopyrightText: 2020 SAP SE
// SPDX-FileCopyrightText: 2021 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/SAP/cgo-ase"
	"github.com/SAP/go-dblib/dsn"
	"github.com/SAP/go-dblib/term"

	"github.com/spf13/pflag"
)

func main() {
	if err := doMain(); err != nil {
		log.Fatalf("cgoase failed: %v", err)
	}
}

func doMain() error {
	ase.GlobalServerMessageBroker.RegisterHandler(handleMessage)
	ase.GlobalClientMessageBroker.RegisterHandler(handleMessage)

	info, flagset, err := ase.NewInfoWithFlags()
	if err != nil {
		return fmt.Errorf("error creating info: %w", err)
	}

	// Use pflag to merge flagsets
	flags := pflag.NewFlagSet("cgoase", pflag.ContinueOnError)

	// Merge info flagset
	flags.AddGoFlagSet(flagset)

	// Merge stdlib flag arguments
	flags.AddGoFlagSet(flag.CommandLine)

	if err := flags.Parse(os.Args[1:]); err != nil {
		return err
	}

	if err := dsn.FromEnv("ASE", info); err != nil {
		return fmt.Errorf("error reading values from environment: %w", err)
	}

	db, err := sql.Open("ase", dsn.FormatSimple(info))
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	return term.Entrypoint(db, flags.Args())
}

func handleMessage(msg ase.Message) {
	if msg.MessageSeverity() == 10 {
		return
	}

	log.Printf("Msg %d: %s", msg.MessageNumber(), msg.Content())
}
