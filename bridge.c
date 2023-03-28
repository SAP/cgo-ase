// SPDX-FileCopyrightText: 2020 SAP SE
// SPDX-FileCopyrightText: 2021 SAP SE
// SPDX-FileCopyrightText: 2022 SAP SE
// SPDX-FileCopyrightText: 2023 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

#include <stdio.h>
#include "ctlib.h"
#include "bridge.h"

CS_RETCODE ct_callback_server_message(CS_CONTEXT* ctx, CS_CONNECTION* con, CS_SERVERMSG* msg) {
	return srvMsg(msg);
}

CS_RETCODE ct_callback_client_message(CS_CONTEXT* ctx, CS_CONNECTION* con, CS_CLIENTMSG* msg) {
	return cltMsg(msg);
}
