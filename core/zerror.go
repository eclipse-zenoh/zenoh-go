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

package core

/*
#cgo CFLAGS: -DZENOH_MACOS
#cgo LDFLAGS: -lzenohc

#define ZENOH_MACOS 1

#include <zenoh.h>
*/
import "C"
import "strconv"

func getErrorCodeName(code int) string {
	switch code {
	case C.Z_VLE_PARSE_ERROR:
		return "Z_VLE_PARSE_ERROR"
	case C.Z_ARRAY_PARSE_ERROR:
		return "Z_ARRAY_PARSE_ERROR"
	case C.Z_STRING_PARSE_ERROR:
		return "Z_STRING_PARSE_ERROR"
	case C.ZN_PROPERTY_PARSE_ERROR:
		return "ZN_PROPERTY_PARSE_ERROR"
	case C.ZN_PROPERTIES_PARSE_ERROR:
		return "ZN_PROPERTIES_PARSE_ERROR"
	case C.ZN_MESSAGE_PARSE_ERROR:
		return "ZN_MESSAGE_PARSE_ERROR"
	case C.ZN_INSUFFICIENT_IOBUF_SIZE:
		return "ZN_INSUFFICIENT_IOBUF_SIZE"
	case C.ZN_IO_ERROR:
		return "ZN_IO_ERROR"
	case C.ZN_RESOURCE_DECL_ERROR:
		return "ZN_RESOURCE_DECL_ERROR"
	case C.ZN_PAYLOAD_HEADER_PARSE_ERROR:
		return "ZN_PAYLOAD_HEADER_PARSE_ERROR"
	case C.ZN_TX_CONNECTION_ERROR:
		return "ZN_TX_CONNECTION_ERROR"
	case C.ZN_INVALID_ADDRESS_ERROR:
		return "ZN_INVALID_ADDRESS_ERROR"
	case C.ZN_FAILED_TO_OPEN_SESSION:
		return "ZN_FAILED_TO_OPEN_SESSION"
	case C.ZN_UNEXPECTED_MESSAGE:
		return "ZN_UNEXPECTED_MESSAGE"
	default:
		return "UNKOWN_ERROR_CODE(" + strconv.Itoa(code) + ")"
	}
}

// ZError reports an error that occurred in zenoh.
type ZError struct {
	Msg   string
	Code  int
	Cause error
}

// Error returns the message associated to a ZError
func (e *ZError) Error() string {
	s := e.Msg
	if e.Code != 0 {
		s = s + " (error code: " + getErrorCodeName(e.Code) + ")"
	}
	if e.Cause != nil {
		s = s + ". Caused by: " + e.Cause.Error()
	}
	return s
}
