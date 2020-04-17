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
	"bufio"
	"fmt"
	"os"

	"github.com/alexflint/go-arg"
	znet "github.com/eclipse-zenoh/zenoh-go/net"
)

func listener(rname string, data []byte, info *znet.DataInfo) {
	str := string(data)
	fmt.Printf(">> [Subscription listener] Received ('%s': '%s')\n", rname, str)
}

func main() {
	// --- Command line argument parsing --- --- --- --- --- ---
	var args struct {
		Selector string `default:"/zenoh/examples/**" arg:"-s" help:"The selector specifying the subscription"`
		Locator  string `arg:"-l" help:"The locator to be used to boostrap the zenoh session. By default dynamic discovery is used"`
	}
	arg.MustParse(&args)

	// zenoh-net code  --- --- --- --- --- --- --- --- --- --- ---
	fmt.Println("Opening session...")
	s, err := znet.Open(&args.Locator, nil)
	if err != nil {
		panic(err.Error())
	}
	defer s.Close()

	fmt.Println("Declaring Subscriber on '" + args.Selector + "'...")
	sub, err := s.DeclareSubscriber(args.Selector, znet.NewSubMode(znet.ZNPushMode), listener)
	if err != nil {
		panic(err.Error())
	}
	defer s.UndeclareSubscriber(sub)

	reader := bufio.NewReader(os.Stdin)
	var c rune
	for c != 'q' {
		c, _, _ = reader.ReadRune()
	}
}
