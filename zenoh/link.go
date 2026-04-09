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

	"github.com/BooleanCat/option"
	"github.com/eclipse-zenoh/zenoh-go/zenoh/internal"
)

// Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release.
//
// A Zenoh link. Represents a physical network link within a transport.
// This is a pure Go snapshot: all fields are extracted from the C object at callback/query time,
// so there are no C resources to manage.
type Link struct {
	zId            Id
	src            string
	dst            string
	group          string
	mtu            uint16
	isStreamed     bool
	interfaces     []string
	authIdentifier string
	priorityMin    uint8
	priorityMax    uint8
	hasPriorities  bool
	reliability    Reliability
	hasReliability bool
}

// Return the Zenoh ID of the transport this link belongs to.
func (l Link) ZId() Id {
	return l.zId
}

// Return the source locator (local endpoint) of the link.
func (l Link) Src() string {
	return l.src
}

// Return the destination locator (remote endpoint) of the link.
func (l Link) Dst() string {
	return l.dst
}

// Return the group locator of the link (for multicast links). Empty string if not applicable.
func (l Link) Group() string {
	return l.group
}

// Return the MTU (maximum transmission unit) of the link in bytes.
func (l Link) Mtu() uint16 {
	return l.mtu
}

// Return whether the link is streamed.
func (l Link) IsStreamed() bool {
	return l.isStreamed
}

// Return the network interfaces associated with the link.
func (l Link) Interfaces() []string {
	return l.interfaces
}

// Return the authentication identifier of the link. Empty string if not available.
func (l Link) AuthIdentifier() string {
	return l.authIdentifier
}

// Return the priority range supported by the link if QoS is enabled.
// Returns (min, max, true) if priorities are supported, (0, 0, false) otherwise.
func (l Link) Priorities() (min uint8, max uint8, ok bool) {
	return l.priorityMin, l.priorityMax, l.hasPriorities
}

// Return the reliability of the link if QoS is enabled.
// Returns (reliability, true) if available, (0, false) otherwise.
func (l Link) Reliability() (Reliability, bool) {
	return l.reliability, l.hasReliability
}

// extractLink extracts all fields from a loaned C link into a pure Go Link.
func extractLink(loanedLink *C.z_loaned_link_t) Link {
	l := Link{
		zId:        Id{id: C.z_link_zid(loanedLink)},
		mtu:        uint16(C.z_link_mtu(loanedLink)),
		isStreamed: bool(C.z_link_is_streamed(loanedLink)),
	}

	// Extract src string.
	var s C.z_owned_string_t
	C.z_link_src(loanedLink, &s)
	loanedStr := C.z_string_loan(&s)
	l.src = C.GoStringN(C.z_string_data(loanedStr), C.int(C.z_string_len(loanedStr)))
	C.zc_cgo_string_drop(&s)

	// Extract dst string.
	C.z_link_dst(loanedLink, &s)
	loanedStr = C.z_string_loan(&s)
	l.dst = C.GoStringN(C.z_string_data(loanedStr), C.int(C.z_string_len(loanedStr)))
	C.zc_cgo_string_drop(&s)

	// Extract group string (optional).
	C.z_link_group(loanedLink, &s)
	if bool(C.z_internal_string_check(&s)) {
		loanedStr = C.z_string_loan(&s)
		l.group = C.GoStringN(C.z_string_data(loanedStr), C.int(C.z_string_len(loanedStr)))
	}
	C.zc_cgo_string_drop(&s)

	// Extract auth identifier (optional).
	C.z_link_auth_identifier(loanedLink, &s)
	if bool(C.z_internal_string_check(&s)) {
		loanedStr = C.z_string_loan(&s)
		l.authIdentifier = C.GoStringN(C.z_string_data(loanedStr), C.int(C.z_string_len(loanedStr)))
	}
	C.zc_cgo_string_drop(&s)

	// Extract interfaces.
	var arr C.z_owned_string_array_t
	C.z_link_interfaces(loanedLink, &arr)
	loanedArr := C.z_string_array_loan(&arr)
	length := int(C.z_string_array_len(loanedArr))
	l.interfaces = make([]string, length)
	for i := 0; i < length; i++ {
		is := C.z_string_array_get(loanedArr, C.size_t(i))
		l.interfaces[i] = C.GoStringN(C.z_string_data(is), C.int(C.z_string_len(is)))
	}
	C.z_string_array_drop(C.z_string_array_move(&arr))

	// Extract priorities.
	var cMin, cMax C.uint8_t
	l.hasPriorities = bool(C.z_link_priorities(loanedLink, &cMin, &cMax))
	if l.hasPriorities {
		l.priorityMin = uint8(cMin)
		l.priorityMax = uint8(cMax)
	}

	// Extract reliability.
	var cReliability uint32
	l.hasReliability = bool(C.z_link_reliability(loanedLink, (*uint32)(unsafe.Pointer(&cReliability))))
	if l.hasReliability {
		l.reliability = Reliability(cReliability)
	}

	return l
}

// Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release.
//
// A link event. Indicates a link was added or removed.
// This is a pure Go snapshot: all fields are extracted from the C object at event time,
// so there are no C resources to manage.
type LinkEvent struct {
	kind SampleKind
	link Link
}

// Return the kind of the event. [SampleKindPut] means link added, [SampleKindDelete] means removed.
func (e *LinkEvent) Kind() SampleKind {
	return e.kind
}

// Return the link associated with this event.
func (e *LinkEvent) Link() Link {
	return e.link
}

// extractLinkSnapshot extracts all fields from a loaned link into a pure Go LinkEvent.
func extractLinkSnapshot(kind SampleKind, loanedLink *C.z_loaned_link_t) LinkEvent {
	return LinkEvent{
		kind: kind,
		link: extractLink(loanedLink),
	}
}

//export zenohLinkEventsCallback
func zenohLinkEventsCallback(event *C.z_loaned_link_event_t, context unsafe.Pointer) {
	kind := SampleKind(C.z_link_event_kind(event))
	loanedLink := C.z_link_event_link(event)
	(*internal.ClosureContext[LinkEvent])(context).Call(extractLinkSnapshot(kind, loanedLink))
}

//export zenohLinkEventsDrop
func zenohLinkEventsDrop(context unsafe.Pointer) {
	(*internal.ClosureContext[LinkEvent])(context).Drop()
}

//export zenohLinkCallback
func zenohLinkCallback(link *C.z_loaned_link_t, context unsafe.Pointer) {
	(*internal.ClosureContext[Link])(context).Call(extractLink(link))
}

//export zenohLinkDrop
func zenohLinkDrop(context unsafe.Pointer) {
	(*internal.ClosureContext[Link])(context).Drop()
}

// Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release.
//
// A Zenoh link events listener.
//
// A listener that sends notifications when links are added or removed.
type LinkEventsListener struct {
	listener *C.z_owned_link_events_listener_t
	receiver <-chan LinkEvent
}

// Return link events listener's receiver if it was constructed with channel, nil otherwise.
func (listener *LinkEventsListener) Handler() <-chan LinkEvent {
	return listener.receiver
}

// Destroy the link events listener.
// This is equivalent to calling [LinkEventsListener.Undeclare] and discarding its return value.
func (listener *LinkEventsListener) Drop() {
	C.z_link_events_listener_drop(C.z_link_events_listener_move(listener.listener))
}

// Undeclare and destroy the link events listener.
func (listener *LinkEventsListener) Undeclare() error {
	res := int8(C.z_undeclare_link_events_listener(C.z_link_events_listener_move(listener.listener)))
	if res == 0 {
		return nil
	}
	return newZError(res)
}

// Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release.
//
// Options for declaring a link events listener.
type LinkEventsListenerOptions struct {
	History   bool                    // If true, receive events for already-existing links.
	Transport option.Option[Transport] // Optional transport filter. If set, only receive events for links of this transport.
}

// Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release.
//
// Options for the synchronous [Session.Links] method.
type InfoLinksOptions struct {
	Transport option.Option[Transport] // Optional transport filter. If set, only return links of this transport.
}

// buildCTransport creates a C-heap-allocated owned transport from a Go Transport value using
// zc_internal_create_transport. The returned pointer is C-heap memory; the caller must call
// C.free on it after the C API has consumed it via z_transport_move (C takes ownership of the
// inner transport data, but not the shell — the caller must free the shell).
func buildCTransport(t Transport) *C.z_owned_transport_t {
	owned := (*C.z_owned_transport_t)(C.malloc(C.size_t(unsafe.Sizeof(C.z_owned_transport_t{}))))
	var cOpts C.zc_internal_create_transport_options_t
	C.zc_internal_create_transport_options_default(&cOpts)
	cOpts.zid = t.zId.id
	*(*uint32)(unsafe.Pointer(&cOpts.whatami)) = uint32(t.whatAmI)
	cOpts.is_qos = C.bool(t.isQos)
	cOpts.is_multicast = C.bool(t.isMulticast)
	C.zc_internal_create_transport(owned, &cOpts)
	return owned
}

// Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release.
//
// Fetch all links currently associated with the session.
func (session Session) Links(options *InfoLinksOptions) ([]Link, error) {
	var links []Link

	closure := internal.NewClosure(func(l Link) { links = append(links, l) }, nil)
	var cClosure C.z_owned_closure_link_t
	C.z_closure_link(&cClosure, (*[0]byte)(C.zenohLinkCallback), (*[0]byte)(C.zenohLinkDrop), unsafe.Pointer(closure))

	var res int8
	if options == nil || !options.Transport.IsSome() {
		res = int8(C.z_info_links(C.z_session_loan(session.session), C.z_closure_link_move(&cClosure), nil))
	} else {
		var cOpts C.z_info_links_options_t
		C.z_info_links_options_default(&cOpts)
		ownedPtr := buildCTransport(options.Transport.Unwrap())
		defer C.free(unsafe.Pointer(ownedPtr))
		cOpts.transport = C.z_transport_move(ownedPtr)
		// C takes ownership of inner transport data via move — do NOT drop ownedPtr's contents.
		res = int8(C.z_info_links(C.z_session_loan(session.session), C.z_closure_link_move(&cClosure), &cOpts))
	}
	if res != 0 {
		return []Link{}, newZError(res)
	}
	return links, nil
}

// Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release.
//
// Declare a link events listener, registering a handler for link add/remove notifications.
// Link events listener MUST be explicitly destroyed using [LinkEventsListener.Undeclare] or [LinkEventsListener.Drop] once it is no longer needed.
func (session *Session) DeclareLinkEventsListener(handler Handler[LinkEvent], options *LinkEventsListenerOptions) (LinkEventsListener, error) {
	callback, drop, recv := handler.ToCbDropHandler()
	closure := internal.NewClosure(callback, drop)
	var cClosure C.z_owned_closure_link_event_t
	C.z_closure_link_event(&cClosure, (*[0]byte)(C.zenohLinkEventsCallback), (*[0]byte)(C.zenohLinkEventsDrop), unsafe.Pointer(closure))

	var cListener C.z_owned_link_events_listener_t
	var res int8
	if options == nil {
		res = int8(C.z_declare_link_events_listener(C.z_session_loan(session.session), &cListener, C.z_closure_link_event_move(&cClosure), nil))
	} else {
		var cOpts C.z_link_events_listener_options_t
		C.z_link_events_listener_options_default(&cOpts)
		cOpts.history = C.bool(options.History)
		if options.Transport.IsSome() {
			ownedPtr := buildCTransport(options.Transport.Unwrap())
			defer C.free(unsafe.Pointer(ownedPtr))
			cOpts.transport = C.z_transport_move(ownedPtr)
			// C takes ownership of inner transport data via move — do NOT drop ownedPtr's contents.
		}
		res = int8(C.z_declare_link_events_listener(C.z_session_loan(session.session), &cListener, C.z_closure_link_event_move(&cClosure), &cOpts))
	}

	if res == 0 {
		return LinkEventsListener{listener: &cListener, receiver: recv}, nil
	}
	return LinkEventsListener{}, newZError(res)
}

// Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release.
//
// Declare a background link events listener, registering a callback for link add/remove notifications.
// The callback will be run in the background until the corresponding session is closed or dropped.
func (session *Session) DeclareBackgroundLinkEventsListener(closure Closure[LinkEvent], options *LinkEventsListenerOptions) error {
	listenerClosure := internal.NewClosure(closure.Call, closure.Drop)
	var cClosure C.z_owned_closure_link_event_t
	C.z_closure_link_event(&cClosure, (*[0]byte)(C.zenohLinkEventsCallback), (*[0]byte)(C.zenohLinkEventsDrop), unsafe.Pointer(listenerClosure))

	var res int8
	if options == nil {
		res = int8(C.z_declare_background_link_events_listener(C.z_session_loan(session.session), C.z_closure_link_event_move(&cClosure), nil))
	} else {
		var cOpts C.z_link_events_listener_options_t
		C.z_link_events_listener_options_default(&cOpts)
		cOpts.history = C.bool(options.History)
		if options.Transport.IsSome() {
			ownedPtr := buildCTransport(options.Transport.Unwrap())
			defer C.free(unsafe.Pointer(ownedPtr))
			cOpts.transport = C.z_transport_move(ownedPtr)
			// C takes ownership of inner transport data via move — do NOT drop ownedPtr's contents.
		}
		res = int8(C.z_declare_background_link_events_listener(C.z_session_loan(session.session), C.z_closure_link_event_move(&cClosure), &cOpts))
	}

	if res == 0 {
		return nil
	}
	return newZError(res)
}
