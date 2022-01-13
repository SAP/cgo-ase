// SPDX-FileCopyrightText: 2020 SAP SE
// SPDX-FileCopyrightText: 2021 SAP SE
// SPDX-FileCopyrightText: 2022 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

package ase

// #include "ctlib.h"
import "C"
import "unsafe"

// Message defines the generic interface Server- und ClientMessage
// adhere to.
type Message interface {
	MessageNumber() uint64
	MessageSeverity() int64
	Content() string
}

// ServerMessage is a message sent from the ASE server to the client.
type ServerMessage struct {
	MsgNumber uint64
	State     int64
	Severity  int64
	Text      string
	Server    string
	Proc      string
	Line      int64
	SQLState  string
}

func newServerMessage(msg *C.CS_SERVERMSG) *ServerMessage {
	return &ServerMessage{
		MsgNumber: uint64(msg.msgnumber),
		State:     int64(msg.state),
		Severity:  int64(msg.severity),
		Text:      C.GoString((*C.char)(unsafe.Pointer(&msg.text))),
		Server:    C.GoString((*C.char)(unsafe.Pointer(&msg.svrname))),
		Proc:      C.GoString((*C.char)(unsafe.Pointer(&msg.proc))),
		Line:      int64(msg.line),
		SQLState:  C.GoString((*C.char)(unsafe.Pointer(&msg.sqlstate))),
	}
}

// MessageNumber returns the message-number of a server-message.
func (msg ServerMessage) MessageNumber() uint64 {
	return msg.MsgNumber
}

// MessageSeverity returns the message-severity of a server-message.
func (msg ServerMessage) MessageSeverity() int64 {
	return msg.Severity
}

// Content returns the message-content of a server-message.
func (msg ServerMessage) Content() string {
	return msg.Text
}

// ClientMessage is a message generated by Client-Library.
type ClientMessage struct {
	Severity  int64
	MsgNumber uint64
	Text      string
	OSNumber  int64
	OSString  string
	Status    int64
	SQLState  string
}

func newClientMessage(msg *C.CS_CLIENTMSG) *ClientMessage {
	return &ClientMessage{
		Severity:  int64(msg.severity),
		MsgNumber: uint64(msg.msgnumber),
		Text:      C.GoString((*C.char)(unsafe.Pointer(&msg.msgstring))),
		OSNumber:  int64(msg.osnumber),
		OSString:  C.GoString((*C.char)(unsafe.Pointer(&msg.osstring))),
		Status:    int64(msg.status),
		SQLState:  C.GoString((*C.char)(unsafe.Pointer(&msg.sqlstate))),
	}
}

// MessageNumber returns the message-number of a client-message.
func (msg ClientMessage) MessageNumber() uint64 {
	return msg.MsgNumber
}

// MessageSeverity returns the message-severity of a client-message.
func (msg ClientMessage) MessageSeverity() int64 {
	return msg.Severity
}

// Content returns the message-content of a client-message.
func (msg ClientMessage) Content() string {
	return msg.Text
}
