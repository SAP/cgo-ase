// SPDX-FileCopyrightText: 2020 - 2025 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

package ase

//#include "ctlib.h"
//#include "bridge.h"
import "C"
import (
	"fmt"
	"sync"
)

// context wraps C.CS_CONTEXT to ensure that the context is being closed
// and deallocated after the last connection was closed.
type csContext struct {
	ctx  *C.CS_CONTEXT
	info *Info

	// connections is a counter that keeps track of the number of
	// connections using the context to communicate with an ASE
	// instance.
	// lock is used to guard access to connections.
	connections uint
	lock        sync.Mutex
}

func newCsContext(info *Info) (*csContext, error) {
	ctx := &csContext{}
	ctx.info = info

	if err := ctx.init(); err != nil {
		return nil, err
	}

	return ctx, nil
}

// newConn is called by new connections to register their creation with
// the context.
// If the context is not initialized it will be initialized.
func (context *csContext) newConn() error {
	context.lock.Lock()
	defer context.lock.Unlock()

	if context.ctx == nil {
		if err := context.init(); err != nil {
			return err
		}
	}

	context.connections++

	return nil
}

// dropConn is called by connections to register their deallocation with
// the context.
// If no other connection uses the context it will be deallocated.
func (context *csContext) dropConn() error {
	context.lock.Lock()
	defer context.lock.Unlock()

	context.connections--
	if context.connections > 0 {
		return nil
	}

	// No connections using the context left, proceed with deallocation.
	return context.drop()
}

// init allocates and initializes the context.
// If a DSN is set the DSN options will be applied.
func (context *csContext) init() error {
	if retval := C.cs_ctx_alloc(C.CS_CURRENT_VERSION, &context.ctx); retval != C.CS_SUCCEED {
		return makeError(retval, "C.cs_ctx_alloc failed")
	}

	if retval := C.ct_init(context.ctx, C.CS_CURRENT_VERSION); retval != C.CS_SUCCEED {
		if err := context.drop(); err != nil {
			return err
		}
		return makeError(retval, "C.ct_init failed")
	}

	if err := context.applyDSN(context.info); err != nil {
		return fmt.Errorf("Failed to apply DSN: %w", err)
	}

	return nil
}

// drop deallocates the context.
func (context *csContext) drop() error {
	if retval := C.ct_exit(context.ctx, C.CS_UNUSED); retval != C.CS_SUCCEED {
		return makeError(retval, "C.ct_exit failed, has results pending")
	}

	if retval := C.cs_ctx_drop(context.ctx); retval != C.CS_SUCCEED {
		return makeError(retval, "C.cs_ctx_drop failed")
	}

	context.ctx = nil
	return nil
}

// applyDSN applies the relevant connection properties of a DSN to the
// context.
func (context *csContext) applyDSN(info *Info) error {
	if retval := C.ct_callback(context.ctx, nil, C.CS_SET, C.CS_CLIENTMSG_CB, C.ct_callback_client_message); retval != C.CS_SUCCEED {
		return makeError(retval, "C.ct_callback failed for client messages")
	}

	if retval := C.ct_callback(context.ctx, nil, C.CS_SET, C.CS_SERVERMSG_CB, C.ct_callback_server_message); retval != C.CS_SUCCEED {
		return makeError(retval, "C.ct_callback failed for server messages")
	}

	if info.LogClientMsgs {
		GlobalClientMessageBroker.RegisterHandler(logCltMsg)
	}

	if info.LogServerMsgs {
		GlobalServerMessageBroker.RegisterHandler(logSrvMsg)
	}

	return nil
}
