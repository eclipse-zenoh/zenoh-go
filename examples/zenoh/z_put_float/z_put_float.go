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
	// If not specified as 1st argument, use a relative path (to the workspace below): "zenoh-go-put"
	path := "/zenoh/examples/native/float"
	if len(os.Args) > 1 {
		path = os.Args[1]
	}

	var locator *string
	if len(os.Args) > 3 {
		locator = &os.Args[3]
	}

	p, err := zenoh.NewPath(path)
	if err != nil {
		panic(err.Error())
	}

	y, err := zenoh.Login(locator, nil)
	if err != nil {
		panic(err.Error())
	}

	root, _ := zenoh.NewPath("/")
	w := y.Workspace(root)

	v := ""
	for v != "." {
		fmt.Print("Insert value ('.' to exit): ")
		fmt.Scanf("%s", &v)
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			err = w.PutFloat(p, f)
			if err != nil {
				panic(err.Error())
			}
		} else {
			fmt.Println("Invalid float!")
		}
	}

	err = y.Logout()
	if err != nil {
		panic(err.Error())
	}

}
