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

	"github.com/eclipse-zenoh/zenoh-go"
)

func main() {
	path := "/zenoh/examples/go/put/hello"
	if len(os.Args) > 1 {
		path = os.Args[1]
	}

	value := "Zenitude put from zenoh-go!"
	if len(os.Args) > 2 {
		value = os.Args[2]
	}

	var locator *string
	if len(os.Args) > 3 {
		locator = &os.Args[3]
	}

	p, err := zenoh.NewPath(path)
	if err != nil {
		panic(err.Error())
	}
	v := zenoh.NewStringValue(value)

	fmt.Println("Login to Zenoh...")
	y, err := zenoh.Login(locator, nil)
	if err != nil {
		panic(err.Error())
	}

	root, _ := zenoh.NewPath("/")
	w := y.Workspace(root)

	fmt.Println("Put on " + p.ToString() + " : " + v.ToString())
	err = w.Put(p, v)
	if err != nil {
		panic(err.Error())
	}

	err = y.Logout()
	if err != nil {
		panic(err.Error())
	}

}
