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

// Package net provides the Zenoh-net API in Go.
package net

/*
#cgo CFLAGS: -DZENOH_MACOS
#cgo LDFLAGS: -lzenohc

#define ZENOH_MACOS 1

#include <stdlib.h>
#include <stdio.h>
#include <zenoh/net/session.h>
#include <zenoh/net/recv_loop.h>
#include <zenoh/rname.h>

// Forward declarations of callbacks (see c_callbacks.go)
extern void subscriber_handle_data_cgo(const zn_resource_key_t *rkey, const unsigned char *data, size_t length, const zn_data_info_t *info, void *arg);
extern void storage_handle_data_cgo(const zn_resource_key_t *rkey, const unsigned char *data, size_t length, const zn_data_info_t *info, void *arg);
extern void storage_handle_query_cgo(const char *rname, const char *predicate, zn_replies_sender_t send_replies, void *query_handle, void *arg);
extern void eval_handle_query_cgo(const char *rname, const char *predicate, zn_replies_sender_t send_replies, void *query_handle, void *arg);
extern void handle_reply_cgo(const zn_reply_value_t *reply, void *arg);
*/
import "C"
import (
	"bytes"
	"encoding/binary"
	"fmt"
	"runtime/debug"
	"strconv"
	"sync"
	"unsafe"

	log "github.com/sirupsen/logrus"
)

var logger = log.WithFields(log.Fields{" pkg": "zenoh/net"})

// Open opens a zenoh-net session.
// 'locator' is a pointer to a string representing the network endpoint to which establish the session. A typical locator looks like this : "tcp/127.0.0.1:7447".
// 	   If 'locator' is "nil", 'open' will scout and try to establish the session automatically.
// 'properties' is a map of properties that will be used to establish and configure the zenoh session.
// 	   'properties' will typically contain the username and password informations needed to establish the zenoh session with a secured infrastructure.
// 	   It can be set to "nil".
// Return a handle to the zenoh session.
func Open(locator *string, properties map[int][]byte) (*Session, error) {
	logger.WithField("locator", locator).Debug("Open")

	pvec := ((C.z_vec_t)(C.z_vec_make(C.uint(len(properties)))))
	for k, v := range properties {
		value := C.z_uint8_array_t{length: C.uint(len(v)), elem: (*C.uchar)(unsafe.Pointer(&v[0]))}
		prop := ((*C.zn_property_t)(C.zn_property_make(C.ulong(k), value)))
		C.z_vec_append(&pvec, unsafe.Pointer(prop))
	}

	var l *C.char
	if locator != nil {
		l = C.CString(*locator)
		defer C.free(unsafe.Pointer(l))
	}

	result := C.zn_open(l, nil, &pvec)
	if result.tag == C.Z_ERROR_TAG {
		return nil, &ZError{"zn_open failed", resultValueToErrorCode(result.value), nil}
	}
	s := resultValueToSession(result.value)

	logger.WithField("locator", locator).Debug("Run zn_recv_loop")
	go C.zn_recv_loop(s)

	return s, nil
}

// Close the zenoh-net session 'z'.
func (s *Session) Close() error {
	logger.Debug("Close")
	errcode := C.zn_stop_recv_loop(s)
	if errcode != 0 {
		return &ZError{"zn_stop_recv_loop failed", int(errcode), nil}
	}
	errcode = C.zn_close(s)
	if errcode != 0 {
		return &ZError{"zn_close failed", int(errcode), nil}
	}
	return nil
}

// Info returns a map of properties containing various informations about the
// established zenoh-net session.
func (s *Session) Info() map[int][]byte {
	info := map[int][]byte{}
	cprops := C.zn_info(s)
	propslength := int(C.z_vec_length(&cprops))
	for i := 0; i < propslength; i++ {
		cvalue := ((*C.zn_property_t)(C.z_vec_get(&cprops, C.uint(i)))).value
		info[i] = C.GoBytes(unsafe.Pointer(cvalue.elem), C.int(cvalue.length))
	}
	return info
}

type subscriberHandlersRegistry struct {
	mu       *sync.Mutex
	index    int
	dHandler map[int]DataHandler
}

var subReg = subscriberHandlersRegistry{new(sync.Mutex), 0, make(map[int]DataHandler)}

//export callSubscriberDataHandler
func callSubscriberDataHandler(rkey *C.zn_resource_key_t, data unsafe.Pointer, length C.size_t, info *C.zn_data_info_t, arg unsafe.Pointer) {
	var rname string
	if rkey.kind == C.ZN_STR_RES_KEY {
		rname = resKeyToRName(rkey.key)
	} else {
		fmt.Printf("INTERNAL ERROR: DataHandler received a non-string zn_resource_key_t with kind=%d\n", rkey.kind)
		return
	}

	dataSlice := C.GoBytes(data, C.int(length))

	// Note: 'arg' parameter is used to store the index of handler in subReg.dHandler. Don't use it as a real C memory address !!
	index := uintptr(arg)
	goHandler := subReg.dHandler[int(index)]

	defer func() {
		if r := recover(); r != nil {
			logger.WithField("resource", rname).WithField("error", r).Warn("error in subscriber data handler for")
			debug.PrintStack()
		}
	}()

	goHandler(rname, dataSlice, info)
}

// DeclareSubscriber declares a subscription for all published data matching the provided resource name 'resource'.
// 'resource' is the resource name to subscribe to.
// 'mode' is the subscription mode.
// 'dataHandler' is the callback function that will be called each time a data matching the subscribed resource name 'resource' is received.
// Return a zenoh subscriber.
func (s *Session) DeclareSubscriber(resource string, mode SubMode, dataHandler DataHandler) (*Subscriber, error) {
	logger.WithField("resource", resource).Debug("DeclareSubscriber")

	r := C.CString(resource)
	defer C.free(unsafe.Pointer(r))

	subReg.mu.Lock()
	defer subReg.mu.Unlock()
	subReg.index++
	for subReg.dHandler[subReg.index] != nil {
		subReg.index++
	}
	subReg.dHandler[subReg.index] = dataHandler

	// Note: 'arg' parameter is used to store the index of handler in subReg.dHandler. Don't use it as a real C memory address !!
	result := C.zn_declare_subscriber(s, r, &mode,
		(C.zn_data_handler_t)(unsafe.Pointer(C.subscriber_handle_data_cgo)),
		unsafe.Pointer(uintptr(subReg.index)))
	if result.tag == C.Z_ERROR_TAG {
		delete(subReg.dHandler, subReg.index)
		return nil, &ZError{"zn_declare_subscriber for " + resource + " failed", resultValueToErrorCode(result.value), nil}
	}

	sub := new(Subscriber)
	sub.zsub = resultValueToSubscriber(result.value)
	sub.regIndex = subReg.index

	return sub, nil
}

// DeclarePublisher declares a publication for resource name 'resource'.
// 'resource' is the resource name to publish.
// Return a zenoh publisher.
func (s *Session) DeclarePublisher(resource string) (*Publisher, error) {
	logger.WithField("resource", resource).Debug("DeclarePublisher")

	r := C.CString(resource)
	defer C.free(unsafe.Pointer(r))

	result := C.zn_declare_publisher(s, r)
	if result.tag == C.Z_ERROR_TAG {
		return nil, &ZError{"zn_declare_publisher for " + resource + " failed", resultValueToErrorCode(result.value), nil}
	}

	return resultValueToPublisher(result.value), nil
}

type storageHandlersRegistry struct {
	mu       *sync.Mutex
	index    int
	dHandler map[int]DataHandler
	qHandler map[int]QueryHandler
}

var stoHdlReg = storageHandlersRegistry{new(sync.Mutex), 0, make(map[int]DataHandler), make(map[int]QueryHandler)}

//export callStorageDataHandler
func callStorageDataHandler(rkey *C.zn_resource_key_t, data unsafe.Pointer, length C.size_t, info *C.zn_data_info_t, arg unsafe.Pointer) {
	var rname string
	if rkey.kind == C.ZN_STR_RES_KEY {
		rname = resKeyToRName(rkey.key)
	} else {
		fmt.Printf("INTERNAL ERROR: DataHandler received a non-string zn_resource_key_t with kind=%d\n", rkey.kind)
		return
	}

	dataSlice := C.GoBytes(data, C.int(length))

	// Note: 'arg' parameter is used to store the index of handler in stoHdlReg.subCb. Don't use it as a real C memory address !!
	index := uintptr(arg)
	goHandler := stoHdlReg.dHandler[int(index)]

	defer func() {
		if r := recover(); r != nil {
			logger.WithField("resource", rname).WithField("error", r).Warn("error in storage data handler")
			debug.PrintStack()
		}
	}()

	goHandler(rname, dataSlice, info)
}

//export callStorageQueryHandler
func callStorageQueryHandler(rname *C.char, predicate *C.char, sendReplies unsafe.Pointer, queryHandle unsafe.Pointer, arg unsafe.Pointer) {
	goRname := C.GoString(rname)
	goPredicate := C.GoString(predicate)
	goRepliesSender := new(RepliesSender)
	goRepliesSender.sendRepliesFunc = C.zn_replies_sender_t(sendReplies)
	goRepliesSender.queryHandle = queryHandle

	// Note: 'arg' parameter is used to store the index of handler in stoHdlReg.qHandler. Don't use it as a real C memory address !!
	index := uintptr(arg)
	goHandler := stoHdlReg.qHandler[int(index)]

	defer func() {
		if r := recover(); r != nil {
			logger.WithField("resource", rname).WithField("error", r).Warn("error in query handle for storage")
			debug.PrintStack()
			goRepliesSender.SendReplies([]Resource{})
		}
	}()

	goHandler(goRname, goPredicate, goRepliesSender)
}

// DeclareStorage declares a storage for all data matching the provided resource name 'resource'.
// 'resource' is the resource selection to store.
// 'dataHandler' is the callback function that will be called each time a data matching the stored resource name 'resource' is received.
// 'queryHandler' is the callback function that will be called each time a query for data matching the stored resource name 'resource' is received.
// The 'queryHandler' function MUST call the provided 'RepliesSender.SendReplies()' function with the resulting data.
// 'RepliesSender.SendReplies()' can be called with an empty array.
// Return a zenoh storage.
func (s *Session) DeclareStorage(resource string, dataHandler DataHandler, queryHandler QueryHandler) (*Storage, error) {
	logger.WithField("resource", resource).Debug("DeclareStorage")

	r := C.CString(resource)
	defer C.free(unsafe.Pointer(r))

	stoHdlReg.mu.Lock()
	defer stoHdlReg.mu.Unlock()

	stoHdlReg.index++
	for stoHdlReg.dHandler[stoHdlReg.index] != nil {
		stoHdlReg.index++
	}
	stoHdlReg.dHandler[stoHdlReg.index] = dataHandler
	stoHdlReg.qHandler[stoHdlReg.index] = queryHandler

	// Note: 'arg' parameter is used to store the index of handler in stoHdlReg. Don't use it as a real C memory address !!
	result := C.zn_declare_storage(s, r,
		(C.zn_data_handler_t)(unsafe.Pointer(C.storage_handle_data_cgo)),
		(C.zn_query_handler_t)(unsafe.Pointer(C.storage_handle_query_cgo)),
		unsafe.Pointer(uintptr(stoHdlReg.index)))
	if result.tag == C.Z_ERROR_TAG {
		delete(stoHdlReg.dHandler, stoHdlReg.index)
		delete(stoHdlReg.qHandler, stoHdlReg.index)
		return nil, &ZError{"zn_declare_storage for " + resource + " failed", resultValueToErrorCode(result.value), nil}
	}

	storage := new(Storage)
	storage.zsto = resultValueToStorage(result.value)
	storage.regIndex = subReg.index

	return storage, nil
}

type evalHandlersRegistry struct {
	mu       *sync.Mutex
	index    int
	qHandler map[int]QueryHandler
}

var evalHdlReg = evalHandlersRegistry{new(sync.Mutex), 0, make(map[int]QueryHandler)}

//export callEvalQueryHandler
func callEvalQueryHandler(rname *C.char, predicate *C.char, sendReplies unsafe.Pointer, queryHandle unsafe.Pointer, arg unsafe.Pointer) {
	goRname := C.GoString(rname)
	goPredicate := C.GoString(predicate)
	goRepliesSender := new(RepliesSender)
	goRepliesSender.sendRepliesFunc = C.zn_replies_sender_t(sendReplies)
	goRepliesSender.queryHandle = queryHandle

	// Note: 'arg' parameter is used to store the index of handler in evalHdlReg.qHandler. Don't use it as a real C memory address !!
	index := uintptr(arg)
	goHandler := evalHdlReg.qHandler[int(index)]

	defer func() {
		if r := recover(); r != nil {
			logger.WithField("resource", rname).WithField("error", r).Warn("error in query handle for eval")
			debug.PrintStack()
			goRepliesSender.SendReplies([]Resource{})
		}
	}()

	goHandler(goRname, goPredicate, goRepliesSender)
}

// DeclareEval declares an eval able to provide data matching the provided resource name 'resource'.
// 'resource' is the resource to evaluate.
// 'handler' is the callback function that will be called each time a query for data matching the evaluated resource name 'resource' is received.
// The 'handler' function MUST call the provided 'sendReplies' function with the resulting data. 'sendReplies'can be called with an empty array.
// Return a zenoh-net eval.
func (s *Session) DeclareEval(resource string, handler QueryHandler) (*Eval, error) {
	logger.WithField("resource", resource).Debug("DeclareEval")

	r := C.CString(resource)
	defer C.free(unsafe.Pointer(r))

	evalHdlReg.mu.Lock()
	defer evalHdlReg.mu.Unlock()

	evalHdlReg.index++
	for evalHdlReg.qHandler[evalHdlReg.index] != nil {
		evalHdlReg.index++
	}
	evalHdlReg.qHandler[evalHdlReg.index] = handler

	// Note: 'arg' parameter is used to store the index of handler in evalHdlReg. Don't use it as a real C memory address !!
	result := C.zn_declare_eval(s, r,
		(C.zn_query_handler_t)(unsafe.Pointer(C.eval_handle_query_cgo)),
		unsafe.Pointer(uintptr(evalHdlReg.index)))
	if result.tag == C.Z_ERROR_TAG {
		delete(evalHdlReg.qHandler, evalHdlReg.index)
		return nil, &ZError{"zn_declare_eval for " + resource + " failed", resultValueToErrorCode(result.value), nil}
	}

	eval := new(Eval)
	eval.zeval = resultValueToEval(result.value)
	eval.regIndex = evalHdlReg.index

	return eval, nil
}

// StreamCompactData sends data in a 'compact_data' message for the resource published by publisher 'p'.
// 'payload' is the data to be sent.
func (p *Publisher) StreamCompactData(payload []byte) error {
	b, l := bufferToC(payload)
	result := C.zn_stream_compact_data(p, b, l)
	if result != 0 {
		return &ZError{"zn_stream_compact_data of " + strconv.Itoa(len(payload)) + " bytes buffer failed", int(result), nil}
	}
	return nil
}

// StreamData sends data in a 'stream_data' message for the resource published by publisher 'p'.
// 'payload' is the data to be sent.
func (p *Publisher) StreamData(payload []byte) error {
	b, l := bufferToC(payload)
	result := C.zn_stream_data(p, b, l)
	if result != 0 {
		return &ZError{"zn_stream_data of " + strconv.Itoa(len(payload)) + " bytes buffer failed", int(result), nil}
	}
	return nil
}

// WriteData sends data in a 'write_data' message for the resource 'resource'.
// 'resource' is the resource name of the data to be sent.
// 'payload' is the data to be sent.
func (s *Session) WriteData(resource string, payload []byte) error {
	r := C.CString(resource)
	defer C.free(unsafe.Pointer(r))

	b, l := bufferToC(payload)
	result := C.zn_write_data(s, r, b, l)
	if result != 0 {
		return &ZError{"zn_write_data of " + strconv.Itoa(len(payload)) + " bytes buffer on " + resource + "failed", int(result), nil}
	}
	return nil
}

// StreamDataWO sends data in a 'stream_data' message for the resource published by publisher 'p'.
// 'payload' is the data to be sent.
// 'encoding' is a metadata information associated with the published data that represents the encoding of the published data.
// 'kind' is a metadata information associated with the published data that represents the kind of publication.
func (p *Publisher) StreamDataWO(payload []byte, encoding uint8, kind uint8) error {
	b, l := bufferToC(payload)
	result := C.zn_stream_data_wo(p, b, l, C.uchar(encoding), C.uchar(kind))
	if result != 0 {
		return &ZError{"zn_stream_data_wo of " + strconv.Itoa(len(payload)) + " bytes buffer failed", int(result), nil}
	}
	return nil
}

// WriteDataWO sends data in a 'write_data' message for the resource 'resource'.
// 'resource' is the resource name of the data to be sent.
// 'payload' is the data to be sent.
// 'encoding' is a metadata information associated with the published data that represents the encoding of the published data.
// 'kind' is a metadata information associated with the published data that represents the kind of publication.
func (s *Session) WriteDataWO(resource string, payload []byte, encoding uint8, kind uint8) error {
	r := C.CString(resource)
	defer C.free(unsafe.Pointer(r))

	b, l := bufferToC(payload)
	result := C.zn_write_data_wo(s, r, b, l, C.uchar(encoding), C.uchar(kind))
	if result != 0 {
		return &ZError{"zn_write_data_wo of " + strconv.Itoa(len(payload)) + " bytes buffer on " + resource + "failed", int(result), nil}
	}
	return nil
}

// Pull data for the `ZPullMode` or `ZPeriodicPullMode` subscription 's'. The pulled data will be provided
// by calling the 'dataHandler' function provided to the `DeclareSubscriber` function.
func (s *Subscriber) Pull() error {
	result := C.zn_pull(s.zsub)
	if result != 0 {
		return &ZError{"zn_pull failed", int(result), nil}
	}
	return nil
}

var nullCPtr = (*C.uchar)(unsafe.Pointer(nil))

func bufferToC(buf []byte) (*C.uchar, C.ulong) {
	if buf == nil {
		return nullCPtr, C.ulong(0)
	}
	return (*C.uchar)(unsafe.Pointer(&buf[0])), C.ulong(len(buf))
}

// RNameIntersect returns true if the resource name 'rname1' intersects with the resource name 'rname2'.
func RNameIntersect(rname1 string, rname2 string) bool {
	r1 := C.CString(rname1)
	defer C.free(unsafe.Pointer(r1))
	r2 := C.CString(rname2)
	defer C.free(unsafe.Pointer(r2))

	return C.zn_rname_intersect(r1, r2) != 0
}

type replyHandlersRegistry struct {
	mu       *sync.Mutex
	index    int
	rHandler map[int]ReplyHandler
}

var replyReg = replyHandlersRegistry{new(sync.Mutex), 0, make(map[int]ReplyHandler)}

//export callReplyHandler
func callReplyHandler(reply *C.zn_reply_value_t, arg unsafe.Pointer) {
	index := uintptr(arg)
	goHandler := replyReg.rHandler[int(index)]
	goHandler(reply)
}

// Query queries data matching resource name 'resource'.
// 'resource' is the resource to query.
// 'predicate' is a string that will be  propagated to the storages and evals that should provide the queried data.
// It may allow them to filter, transform and/or compute the queried data.
// 'replyHandler' is the callback function that will be called on reception of the replies of the query.
func (s *Session) Query(resource string, predicate string, replyHandler ReplyHandler) error {
	r := C.CString(resource)
	defer C.free(unsafe.Pointer(r))
	p := C.CString(predicate)
	defer C.free(unsafe.Pointer(p))

	replyReg.mu.Lock()
	defer replyReg.mu.Unlock()
	replyReg.index++
	for replyReg.rHandler[replyReg.index] != nil {
		replyReg.index++
	}
	replyReg.rHandler[replyReg.index] = replyHandler

	result := C.zn_query(s, r, p,
		(C.zn_reply_handler_t)(unsafe.Pointer(C.handle_reply_cgo)),
		unsafe.Pointer(uintptr(replyReg.index)))
	if result != 0 {
		return &ZError{"zn_query on " + resource + "failed", int(result), nil}
	}
	return nil
}

// QueryWO queries data matching resource name 'resource'.
// 'resource' is the resource to query.
// 'predicate' is a string that will be  propagated to the storages and evals that should provide the queried data.
// It may allow them to filter, transform and/or compute the queried data.
// 'replyHandler' is the callback function that will be called on reception of the replies of the query.
// 'destStorages' indicates which matching storages should be destination of the query.
// 'destEvals' indicates which matching evals should be destination of the query.
func (s *Session) QueryWO(resource string, predicate string, replyHandler ReplyHandler, destStorages QueryDest, destEvals QueryDest) error {
	r := C.CString(resource)
	defer C.free(unsafe.Pointer(r))
	p := C.CString(predicate)
	defer C.free(unsafe.Pointer(p))

	replyReg.mu.Lock()
	defer replyReg.mu.Unlock()
	replyReg.index++
	for replyReg.rHandler[replyReg.index] != nil {
		replyReg.index++
	}
	replyReg.rHandler[replyReg.index] = replyHandler

	result := C.zn_query_wo(s, r, p,
		(C.zn_reply_handler_t)(unsafe.Pointer(C.handle_reply_cgo)),
		unsafe.Pointer(uintptr(replyReg.index)),
		destStorages, destEvals)
	if result != 0 {
		return &ZError{"zn_query on " + resource + "failed", int(result), nil}
	}
	return nil
}

// UndeclareSubscriber undeclares the subscription 's'.
func (s *Session) UndeclareSubscriber(sub *Subscriber) error {
	result := C.zn_undeclare_subscriber(sub.zsub)
	if result != 0 {
		return &ZError{"zn_undeclare_subscriber failed", int(result), nil}
	}
	subReg.mu.Lock()
	delete(subReg.dHandler, sub.regIndex)
	subReg.mu.Unlock()

	return nil
}

// UndeclarePublisher undeclares the publication 'p'.
func (s *Session) UndeclarePublisher(p *Publisher) error {
	result := C.zn_undeclare_publisher(p)
	if result != 0 {
		return &ZError{"zn_undeclare_publisher failed", int(result), nil}
	}
	return nil
}

// UndeclareStorage undeclares the storage 's'.
func (s *Session) UndeclareStorage(sto *Storage) error {
	result := C.zn_undeclare_storage(sto.zsto)
	if result != 0 {
		return &ZError{"zn_undeclare_storage failed", int(result), nil}
	}
	stoHdlReg.mu.Lock()
	delete(stoHdlReg.dHandler, sto.regIndex)
	delete(stoHdlReg.qHandler, sto.regIndex)
	stoHdlReg.mu.Unlock()

	return nil
}

// UndeclareEval undeclares the eval 'e'.
func (s *Session) UndeclareEval(e *Eval) error {
	result := C.zn_undeclare_eval(e.zeval)
	if result != 0 {
		return &ZError{"zn_undeclare_eval failed", int(result), nil}
	}
	evalHdlReg.mu.Lock()
	delete(evalHdlReg.qHandler, e.regIndex)
	evalHdlReg.mu.Unlock()

	return nil
}

func resultValueToErrorCode(cbytes [8]byte) int {
	buf := bytes.NewBuffer(cbytes[:])
	var code C.int
	if err := binary.Read(buf, binary.LittleEndian, &code); err == nil {
		return int(code)
	}
	return -42
}

func resultValueToSession(cbytes [8]byte) *Session {
	buf := bytes.NewBuffer(cbytes[:])
	var ptr uint64
	if err := binary.Read(buf, binary.LittleEndian, &ptr); err == nil {
		uptr := uintptr(ptr)
		return (*Session)(unsafe.Pointer(uptr))
	}
	return nil
}

// resultValueToPublisher gets the Publisher (zn_pub_t) from a zn_pub_p_result_t.value (union type)
func resultValueToPublisher(cbytes [8]byte) *Publisher {
	buf := bytes.NewBuffer(cbytes[:])
	var ptr uint64
	if err := binary.Read(buf, binary.LittleEndian, &ptr); err == nil {
		uptr := uintptr(ptr)
		return (*Publisher)(unsafe.Pointer(uptr))
	}
	return nil
}

// resultValueToSubscriber gets the Subscriber (zn_sub_t) from a zn_sub_p_result_t.value (union type)
func resultValueToSubscriber(cbytes [8]byte) *C.zn_sub_t {
	buf := bytes.NewBuffer(cbytes[:])
	var ptr uint64
	if err := binary.Read(buf, binary.LittleEndian, &ptr); err == nil {
		uptr := uintptr(ptr)
		return (*C.zn_sub_t)(unsafe.Pointer(uptr))
	}
	return nil
}

// resultValueToStorage gets the Storage (zn_sto_t) from a zn_sto_p_result_t.value (union type)
func resultValueToStorage(cbytes [8]byte) *C.zn_sto_t {
	buf := bytes.NewBuffer(cbytes[:])
	var ptr uint64
	if err := binary.Read(buf, binary.LittleEndian, &ptr); err == nil {
		uptr := uintptr(ptr)
		return (*C.zn_sto_t)(unsafe.Pointer(uptr))
	}
	return nil
}

// resultValueToEval gets the Eval (zn_eva_t) from a zn_eval_p_result_t.value (union type)
func resultValueToEval(cbytes [8]byte) *C.zn_eva_t {
	buf := bytes.NewBuffer(cbytes[:])
	var ptr uint64
	if err := binary.Read(buf, binary.LittleEndian, &ptr); err == nil {
		uptr := uintptr(ptr)
		return (*C.zn_eva_t)(unsafe.Pointer(uptr))
	}
	return nil
}

// resKeyToRName gets the rname (string) from a zn_res_key_t (union type)
func resKeyToRName(cbytes [8]byte) string {
	buf := bytes.NewBuffer(cbytes[:])
	var ptr uint64
	if err := binary.Read(buf, binary.LittleEndian, &ptr); err == nil {
		uptr := uintptr(ptr)
		return C.GoString((*C.char)(unsafe.Pointer(uptr)))
	}
	panic("resKeyToRName: failed to read 64bits pointer from zn_res_key_t union (represented as a [8]byte)")

}
