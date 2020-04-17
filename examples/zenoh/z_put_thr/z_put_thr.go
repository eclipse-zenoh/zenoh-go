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
	"github.com/eclipse-zenoh/zenoh-go"
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
	data := make([]byte, args.Size)
	for i := 0; i < args.Size; i++ {
		data[i] = byte(i % 10)
	}

	p, err := zenoh.NewPath(args.Path)
	if err != nil {
		panic(err.Error())
	}
	v := zenoh.NewRawValue(data)

	fmt.Println("Login to Zenoh...")
	y, err := zenoh.Login(&args.Locator, nil)
	if err != nil {
		panic(err.Error())
	}

	fmt.Println("Use Workspace on '/'")
	root, _ := zenoh.NewPath("/")
	w := y.Workspace(root)

	fmt.Printf("Put on %s : %db\n", p.ToString(), len(data))

	for {
		err = w.Put(p, v)
		if err != nil {
			panic(err.Error())
		}
	}

}
