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
	"os"
	"time"

	znet "github.com/eclipse-zenoh/zenoh-go/net"
)

func main() {
	path := "/zenoh/examples/go/stream/hello"
	if len(os.Args) > 1 {
		path = os.Args[1]
	}

	value := "Zenitude streamed from zenoh-net-go!"
	if len(os.Args) > 2 {
		value = os.Args[2]
	}

	var locator *string
	if len(os.Args) > 3 {
		locator = &os.Args[3]
	}

	fmt.Println("Opening session...")
	s, err := znet.Open(locator, nil)
	if err != nil {
		panic(err.Error())
	}
	defer s.Close()

	fmt.Println("Declaring Publisher on '" + path + "'...")
	pub, err := s.DeclarePublisher(path)
	if err != nil {
		panic(err.Error())
	}
	defer s.UndeclarePublisher(pub)

	for idx := 0; idx < 100; idx++ {
		time.Sleep(1 * time.Second)
		str := fmt.Sprintf("[%4d] %s", idx, value)
		fmt.Printf("Streaming Data ('%s': '%s')...\n", path, str)
		pub.StreamData([]byte(str))
	}
}
