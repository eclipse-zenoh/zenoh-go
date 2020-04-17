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

package main

import (
	"fmt"

	"github.com/alexflint/go-arg"
	znet "github.com/eclipse-zenoh/zenoh-go/net"
)

func main() {
	// --- Command line argument parsing --- --- --- --- --- ---
	var args struct {
		Size    int    `default:"256" arg:"-s"help:"the size in bytes of the payload used for the throughput test"`
		Locator string `arg:"-l" help:"The locator to be used to boostrap the zenoh session. By default dynamic discovery is used"`
		Path    string `default:"/zenoh/examples/throughput/data" arg:"-p" help:"the resource used to write throughput data"`
	}
	arg.MustParse(&args)

	// zenoh-net code  --- --- --- --- --- --- --- --- --- --- ---
	fmt.Printf("Running throughput test for payload of %d bytes\n", args.Size)
	data := make([]byte, args.Size)
	for i := 0; i < args.Size; i++ {
		data[i] = byte(i % 10)
	}

	s, err := znet.Open(&args.Locator, nil)
	if err != nil {
		panic(err.Error())
	}
	defer s.Close()

	pub, err := s.DeclarePublisher(args.Path)
	if err != nil {
		panic(err.Error())
	}
	defer s.UndeclarePublisher(pub)

	for true {
		pub.StreamData(data)
	}

}
