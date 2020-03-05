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
	path := "/demo/example/zenoh-go-eval"
	if len(os.Args) > 1 {
		path = os.Args[1]
	}

	var locator *string
	if len(os.Args) > 2 {
		locator = &os.Args[2]
	}

	p, err := zenoh.NewPath(path)
	if err != nil {
		panic(err.Error())
	}

	fmt.Println("Login to Zenoh...")
	y, err := zenoh.Login(locator, nil)
	if err != nil {
		panic(err.Error())
	}

	fmt.Println("Use Workspace on '/'")
	root, _ := zenoh.NewPath("/")
	w := y.WorkspaceWithExecutor(root)

	fmt.Println("Register eval " + p.ToString())
	err = w.RegisterEval(p,
		func(path *zenoh.Path, props zenoh.Properties) zenoh.Value {
			// In this Eval function, we choosed to get the name to be returned in the StringValue in 3 possible ways,
			// depending the properties specified in the selector. For example, with the following selectors:
			//   - "/demo/example/zenoh-go-eval" : no properties are set, a default value is used for the name
			//   - "/demo/example/zenoh-go-eval?(name=Bob)" : "Bob" is used for the name
			//   - "/demo/example/zenoh-go-eval?(name=/demo/example/name)" :
			//     the Eval function does a GET on "/demo/example/name" an uses the 1st result for the name

			fmt.Printf(">> Processing eval for path %s with properties: %s\n", path, props)
			name := props["name"]
			if name == "" {
				name = "Zenoh Go!"
			}

			if name[0] == '/' {
				fmt.Printf("   >> Get name to use from Zenoh at path: %s\n", name)
				s, err := zenoh.NewSelector(name)
				if err == nil {
					kvs := w.Get(s)
					if len(kvs) > 0 {
						name = kvs[0].Value().ToString()
					}
				}
			}
			fmt.Printf("   >> Returning string: \"Eval from %s\"\n", name)
			return zenoh.NewStringValue("Eval from " + name)
		})
	if err != nil {
		panic(err.Error())
	}

	fmt.Println("Enter 'q' to quit...")
	fmt.Println()
	var b = make([]byte, 1)
	for b[0] != 'q' {
		os.Stdin.Read(b)
	}

	w.UnregisterEval(p)
	if err != nil {
		panic(err.Error())
	}

	err = y.Logout()
	if err != nil {
		panic(err.Error())
	}

}
