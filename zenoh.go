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

// Package zenoh provides the Zenoh client API in Go.
package zenoh

import (
	"encoding/hex"

	znet "github.com/eclipse-zenoh/zenoh-go/net"
	log "github.com/sirupsen/logrus"
)

// PropUser is the "user" property key
const PropUser = "user"

// PropPassword is the "password" property key
const PropPassword = "password"

// Zenoh is the Zenoh client API
type Zenoh struct {
	session *znet.Session
	zenohid string
	admin   *Admin
}

var logger = log.WithFields(log.Fields{" pkg": "zenoh"})

func newZenoh(s *znet.Session) (*Zenoh, error) {
	props := s.Info()
	pid, ok := props[znet.InfoPeerPidKey]
	if !ok {
		return nil, &ZError{"Failed to retrieve Zenoh id from Session info", 0, nil}
	}
	zenohid := hex.EncodeToString(pid)
	adminPath, _ := NewPath("/@")
	adminWS := &Workspace{adminPath, s, make(map[Path]*znet.Eval), false}
	return &Zenoh{s, zenohid, &Admin{adminWS, zenohid}}, nil
}

func getZProps(properties Properties) map[int][]byte {
	zprops := make(map[int][]byte)
	user, ok := properties[PropUser]
	if ok {
		zprops[znet.UserKey] = []byte(user)
	}
	password, ok := properties[PropPassword]
	if ok {
		zprops[znet.PasswdKey] = []byte(password)
	}
	return zprops
}

// Login establishes a zenoh session via a provided locator.
//
// Locator is a string representing the network endpoint to which establish the session.
// If the provided locator is nil, login will perform some dynamic discovery and
// try to establish the session automatically. When not nil, the locator
// must have the format: "tcp/<ip>:<port>" (for instance: "tcp/127.0.0.1:7447").
//
// Properties contains the configuration to be used for this session
// (e.g. "user", "password"...). It can be nil.
func Login(locator *string, properties Properties) (*Zenoh, error) {
	logger.WithField("locator", locator).Debug("Establishing session to Zenoh router")
	z, e := znet.Open(locator, getZProps(properties))
	if e != nil {
		return nil, &ZError{"Login failed", 0, e}
	}
	return newZenoh(z)
}

// Logout terminates the Zenoh session.
func (z *Zenoh) Logout() error {
	if e := z.session.Close(); e != nil {
		return &ZError{"Error during logout", 0, e}
	}
	return nil
}

// Workspace creates a Workspace using the provided path.
// All relative Selector or Path used with this Workspace will be relative to this path.
//
// Notice that all subscription listeners and eval callbacks declared in this workspace will be
// executed by the I/O subroutine. This implies that no long operations or other call to Zenoh
// shall be performed in those callbacks.
func (z *Zenoh) Workspace(path *Path) *Workspace {
	return &Workspace{path, z.session, make(map[Path]*znet.Eval), false}
}

// WorkspaceWithExecutor creates a Workspace using the provided path.
// All relative Selector or Path used with this Workspace will be relative to this path.
//
// Notice that all subscription listeners and eval callbacks declared in this workspace will be
// executed by their own subroutine. This is useful when listeners and/or callbacks need to perform
// long operations or need to call other Zenoh operations.
func (z *Zenoh) WorkspaceWithExecutor(path *Path) *Workspace {
	return &Workspace{path, z.session, make(map[Path]*znet.Eval), true}
}

// Admin returns the admin object that provides
// helper operations to administer Zenoh.
func (z *Zenoh) Admin() *Admin {
	return z.admin
}
