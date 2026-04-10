//
// Copyright (c) 2026 ZettaScale Technology
//
// This program and the accompanying materials are made available under the
// terms of the Eclipse Public License 2.0 which is available at
// http://www.eclipse.org/legal/epl-2.0, or the Apache License, Version 2.0
// which is available at https://www.apache.org/licenses/LICENSE-2.0.
//
// SPDX-License-Identifier: EPL-2.0 OR Apache-2.0
//
// Contributors:
//   ZettaScale Zenoh Team, <zenoh@zettascale.tech>
//

package zenoh

// #include "zenoh.h"
// #include "zenoh_cgo.h"
import "C"
import (
	"unsafe"

	"github.com/eclipse-zenoh/zenoh-go/zenoh/internal"
)

// Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release.
//
// A Zenoh transport. Represents a network-level connection to a remote peer.
// This is a pure Go snapshot: all fields are extracted from the C object at callback/query time,
// so there are no C resources to manage.
type Transport struct {
	zId         Id
	whatAmI     WhatAmI
	isQos       bool
	isMulticast bool
	isShm       bool
}

// Return the Zenoh ID of the remote peer.
func (t Transport) ZId() Id {
	return t.zId
}

// Return what kind of Zenoh entity is at the other end of the transport.
func (t Transport) WhatAmI() WhatAmI {
	return t.whatAmI
}

// Return whether Quality of Service is enabled for this transport.
func (t Transport) IsQos() bool {
	return t.isQos
}

// Return whether this transport is multicast.
func (t Transport) IsMulticast() bool {
	return t.isMulticast
}

// Return whether shared memory is enabled for this transport.
func (t Transport) IsShm() bool {
	return t.isShm
}

// toCPtr creates a C-owned transport from the Go Transport snapshot.
// No C heap allocation — the local z_owned_transport_t escapes to the heap via Go's escape analysis.
func (t Transport) toCPtr() *C.z_owned_transport_t {
	var owned C.z_owned_transport_t
	C.zc_cgo_create_transport(&owned, t.zId.id, C.z_whatami_t(t.whatAmI),
		C.bool(t.isQos), C.bool(t.isMulticast), C.bool(t.isShm))
	return &owned
}

// extractTransportSnapshot extracts all fields from a loaned C transport into a pure Go Transport.
func extractTransportSnapshot(loanedTransport *C.z_loaned_transport_t) Transport {
	return Transport{
		zId:         Id{id: C.z_transport_zid(loanedTransport)},
		whatAmI:     WhatAmI(C.z_transport_whatami(loanedTransport)),
		isQos:       bool(C.z_transport_is_qos(loanedTransport)),
		isMulticast: bool(C.z_transport_is_multicast(loanedTransport)),
		isShm:       bool(C.zc_cgo_transport_is_shm(loanedTransport)),
	}
}

// Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release.
//
// A transport event. Indicates a transport was connected or disconnected.
// This is a pure Go snapshot: all fields are extracted from the C object at event time,
// so there are no C resources to manage.
type TransportEvent struct {
	kind      SampleKind
	transport Transport
}

// Return the kind of the event. [SampleKindPut] means transport connected, [SampleKindDelete] means disconnected.
func (e *TransportEvent) Kind() SampleKind {
	return e.kind
}

// Return the transport associated with this event.
func (e *TransportEvent) Transport() Transport {
	return e.transport
}

//export zenohTransportEventsCallback
func zenohTransportEventsCallback(event *C.z_loaned_transport_event_t, context unsafe.Pointer) {
	kind := SampleKind(C.z_transport_event_kind(event))
	loanedTransport := C.z_transport_event_transport(event)
	evt := TransportEvent{
		kind:      kind,
		transport: extractTransportSnapshot(loanedTransport),
	}
	(*internal.ClosureContext[TransportEvent])(context).Call(evt)
}

//export zenohTransportEventsDrop
func zenohTransportEventsDrop(context unsafe.Pointer) {
	(*internal.ClosureContext[TransportEvent])(context).Drop()
}

//export zenohTransportCallback
func zenohTransportCallback(transport *C.z_loaned_transport_t, context unsafe.Pointer) {
	(*internal.ClosureContext[Transport])(context).Call(extractTransportSnapshot(transport))
}

//export zenohTransportDrop
func zenohTransportDrop(context unsafe.Pointer) {
	(*internal.ClosureContext[Transport])(context).Drop()
}

// Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release.
//
// A Zenoh transport events listener.
//
// A listener that sends notifications when transports are connected or disconnected.
type TransportEventsListener struct {
	listener *C.z_owned_transport_events_listener_t
	receiver <-chan TransportEvent
}

// Return transport events listener's receiver if it was constructed with channel, nil otherwise.
func (listener *TransportEventsListener) Handler() <-chan TransportEvent {
	return listener.receiver
}

// Destroy the transport events listener.
// This is equivalent to calling [TransportEventsListener.Undeclare] and discarding its return value.
func (listener *TransportEventsListener) Drop() {
	C.z_transport_events_listener_drop(C.z_transport_events_listener_move(listener.listener))
}

// Undeclare and destroy the transport events listener.
func (listener *TransportEventsListener) Undeclare() error {
	res := int8(C.z_undeclare_transport_events_listener(C.z_transport_events_listener_move(listener.listener)))
	if res == 0 {
		return nil
	}
	return newZError(res)
}

// Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release.
//
// Options for declaring a transport events listener.
type TransportEventsListenerOptions struct {
	History bool // If true, receive events for already-existing transports.
}

// Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release.
//
// Fetch all transports currently connected to the session.
func (session Session) Transports() ([]Transport, error) {
	var transports []Transport

	closure := internal.NewClosure(func(t Transport) { transports = append(transports, t) }, nil)
	var cClosure C.z_owned_closure_transport_t
	C.z_closure_transport(&cClosure, (*[0]byte)(C.zenohTransportCallback), (*[0]byte)(C.zenohTransportDrop), unsafe.Pointer(closure))
	res := int8(C.z_info_transports(C.z_session_loan(session.session), C.z_closure_transport_move(&cClosure)))
	if res != 0 {
		return []Transport{}, newZError(res)
	}
	return transports, nil
}

// Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release.
//
// Declare a transport events listener, registering a handler for transport connect/disconnect notifications.
// Transport events listener MUST be explicitly destroyed using [TransportEventsListener.Undeclare] or [TransportEventsListener.Drop] once it is no longer needed.
func (session *Session) DeclareTransportEventsListener(handler Handler[TransportEvent], options *TransportEventsListenerOptions) (TransportEventsListener, error) {
	callback, drop, recv := handler.ToCbDropHandler()
	closure := internal.NewClosure(callback, drop)
	var cClosure C.z_owned_closure_transport_event_t
	C.z_closure_transport_event(&cClosure, (*[0]byte)(C.zenohTransportEventsCallback), (*[0]byte)(C.zenohTransportEventsDrop), unsafe.Pointer(closure))

	var cListener C.z_owned_transport_events_listener_t
	var res int8
	if options == nil {
		res = int8(C.z_declare_transport_events_listener(C.z_session_loan(session.session), &cListener, C.z_closure_transport_event_move(&cClosure), nil))
	} else {
		var cOpts C.z_transport_events_listener_options_t
		C.z_transport_events_listener_options_default(&cOpts)
		cOpts.history = C.bool(options.History)
		res = int8(C.z_declare_transport_events_listener(C.z_session_loan(session.session), &cListener, C.z_closure_transport_event_move(&cClosure), &cOpts))
	}

	if res == 0 {
		return TransportEventsListener{listener: &cListener, receiver: recv}, nil
	}
	return TransportEventsListener{}, newZError(res)
}

// Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release.
//
// Declare a background transport events listener, registering a callback for transport connect/disconnect notifications.
// The callback will be run in the background until the corresponding session is closed or dropped.
func (session *Session) DeclareBackgroundTransportEventsListener(closure Closure[TransportEvent], options *TransportEventsListenerOptions) error {
	listenerClosure := internal.NewClosure(closure.Call, closure.Drop)
	var cClosure C.z_owned_closure_transport_event_t
	C.z_closure_transport_event(&cClosure, (*[0]byte)(C.zenohTransportEventsCallback), (*[0]byte)(C.zenohTransportEventsDrop), unsafe.Pointer(listenerClosure))

	var res int8
	if options == nil {
		res = int8(C.z_declare_background_transport_events_listener(C.z_session_loan(session.session), C.z_closure_transport_event_move(&cClosure), nil))
	} else {
		var cOpts C.z_transport_events_listener_options_t
		C.z_transport_events_listener_options_default(&cOpts)
		cOpts.history = C.bool(options.History)
		res = int8(C.z_declare_background_transport_events_listener(C.z_session_loan(session.session), C.z_closure_transport_event_move(&cClosure), &cOpts))
	}

	if res == 0 {
		return nil
	}
	return newZError(res)
}
