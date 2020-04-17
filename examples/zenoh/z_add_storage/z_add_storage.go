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
		Selector string `default:"/zenoh/examples/**" arg:"-s" help:"the selector associated with this storage"`
		ID       string `default:"zenoh-examples-storage" arg:"-i" help:"the storage identifier"`
		Locator  string `arg:"-l" help:"The locator to be used to boostrap the zenoh session. By default dynamic discovery is used"`
	}
	arg.MustParse(&args)

	// zenoh-net code  --- --- --- --- --- --- --- --- --- --- ---
	fmt.Println("Login to Zenoh...")
	y, err := zenoh.Login(&args.Locator, nil)
	if err != nil {
		panic(err.Error())
	}

	admin := y.Admin()

	fmt.Println("Add storage " + args.ID + " with selector " + args.Selector)
	p := make(map[string]string)
	p["selector"] = args.Selector
	admin.AddStorage(args.ID, p)

	err = y.Logout()
	if err != nil {
		panic(err.Error())
	}

}
