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
	selector := "/zenoh/examples/**"
	if len(os.Args) > 1 {
		selector = os.Args[1]
	}

	storageID := "Demo"
	if len(os.Args) > 2 {
		storageID = os.Args[2]
	}

	var locator *string
	if len(os.Args) > 3 {
		locator = &os.Args[3]
	}

	fmt.Println("Login to Zenoh...")
	y, err := zenoh.Login(locator, nil)
	if err != nil {
		panic(err.Error())
	}

	admin := y.Admin()

	fmt.Println("Add storage " + storageID + " with selector " + selector)
	p := make(map[string]string)
	p["selector"] = selector
	admin.AddStorage(storageID, p)

	err = y.Logout()
	if err != nil {
		panic(err.Error())
	}

}
