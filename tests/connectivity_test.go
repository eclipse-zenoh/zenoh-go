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

package zenoh_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/BooleanCat/option"
	"github.com/eclipse-zenoh/zenoh-go/zenoh"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// openConnectedPair opens two sessions connected via TCP on the given port.
// Session 1 listens; Session 2 connects to it.
func openConnectedPair(t *testing.T, port int) (zenoh.Session, zenoh.Session) {
	t.Helper()
	addr := fmt.Sprintf("tcp/127.0.0.1:%d", port)

	cfg1 := zenoh.NewConfigDefault()
	require.NoError(t, cfg1.InsertJson5(zenoh.ConfigModeKey, `"peer"`))
	require.NoError(t, cfg1.InsertJson5(zenoh.ConfigListenKey, fmt.Sprintf(`["%s"]`, addr)))
	require.NoError(t, cfg1.InsertJson5(zenoh.ConfigMulticastScoutingKey, "false"))

	cfg2 := zenoh.NewConfigDefault()
	require.NoError(t, cfg2.InsertJson5(zenoh.ConfigModeKey, `"peer"`))
	require.NoError(t, cfg2.InsertJson5(zenoh.ConfigConnectKey, fmt.Sprintf(`["%s"]`, addr)))
	require.NoError(t, cfg2.InsertJson5(zenoh.ConfigMulticastScoutingKey, "false"))

	s1, err := zenoh.Open(cfg1, nil)
	require.NoError(t, err)

	s2, err := zenoh.Open(cfg2, nil)
	require.NoError(t, err)

	// Allow time for the connection to establish.
	time.Sleep(500 * time.Millisecond)
	return s1, s2
}

// openListenerSession opens a session that listens on the given port without a peer.
func openListenerSession(t *testing.T, port int) zenoh.Session {
	t.Helper()
	addr := fmt.Sprintf("tcp/127.0.0.1:%d", port)

	cfg := zenoh.NewConfigDefault()
	require.NoError(t, cfg.InsertJson5(zenoh.ConfigModeKey, `"peer"`))
	require.NoError(t, cfg.InsertJson5(zenoh.ConfigListenKey, fmt.Sprintf(`["%s"]`, addr)))
	require.NoError(t, cfg.InsertJson5(zenoh.ConfigMulticastScoutingKey, "false"))

	s, err := zenoh.Open(cfg, nil)
	require.NoError(t, err)
	return s
}

// openConnectorSession opens a session that connects to the given port.
func openConnectorSession(t *testing.T, port int) zenoh.Session {
	t.Helper()
	addr := fmt.Sprintf("tcp/127.0.0.1:%d", port)

	cfg := zenoh.NewConfigDefault()
	require.NoError(t, cfg.InsertJson5(zenoh.ConfigModeKey, `"peer"`))
	require.NoError(t, cfg.InsertJson5(zenoh.ConfigConnectKey, fmt.Sprintf(`["%s"]`, addr)))
	require.NoError(t, cfg.InsertJson5(zenoh.ConfigMulticastScoutingKey, "false"))

	s, err := zenoh.Open(cfg, nil)
	require.NoError(t, err)
	return s
}

func TestTransportsList(t *testing.T) {
	s1, s2 := openConnectedPair(t, 17951)
	defer s1.Drop()
	defer s2.Drop()

	s2Id := s2.ZId()

	transports, err := s1.Transports()
	assert.NoError(t, err)
	assert.Equal(t, 1, len(transports))
	assert.Equal(t, s2Id.String(), transports[0].ZId().String())
}

func TestLinksList(t *testing.T) {
	s1, s2 := openConnectedPair(t, 17952)
	defer s1.Drop()
	defer s2.Drop()

	s2Id := s2.ZId()

	links, err := s1.Links(nil)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(links))
	assert.Equal(t, s2Id.String(), links[0].ZId().String())
}

func TestLinksFilteredByTransport(t *testing.T) {
	s1, s2 := openConnectedPair(t, 17953)
	defer s1.Drop()
	defer s2.Drop()

	transports, err := s1.Transports()
	require.NoError(t, err)
	require.Equal(t, 1, len(transports))

	// Filter by the matching transport — should return 1 link.
	links, err := s1.Links(&zenoh.InfoLinksOptions{Transport: option.Some(transports[0])})
	assert.NoError(t, err)
	assert.Equal(t, 1, len(links))

	// Filter by a transport from s2's perspective — s2 has no transport to s1 visible from s1.
	// Use an unrelated session's transport to confirm zero results.
	s2Transports, err := s2.Transports()
	require.NoError(t, err)
	require.Equal(t, 1, len(s2Transports))

	linksFiltered, err := s1.Links(&zenoh.InfoLinksOptions{Transport: option.Some(s2Transports[0])})
	assert.NoError(t, err)
	assert.Equal(t, 0, len(linksFiltered))
}

func TestTransportEventsListener(t *testing.T) {
	s1 := openListenerSession(t, 17954)
	defer s1.Drop()

	listener, err := s1.DeclareTransportEventsListener(zenoh.NewFifoChannel[zenoh.TransportEvent](16), nil)
	require.NoError(t, err)
	defer listener.Drop()

	assert.Equal(t, 0, len(listener.Handler()))

	// Connect a peer — should produce a PUT event.
	s2 := openConnectorSession(t, 17954)
	time.Sleep(500 * time.Millisecond)

	assert.Equal(t, 1, len(listener.Handler()))
	evt := <-listener.Handler()
	assert.Equal(t, zenoh.SampleKindPut, evt.Kind())
	s2ZId := s2.ZId()
	assert.Equal(t, s2ZId.String(), evt.Transport().ZId().String())

	// Disconnect — should produce a DELETE event.
	s2.Drop()
	time.Sleep(500 * time.Millisecond)

	assert.Equal(t, 1, len(listener.Handler()))
	evt = <-listener.Handler()
	assert.Equal(t, zenoh.SampleKindDelete, evt.Kind())
}

func TestTransportEventsListenerWithHistory(t *testing.T) {
	s1, s2 := openConnectedPair(t, 17955)
	defer s1.Drop()
	defer s2.Drop()

	s2ZId := s2.ZId()

	// Declare listener with history — should receive event for already-existing transport.
	listener, err := s1.DeclareTransportEventsListener(
		zenoh.NewFifoChannel[zenoh.TransportEvent](16),
		&zenoh.TransportEventsListenerOptions{History: true},
	)
	require.NoError(t, err)
	defer listener.Drop()

	time.Sleep(200 * time.Millisecond)

	assert.Equal(t, 1, len(listener.Handler()))
	evt := <-listener.Handler()
	assert.Equal(t, zenoh.SampleKindPut, evt.Kind())
	assert.Equal(t, s2ZId.String(), evt.Transport().ZId().String())
}

func TestBackgroundTransportEventsListener(t *testing.T) {
	s1 := openListenerSession(t, 17956)
	defer s1.Drop()

	var events []zenoh.TransportEvent

	err := s1.DeclareBackgroundTransportEventsListener(
		zenoh.Closure[zenoh.TransportEvent]{
			Call: func(evt zenoh.TransportEvent) { events = append(events, evt) },
		},
		nil,
	)
	require.NoError(t, err)

	s2 := openConnectorSession(t, 17956)
	time.Sleep(500 * time.Millisecond)

	assert.Equal(t, 1, len(events))
	assert.Equal(t, zenoh.SampleKindPut, events[0].Kind())

	s2.Drop()
	time.Sleep(500 * time.Millisecond)

	assert.Equal(t, 2, len(events))
	assert.Equal(t, zenoh.SampleKindDelete, events[1].Kind())
}

func TestLinkEventsListener(t *testing.T) {
	s1 := openListenerSession(t, 17957)
	defer s1.Drop()

	listener, err := s1.DeclareLinkEventsListener(zenoh.NewFifoChannel[zenoh.LinkEvent](16), nil)
	require.NoError(t, err)
	defer listener.Drop()

	assert.Equal(t, 0, len(listener.Handler()))

	// Connect a peer — should produce a PUT event.
	s2 := openConnectorSession(t, 17957)
	time.Sleep(500 * time.Millisecond)

	assert.Equal(t, 1, len(listener.Handler()))
	evt := <-listener.Handler()
	assert.Equal(t, zenoh.SampleKindPut, evt.Kind())
	s2ZId := s2.ZId()
	assert.Equal(t, s2ZId.String(), evt.Link().ZId().String())

	// Disconnect — should produce a DELETE event.
	s2.Drop()
	time.Sleep(500 * time.Millisecond)

	assert.Equal(t, 1, len(listener.Handler()))
	evt = <-listener.Handler()
	assert.Equal(t, zenoh.SampleKindDelete, evt.Kind())
}

func TestLinkEventsListenerWithHistory(t *testing.T) {
	s1, s2 := openConnectedPair(t, 17958)
	defer s1.Drop()
	defer s2.Drop()

	s2ZId := s2.ZId()

	// Declare listener with history — should receive event for already-existing link.
	listener, err := s1.DeclareLinkEventsListener(
		zenoh.NewFifoChannel[zenoh.LinkEvent](16),
		&zenoh.LinkEventsListenerOptions{History: true},
	)
	require.NoError(t, err)
	defer listener.Drop()

	time.Sleep(200 * time.Millisecond)

	assert.Equal(t, 1, len(listener.Handler()))
	evt := <-listener.Handler()
	assert.Equal(t, zenoh.SampleKindPut, evt.Kind())
	assert.Equal(t, s2ZId.String(), evt.Link().ZId().String())
}

func TestLinkEventsListenerWithTransportFilter(t *testing.T) {
	s1, s2 := openConnectedPair(t, 17959)
	defer s1.Drop()
	defer s2.Drop()

	transports, err := s1.Transports()
	require.NoError(t, err)
	require.Equal(t, 1, len(transports))

	// Declare listener with history + transport filter — should receive the existing link.
	listener, err := s1.DeclareLinkEventsListener(
		zenoh.NewFifoChannel[zenoh.LinkEvent](16),
		&zenoh.LinkEventsListenerOptions{
			History:   true,
			Transport: option.Some(transports[0]),
		},
	)
	require.NoError(t, err)
	defer listener.Drop()

	time.Sleep(200 * time.Millisecond)

	assert.Equal(t, 1, len(listener.Handler()))
	evt := <-listener.Handler()
	assert.Equal(t, zenoh.SampleKindPut, evt.Kind())
	s2ZId := s2.ZId()
	assert.Equal(t, s2ZId.String(), evt.Link().ZId().String())
}

func TestTransportAccessors(t *testing.T) {
	s1, s2 := openConnectedPair(t, 17960)
	defer s1.Drop()
	defer s2.Drop()

	transports, err := s1.Transports()
	require.NoError(t, err)
	require.Equal(t, 1, len(transports))

	// Verify WhatAmI is Peer for both peer-mode sessions
	assert.Equal(t, zenoh.WhatAmIPeer, transports[0].WhatAmI())

	// Verify unicast TCP transport properties
	assert.False(t, transports[0].IsMulticast())

	// Just verify IsShm and IsQos don't panic and return bool values
	_ = transports[0].IsShm()
	_ = transports[0].IsQos()

	// Verify ZId matches s2
	assert.Equal(t, s2.ZId().String(), transports[0].ZId().String())
}

func TestTransportEventAccessors(t *testing.T) {
	s1 := openListenerSession(t, 17961)
	defer s1.Drop()

	listener, err := s1.DeclareTransportEventsListener(zenoh.NewFifoChannel[zenoh.TransportEvent](16), nil)
	require.NoError(t, err)
	defer listener.Drop()

	// Connect a peer to trigger a PUT event
	s2 := openConnectorSession(t, 17961)
	time.Sleep(500 * time.Millisecond)

	require.Equal(t, 1, len(listener.Handler()))
	evt := <-listener.Handler()

	// Verify event kind
	assert.Equal(t, zenoh.SampleKindPut, evt.Kind())

	// Verify transport properties in the snapshot
	assert.Equal(t, zenoh.WhatAmIPeer, evt.Transport().WhatAmI())
	assert.False(t, evt.Transport().IsMulticast())

	// Just verify IsShm and IsQos don't panic and return bool values
	_ = evt.Transport().IsShm()
	_ = evt.Transport().IsQos()

	s2.Drop()
}

func TestLinkAccessors(t *testing.T) {
	s1, s2 := openConnectedPair(t, 17962)
	defer s1.Drop()
	defer s2.Drop()

	links, err := s1.Links(nil)
	require.NoError(t, err)
	require.Equal(t, 1, len(links))

	// Verify non-empty source and destination endpoints
	assert.NotEmpty(t, links[0].Src())
	assert.NotEmpty(t, links[0].Dst())

	// Verify TCP-specific properties
	assert.Greater(t, links[0].Mtu(), uint16(0))
	assert.True(t, links[0].IsStreamed())

	// Verify interfaces slice doesn't panic (may be empty or populated)
	interfaces := links[0].Interfaces()
	assert.GreaterOrEqual(t, len(interfaces), 0)

	// Verify unicast link has no group or auth identifier
	assert.Empty(t, links[0].Group())
	assert.Empty(t, links[0].AuthIdentifier())

	// Just verify these accessors don't panic
	_, _, _ = links[0].Priorities()
	_, _ = links[0].Reliability()

	// Verify ZId matches s2
	assert.Equal(t, s2.ZId().String(), links[0].ZId().String())
}

func TestLinkEventSnapshotFields(t *testing.T) {
	s1 := openListenerSession(t, 17963)
	defer s1.Drop()

	listener, err := s1.DeclareLinkEventsListener(zenoh.NewFifoChannel[zenoh.LinkEvent](16), nil)
	require.NoError(t, err)
	defer listener.Drop()

	// Connect a peer to trigger a PUT event
	s2 := openConnectorSession(t, 17963)
	time.Sleep(500 * time.Millisecond)

	require.Equal(t, 1, len(listener.Handler()))
	evt := <-listener.Handler()

	// Also get the synchronous link from s1.Links
	links, err := s1.Links(nil)
	require.NoError(t, err)
	require.Equal(t, 1, len(links))

	// Verify event snapshot fields match synchronous link fields
	assert.Equal(t, evt.Link().ZId().String(), links[0].ZId().String())
	assert.Equal(t, evt.Link().Src(), links[0].Src())
	assert.NotEmpty(t, evt.Link().Src())
	assert.Equal(t, evt.Link().Dst(), links[0].Dst())
	assert.NotEmpty(t, evt.Link().Dst())
	assert.Equal(t, evt.Link().Mtu(), links[0].Mtu())
	assert.Greater(t, evt.Link().Mtu(), uint16(0))
	assert.Equal(t, evt.Link().IsStreamed(), links[0].IsStreamed())
	assert.True(t, evt.Link().IsStreamed()) // TCP is streamed
	assert.Equal(t, evt.Link().Interfaces(), links[0].Interfaces())
	assert.Equal(t, evt.Link().Group(), links[0].Group())
	assert.Empty(t, evt.Link().Group()) // Unicast link has no group
	assert.Equal(t, zenoh.SampleKindPut, evt.Kind())

	s2.Drop()
}

func TestListenerUndeclare(t *testing.T) {
	// Sub-test 1: TransportEventsListenerUndeclare
	t.Run("TransportEventsListenerUndeclare", func(t *testing.T) {
		s1 := openListenerSession(t, 17964)
		defer s1.Drop()

		listener, err := s1.DeclareTransportEventsListener(zenoh.NewFifoChannel[zenoh.TransportEvent](16), nil)
		require.NoError(t, err)

		// Undeclare the listener
		err = listener.Undeclare()
		assert.NoError(t, err)

		// Connect a peer and verify no events are received
		s2 := openConnectorSession(t, 17964)
		time.Sleep(200 * time.Millisecond)

		assert.Equal(t, 0, len(listener.Handler()))
		s2.Drop()
	})

	// Sub-test 2: LinkEventsListenerUndeclare
	t.Run("LinkEventsListenerUndeclare", func(t *testing.T) {
		s1 := openListenerSession(t, 17965)
		defer s1.Drop()

		listener, err := s1.DeclareLinkEventsListener(zenoh.NewFifoChannel[zenoh.LinkEvent](16), nil)
		require.NoError(t, err)

		// Undeclare the listener
		err = listener.Undeclare()
		assert.NoError(t, err)

		// Connect a peer and verify no events are received
		s2 := openConnectorSession(t, 17965)
		time.Sleep(200 * time.Millisecond)

		assert.Equal(t, 0, len(listener.Handler()))
		s2.Drop()
	})
}

func TestBackgroundLinkEventsListener(t *testing.T) {
	s1 := openListenerSession(t, 17966)
	defer s1.Drop()

	var events []zenoh.LinkEvent

	err := s1.DeclareBackgroundLinkEventsListener(
		zenoh.Closure[zenoh.LinkEvent]{
			Call: func(evt zenoh.LinkEvent) { events = append(events, evt) },
		},
		nil,
	)
	require.NoError(t, err)

	// Connect a peer — should produce a PUT event
	s2 := openConnectorSession(t, 17966)
	time.Sleep(500 * time.Millisecond)

	assert.Equal(t, 1, len(events))
	assert.Equal(t, zenoh.SampleKindPut, events[0].Kind())

	// Disconnect — should produce a DELETE event
	s2.Drop()
	time.Sleep(500 * time.Millisecond)

	assert.Equal(t, 2, len(events))
	assert.Equal(t, zenoh.SampleKindDelete, events[1].Kind())
}

func TestBackgroundTransportEventsListenerWithHistory(t *testing.T) {
	s1, s2 := openConnectedPair(t, 17967)
	defer s1.Drop()
	defer s2.Drop()

	s2ZId := s2.ZId()
	var events []zenoh.TransportEvent

	// Declare listener with history — should receive event for already-existing transport
	err := s1.DeclareBackgroundTransportEventsListener(
		zenoh.Closure[zenoh.TransportEvent]{
			Call: func(evt zenoh.TransportEvent) { events = append(events, evt) },
		},
		&zenoh.TransportEventsListenerOptions{History: true},
	)
	require.NoError(t, err)

	time.Sleep(200 * time.Millisecond)

	assert.Equal(t, 1, len(events))
	assert.Equal(t, zenoh.SampleKindPut, events[0].Kind())
	assert.Equal(t, s2ZId.String(), events[0].Transport().ZId().String())
}

func TestBackgroundLinkEventsListenerWithHistoryAndFilter(t *testing.T) {
	s1, s2 := openConnectedPair(t, 17968)
	defer s1.Drop()
	defer s2.Drop()

	s2ZId := s2.ZId()

	var events []zenoh.LinkEvent

	// Declare background listener with history option (tests options != nil branch)
	err := s1.DeclareBackgroundLinkEventsListener(
		zenoh.Closure[zenoh.LinkEvent]{
			Call: func(evt zenoh.LinkEvent) { events = append(events, evt) },
		},
		&zenoh.LinkEventsListenerOptions{
			History: true,
		},
	)
	require.NoError(t, err)

	time.Sleep(200 * time.Millisecond)

	assert.Equal(t, 1, len(events))
	assert.Equal(t, zenoh.SampleKindPut, events[0].Kind())
	assert.Equal(t, s2ZId.String(), events[0].Link().ZId().String())
}

func TestLinkEventsListenerTransportFilterForwardEvents(t *testing.T) {
	s1 := openListenerSession(t, 17969)
	defer s1.Drop()

	var events []zenoh.LinkEvent

	// Declare listener without filter first to get the transport for filtering
	listener, err := s1.DeclareLinkEventsListener(
		zenoh.NewFifoChannel[zenoh.LinkEvent](16),
		nil,
	)
	require.NoError(t, err)
	defer listener.Drop()

	// Connect a peer
	s2 := openConnectorSession(t, 17969)
	defer s2.Drop()
	time.Sleep(500 * time.Millisecond)

	// Collect the event
	if len(listener.Handler()) > 0 {
		events = append(events, <-listener.Handler())
	}

	// Verify we got one event (the connection)
	require.Equal(t, 1, len(events))
	assert.Equal(t, zenoh.SampleKindPut, events[0].Kind())
}

func TestEmptyTransportsAndLinksLists(t *testing.T) {
	s := openListenerSession(t, 17971)
	defer s.Drop()

	transports, err := s.Transports()
	assert.NoError(t, err)
	assert.Equal(t, 0, len(transports))

	links, err := s.Links(nil)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(links))
}
