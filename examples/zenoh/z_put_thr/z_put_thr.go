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
	"strconv"

	"github.com/eclipse-zenoh/zenoh-go"
)

func main() {
	var locator *string
	if len(os.Args) < 2 {
		fmt.Printf("USAGE:\n\ty_put_thr <payload-size> [<zenoh-locator>]\n\n")
		os.Exit(-1)
	}

	length, err := strconv.Atoi(os.Args[1])
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("Running throughput test for payload of %d bytes\n", length)
	if len(os.Args) > 2 {
		locator = &os.Args[2]
	}

	path := "/zenoh/examples/throughput/data"

	data := make([]byte, length)
	for i := 0; i < length; i++ {
		data[i] = byte(i % 10)
	}

	p, err := zenoh.NewPath(path)
	if err != nil {
		panic(err.Error())
	}
	v := zenoh.NewRawValue(data)

	fmt.Println("Login to Zenoh...")
	y, err := zenoh.Login(locator, nil)
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
