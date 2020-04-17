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
	"encoding/hex"
	"fmt"

	"github.com/alexflint/go-arg"
	znet "github.com/eclipse-zenoh/zenoh-go/net"
)

func main() {
	// --- Command line argument parsing --- --- --- --- --- ---
	var args struct {
		Locator string `arg:"-l" help:"The locator to be used to boostrap the zenoh session. By default dynamic discovery is used"`
	}
	arg.MustParse(&args)

	// zenoh-net code  --- --- --- --- --- --- --- --- --- --- ---
	fmt.Println("Opening session...")
	properties := map[int][]byte{
		znet.UserKey:   []byte("user"),
		znet.PasswdKey: []byte("password")}
	s, err := znet.Open(&args.Locator, properties)
	if err != nil {
		panic(err.Error())
	}
	defer s.Close()

	info := s.Info()
	fmt.Println("LOCATOR :  " + string(info[znet.InfoPeerKey]))
	fmt.Println("PID :      " + hex.EncodeToString(info[znet.InfoPidKey]))
	fmt.Println("PEER PID : " + hex.EncodeToString(info[znet.InfoPeerPidKey]))
}
