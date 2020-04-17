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

var stored map[string][]byte

func listener(rname string, data []byte, info *znet.DataInfo) {
	str := string(data)
	fmt.Printf(">> [Storage listener] Received ('%20s' : '%s')\n", rname, str)
	stored[rname] = data
}

func queryHandler(rname string, predicate string, repliesSender *znet.RepliesSender) {
	fmt.Printf(">> [Query handler   ] Handling '%s?%s'\n", rname, predicate)
	replies := make([]znet.Resource, 0, len(stored))
	for k, v := range stored {
		if znet.RNameIntersect(rname, k) {
			var res znet.Resource
			res.RName = k
			res.Data = v
			res.Encoding = 0
			res.Kind = 0
			replies = append(replies, res)
		}
	}

	repliesSender.SendReplies(replies)
}

func main() {
	// --- Command line argument parsing --- --- --- --- --- ---
	var args struct {
		Selector string `default:"/zenoh/examples/**" arg:"-s" help:"the selector associated with this storage"`
		Locator  string `arg:"-l" help:"The locator to be used to boostrap the zenoh session. By default dynamic discovery is used"`
	}
	arg.MustParse(&args)

	// zenoh-net code  --- --- --- --- --- --- --- --- --- --- ---
	stored = make(map[string][]byte)

	fmt.Println("Opening session...")
	s, err := znet.Open(&args.Locator, nil)
	if err != nil {
		panic(err.Error())
	}
	defer s.Close()

	fmt.Println("Declaring Storage on '" + args.Selector + "'...")
	sto, err := s.DeclareStorage(args.Selector, listener, queryHandler)
	if err != nil {
		panic(err.Error())
	}
	defer s.UndeclareStorage(sto)

	reader := bufio.NewReader(os.Stdin)
	var c rune
	for c != 'q' {
		c, _, _ = reader.ReadRune()
	}
}
