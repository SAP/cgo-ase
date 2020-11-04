// SPDX-FileCopyrightText: 2020 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

package ase

// #include "ctlib.h"
import "C"
import (
	"fmt"
	"os"
	"sync"
)

var (
	cbTarget = os.Stderr
	cbRW     = &sync.RWMutex{}
)

// SetCallbackTarget sets the os.File target the callback functions for
// client and server messages will write to.
func SetCallbackTarget(target *os.File) {
	cbRW.Lock()
	defer cbRW.Unlock()

	cbTarget = target
}

// srvMsg is a callback function which will be called from C when the
// server sends a message. The message is then passed to the
// GlobalServerMessageBroker.
// Don't change the following line. It is the directive for cgo to make
// the function available from C.
//export srvMsg
func srvMsg(msg *C.CS_SERVERMSG) C.CS_RETCODE {
	GlobalServerMessageBroker.recvServerMessage(msg)
	return C.CS_SUCCEED
}

func logSrvMsg(msg Message) {
	cbRW.RLock()
	defer cbRW.RUnlock()

	srvMsg, ok := msg.(ServerMessage)
	if !ok {
		return
	}

	fmt.Fprintln(cbTarget, "Server message:")
	fmt.Fprintf(cbTarget, "\tmsgnumber:   %d\n", srvMsg.MsgNumber)
	fmt.Fprintf(cbTarget, "\tstate:       %d\n", srvMsg.State)
	fmt.Fprintf(cbTarget, "\tseverity:    %d\n", srvMsg.Severity)
	fmt.Fprintf(cbTarget, "\ttext:        %s\n", srvMsg.Text)
	fmt.Fprintf(cbTarget, "\tserver:      %s\n", srvMsg.Server)
	fmt.Fprintf(cbTarget, "\tproc:        %s\n", srvMsg.Proc)
	fmt.Fprintf(cbTarget, "\tline:        %d\n", srvMsg.Line)
	fmt.Fprintf(cbTarget, "\tsqlstate:    %s\n", srvMsg.SQLState)
}

// cltMsg is a callback function which will be called from C when the
// client sends a message. The message is then passed to the
// GlobalClientMessageBroker.
// Don't change the following line. It is the directive for cgo to make
// the function available from C.
//export cltMsg
func cltMsg(msg *C.CS_CLIENTMSG) C.CS_RETCODE {
	GlobalClientMessageBroker.recvClientMessage(msg)
	return C.CS_SUCCEED
}

func logCltMsg(msg Message) {
	cbRW.RLock()
	defer cbRW.RUnlock()

	cltMsg, ok := msg.(ClientMessage)
	if !ok {
		return
	}

	fmt.Fprintln(cbTarget, "Client message:")
	fmt.Fprintf(cbTarget, "\tseverity:     %d\n", cltMsg.Severity)
	fmt.Fprintf(cbTarget, "\tmsgnumber:    %d\n", cltMsg.MsgNumber)
	fmt.Fprintf(cbTarget, "\tmsgstring:    %s\n", cltMsg.Text)
	fmt.Fprintf(cbTarget, "\tosnumber:     %d\n", cltMsg.OSNumber)
	fmt.Fprintf(cbTarget, "\tosstring:     %s\n", cltMsg.OSString)
	fmt.Fprintf(cbTarget, "\tstatus:       %d\n", cltMsg.Status)
	fmt.Fprintf(cbTarget, "\tsqlstate:     %s\n", cltMsg.SQLState)
}
