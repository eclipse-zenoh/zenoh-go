//
// Copyright (c) 2025 ZettaScale Technology
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

package main

import (
	"fmt"
	"github.com/eclipse-zenoh/zenoh-go/examples/utils"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/eclipse-zenoh/zenoh-go/zenoh"
)

func main() {
	zenoh.InitLoggerFromEnvOr("error")
	args := parseArgs()

	fmt.Println("Opening session...")
	session, err := zenoh.Open(args.config, nil)
	if err != nil {
		fmt.Printf("Failed to open Zenoh session: %v\n", err)
		os.Exit(-1)
	}
	defer session.Drop()

	fmt.Printf("own id: %s\n", session.ZId())

	fmt.Println("routers ids:")
	ids, _ := session.RoutersZId()
	for _, id := range ids {
		fmt.Printf("%s\n", id)
	}

	fmt.Printf("peers ids:\n")
	ids, _ = session.PeersZId()
	for _, id := range ids {
		fmt.Printf("%s\n", id)
	}

	// Warning: transport/link APIs are unstable.

	fmt.Println("transports:")
	transports, _ := session.Transports()
	for _, tr := range transports {
		fmt.Printf("  zid: %s, whatami: %s, qos: %v, multicast: %v\n",
			tr.ZId(), tr.WhatAmI(), tr.IsQos(), tr.IsMulticast())
	}

	fmt.Println("links:")
	links, _ := session.Links(nil)
	for _, l := range links {
		ifaces := strings.Join(l.Interfaces(), ", ")
		fmt.Printf("  zid: %s, src: %s, dst: %s, mtu: %d, streamed: %v, interfaces: [%s]\n",
			l.ZId(), l.Src(), l.Dst(), l.Mtu(), l.IsStreamed(), ifaces)
	}

	_ = session.DeclareBackgroundTransportEventsListener(
		zenoh.Closure[zenoh.TransportEvent]{
			Call: func(evt zenoh.TransportEvent) {
				action := "Opened"
				if evt.Kind() == zenoh.SampleKindDelete {
					action = "Closed"
				}
				fmt.Printf("[Transport Event] %s: zid=%s\n", action, evt.Transport().ZId())
			},
		},
		nil,
	)

	_ = session.DeclareBackgroundLinkEventsListener(
		zenoh.Closure[zenoh.LinkEvent]{
			Call: func(evt zenoh.LinkEvent) {
				action := "Added"
				if evt.Kind() == zenoh.SampleKindDelete {
					action = "Removed"
				}
				fmt.Printf("[Link Event] %s: zid=%s, src=%s, dst=%s\n",
					action, evt.Link().ZId(), evt.Link().Src(), evt.Link().Dst())
			},
		},
		nil,
	)

	fmt.Println("Monitoring connectivity events. Press CTRL-C to quit...")
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	fmt.Println("Exiting...")
}

type Args struct {
	config zenoh.Config
}

func parseArgs() Args {
	return Args{config: utils.ParseConfig()}
}
