// SPDX-FileCopyrightText: 2020 SAP SE
// SPDX-FileCopyrightText: 2021 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

// This example shows how a custom recorder can be implemented to
// process messages from the TDS server.
package main

import (
	"database/sql"
	"fmt"
	"log"
	"sync"

	"github.com/SAP/cgo-ase"
	"github.com/SAP/go-dblib/dsn"
)

func main() {
	if err := DoMain(); err != nil {
		log.Fatalf("recorderexample: %v", err)
	}
}

func DoMain() error {
	recorder := &Recorder{}
	ase.GlobalServerMessageBroker.RegisterHandler(recorder.HandleMessage)
	ase.GlobalClientMessageBroker.RegisterHandler(recorder.HandleMessage)

	info, err := ase.NewInfoWithEnv()
	if err != nil {
		return fmt.Errorf("error reading DSN info from env: %w", err)
	}

	db, err := sql.Open("ase", dsn.FormatSimple(info))
	if err != nil {
		return fmt.Errorf("error opening database: %w", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("error closing db: %v", err)
		}
	}()

	rows, err := db.Query("sp_adduser nologin")
	if err != nil {
		fmt.Println("Messages from ASE server:")
		for _, msg := range recorder.Messages {
			fmt.Printf("    %d: %s\n", msg.MessageNumber(), msg.Content())
		}
		return err
	}

	var returnStatus int
	for rows.Next() {
		if err := rows.Scan(&returnStatus); err != nil {
			return fmt.Errorf("error scanning return status: %w", err)
		}

		if returnStatus != 0 {
			fmt.Println("Messages from ASE server:")
			for _, msg := range recorder.Messages {
				fmt.Printf("    %d: %s\n", msg.MessageNumber(), msg.Content())
			}
			return fmt.Errorf("sp_adduser failed with return status %d", returnStatus)
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("rows reported error: %w", err)
	}

	return nil
}

type Recorder struct {
	sync.RWMutex
	Messages []ase.Message
}

func (rec *Recorder) HandleMessage(msg ase.Message) {
	rec.Lock()
	defer rec.Unlock()

	rec.Messages = append(rec.Messages, msg)
}

func (rec *Recorder) Reset() {
	rec.Lock()
	defer rec.Unlock()

	rec.Messages = []ase.Message{}
}
