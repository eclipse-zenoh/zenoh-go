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
	"runtime"
	"unsafe"

	"github.com/BooleanCat/option"
	"github.com/eclipse-zenoh/zenoh-go/zenoh/internal"
)

// Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release.
//
// A priority range supported by a link (when QoS is enabled).
type PriorityRange struct {
	Min uint8
	Max uint8
}

// Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release.
//
// A Zenoh link. Represents a physical network link within a transport.
// This is a pure Go snapshot: all fields are extracted from the C object at callback/query time,
// so there are no C resources to manage.
type Link struct {
	zId            Id
	src            string
	dst            string
	group          option.Option[string]
	mtu            uint16
	isStreamed     bool
	interfaces     []string
	authIdentifier option.Option[string]
	priorities     option.Option[PriorityRange]
	reliability    option.Option[Reliability]
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

// Return the group locator of the link (for multicast links). None if not applicable.
func (l Link) Group() option.Option[string] {
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

// Return the authentication identifier of the link. None if not available.
func (l Link) AuthIdentifier() option.Option[string] {
	return l.authIdentifier
}

// Return the priority range supported by the link if QoS is enabled.
// Returns Some(PriorityRange) if priorities are supported, None otherwise.
func (l Link) Priorities() option.Option[PriorityRange] {
	return l.priorities
}

// Return the reliability of the link if QoS is enabled.
// Returns Some(Reliability) if available, None otherwise.
func (l Link) Reliability() option.Option[Reliability] {
	return l.reliability
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
		l.group = option.Some(C.GoStringN(C.z_string_data(loanedStr), C.int(C.z_string_len(loanedStr))))
	}
	C.zc_cgo_string_drop(&s)

	// Extract auth identifier (optional).
	C.z_link_auth_identifier(loanedLink, &s)
	if bool(C.z_internal_string_check(&s)) {
		loanedStr = C.z_string_loan(&s)
		l.authIdentifier = option.Some(C.GoStringN(C.z_string_data(loanedStr), C.int(C.z_string_len(loanedStr))))
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
	if bool(C.z_link_priorities(loanedLink, &cMin, &cMax)) {
		l.priorities = option.Some(PriorityRange{Min: uint8(cMin), Max: uint8(cMax)})
	}

	// Extract reliability.
	var cReliability uint32
	if bool(C.z_link_reliability(loanedLink, (*uint32)(unsafe.Pointer(&cReliability)))) {
		l.reliability = option.Some(Reliability(cReliability))
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
		pinner := runtime.Pinner{}
		defer pinner.Unpin()
		ownedPtr := options.Transport.Unwrap().toCPtr()
		pinner.Pin(ownedPtr)
		cOpts.transport = C.z_transport_move(ownedPtr)
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
			pinner := runtime.Pinner{}
			defer pinner.Unpin()
			ownedPtr := options.Transport.Unwrap().toCPtr()
			pinner.Pin(ownedPtr)
			cOpts.transport = C.z_transport_move(ownedPtr)
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
			pinner := runtime.Pinner{}
			defer pinner.Unpin()
			ownedPtr := options.Transport.Unwrap().toCPtr()
			pinner.Pin(ownedPtr)
			cOpts.transport = C.z_transport_move(ownedPtr)
		}
		res = int8(C.z_declare_background_link_events_listener(C.z_session_loan(session.session), C.z_closure_link_event_move(&cClosure), &cOpts))
	}

	if res == 0 {
		return nil
	}
	return newZError(res)
}
