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
	"strconv"

	"github.com/alexflint/go-arg"
	"github.com/eclipse-zenoh/zenoh-go"
)

func main() {
	// --- Command line argument parsing --- --- --- --- --- ---
	var args struct {
		Path    string `default:"/zenoh/examples/native/float" arg:"-p" help:"the path representing the float resource"`
		Locator string `arg:"-l" help:"The locator to be used to boostrap the zenoh session. By default dynamic discovery is used"`
	}
	arg.MustParse(&args)

	// zenoh-net code  --- --- --- --- --- --- --- --- --- --- ---
	p, err := zenoh.NewPath(args.Path)
	if err != nil {
		panic(err.Error())
	}

	y, err := zenoh.Login(&args.Locator, nil)
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
