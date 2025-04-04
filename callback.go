// SPDX-FileCopyrightText: 2020 - 2025 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

// +build integration

package ase

import (
	"log"
)

func genMessageHandler() MessageHandler {
	return func(msg Message) {
		// Ignore CS_SV_INFORM
		if msg.MessageSeverity() == 10 {
			return
		}

		log.Printf("Callback: %v", msg.Content())
	}
}
