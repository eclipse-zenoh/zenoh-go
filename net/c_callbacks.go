/*
 * Copyright (c) 2017, 2020 ADLINK Technology Inc.
 *
 * This program and the accompanying materials are made available under the
 * terms of the Eclipse Public License 2.0 which is available at
 * http://www.eclipse.org/legal/epl-2.0, or the Apache License, Version 2.0
 * which is available at https://www.apache.org/licenses/LICENSE-2.0.
 *
 * SPDX-License-Identifier: EPL-2.0 OR Apache-2.0
 *
 * Contributors:
 *   ADLINK zenoh team, <zenoh@adlink-labs.tech>
 */

package net

/*
#include <zenoh.h>
#include <stdio.h>
#include <zenoh/codec.h>

void subscriber_handle_data_cgo(const zn_resource_key_t *rkey, const unsigned char *data, size_t length, const zn_data_info_t *info, void *arg) {
	void callSubscriberDataHandler(const zn_resource_key_t*, const unsigned char*, size_t, const zn_data_info_t*, void*);
	callSubscriberDataHandler(rkey, data, length, info, arg);
}

void storage_handle_data_cgo(const zn_resource_key_t *rkey, const unsigned char *data, size_t length, const zn_data_info_t *info, void *arg) {
	void callStorageDataHandler(const zn_resource_key_t*, const unsigned char*, size_t, const zn_data_info_t*, void*);
	callStorageDataHandler(rkey, data, length, info, arg);
}

void storage_handle_query_cgo(const char *rname, const char *predicate, zn_replies_sender_t send_replies, void *query_handle, void *arg) {
	void callStorageQueryHandler(const char*, const char*, zn_replies_sender_t, void*, void*);
	callStorageQueryHandler(rname, predicate, send_replies, query_handle, arg);
}

void eval_handle_query_cgo(const char *rname, const char *predicate, zn_replies_sender_t send_replies, void *query_handle, void *arg) {
	void callEvalQueryHandler(const char*, const char*, zn_replies_sender_t, void*, void*);
	callEvalQueryHandler(rname, predicate, send_replies, query_handle, arg);
}

void handle_reply_cgo(const zn_reply_value_t *reply, void *arg) {
	void callReplyHandler(const zn_reply_value_t*, void*);
	callReplyHandler(reply, arg);
}

*/
import "C"
