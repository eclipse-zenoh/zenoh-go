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

var path string

func queryHandler(rname string, predicate string, repliesSender *znet.RepliesSender) {
	fmt.Printf(">> [Query handler] Handling '%s?%s'\n", rname, predicate)

	replies := make([]znet.Resource, 1, 1)
	replies[0].RName = path
	replies[0].Data = []byte("Eval from Go!")
	replies[0].Encoding = 0
	replies[0].Kind = 0

	repliesSender.SendReplies(replies)
}

func main() {

	// --- Command line argument parsing --- --- --- --- --- ---
	var args struct {
		Path    string `default:"/zenoh/examples/go/eval" arg:"-p" help:"the path representing the URI"`
		Locator string `arg:"-l" help:"The locator to be used to boostrap the zenoh session. By default dynamic discovery is used"`
	}
	arg.MustParse(&args)

	// zenoh-net code  --- --- --- --- --- --- --- --- --- --- ---
	fmt.Println("Opening session...")
	s, err := znet.Open(&args.Locator, nil)
	if err != nil {
		panic(err.Error())
	}
	defer s.Close()

	fmt.Println("Declaring Eval on '" + args.Path + "'...")
	e, err := s.DeclareEval(args.Path, queryHandler)
	if err != nil {
		panic(err.Error())
	}
	defer s.UndeclareEval(e)

	reader := bufio.NewReader(os.Stdin)
	var c rune
	for c != 'q' {
		c, _, _ = reader.ReadRune()
	}

}
