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
	"time"

	"github.com/alexflint/go-arg"
	znet "github.com/eclipse-zenoh/zenoh-go/net"
)

func replyHandler(reply *znet.ReplyValue) {
	switch reply.Kind() {
	case znet.ZNStorageData, znet.ZNEvalData:
		str := string(reply.Data())
		switch reply.Kind() {
		case znet.ZNStorageData:
			fmt.Printf(">> [Reply handler] received -Storage Data- ('%s': '%s')\n", reply.RName(), str)
		case znet.ZNEvalData:
			fmt.Printf(">> [Reply handler] received -Eval Data-    ('%s': '%s')\n", reply.RName(), str)
		}

	case znet.ZNStorageFinal:
		fmt.Println(">> [Reply handler] received -Storage Final-")

	case znet.ZNEvalFinal:
		fmt.Println(">> [Reply handler] received -Eval Final-")

	case znet.ZNReplyFinal:
		fmt.Println(">> [Reply handler] received -Reply Final-")
	}
}

func main() {
	// --- Command line argument parsing --- --- --- --- --- ---
	var args struct {
		Selector string `default:"/zenoh/examples/**" arg:"-s" help:"The selector to be used for issuing the query"`
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

	fmt.Println("Sending Query '" + args.Selector + "'...")
	err = s.QueryWO(args.Selector, "", replyHandler, znet.NewQueryDest(znet.ZNAll), znet.NewQueryDest(znet.ZNAll))
	if err != nil {
		panic(err.Error())
	}

	time.Sleep(1 * time.Second)
}
