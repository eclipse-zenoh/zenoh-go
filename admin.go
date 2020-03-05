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

package zenoh

import (
	"fmt"
	"strings"
)

// Admin is the zenoh administration class.
type Admin struct {
	w       *Workspace
	zenohid string
}

//
// Backends management
//

// AddBackend adds a backend in the connected Zenoh router
// (i.e. the one you are directly connected to).
//
// The backend will use the properties for initialization and "beid" as identifier.
func (a *Admin) AddBackend(beid string, properties Properties) error {
	return a.AddBackendAt(beid, properties, a.zenohid)
}

// AddBackendAt adds a backend in the specified Zenoh router,
// not necessarily the one you are connected to.
//
// The backend will use the properties for initialization and "beid" as identifier.
func (a *Admin) AddBackendAt(beid string, properties Properties, zenoh string) error {
	path, err := NewPath(fmt.Sprintf("/@/router/%s/plugin/storages/backend/%s", zenoh, beid))
	if err != nil {
		return &ZError{"Invalid backend id: " + beid, 0, err}
	}
	return a.w.Put(path, NewPropertiesValue(properties))
}

// GetBackend gets a backend's properties from the connected Zenoh router.
// (i.e. the one you are directly connected to).
func (a *Admin) GetBackend(beid string) (Properties, error) {
	return a.GetBackendAt(beid, a.zenohid)
}

// GetBackendAt gets a backend's properties from the specified Zenoh router,
// not necessarily the one you are connected to.
func (a *Admin) GetBackendAt(beid string, zenoh string) (Properties, error) {
	selector, err := NewSelector(fmt.Sprintf("/@/router/%s/plugin/storages/backend/%s", zenoh, beid))
	if err != nil {
		return nil, &ZError{"Invalid backend id: " + beid, 0, err}
	}
	pvs := a.w.Get(selector)
	if len(pvs) == 0 {
		return nil, nil
	}
	return propertiesOfValue(pvs[0].Value()), nil
}

// GetBackends gets all the backends from the connected Zenoh router
// (i.e. the one you are directly connected to).
//
// It returns a map of the backends properties, indexed by the backends identifiers.
func (a *Admin) GetBackends() (map[string]Properties, error) {
	return a.GetBackendsAt(a.zenohid)
}

// GetBackendsAt gets all the backends from the specified Zenoh router,
// not necessarily the one you are connected to.
//
// It returns a map of the backends properties, indexed by the backends identifiers.
func (a *Admin) GetBackendsAt(zenoh string) (map[string]Properties, error) {
	sel := fmt.Sprintf("/@/router/%s/plugin/storages/backend/*", zenoh)
	selector, _ := NewSelector(sel)
	pvs := a.w.Get(selector)
	result := make(map[string]Properties)
	for _, pv := range pvs {
		beid := pv.Path().ToString()[len(sel)-1:]
		result[beid] = propertiesOfValue(pv.Value())
	}
	return result, nil
}

// RemoveBackend removes a backend from the connected Zenoh router
// (i.e. the one you are directly connected to).
func (a *Admin) RemoveBackend(beid string) error {
	return a.RemoveBackendAt(beid, a.zenohid)
}

// RemoveBackendAt removes a backend from the specified Zenoh router,
// not necessarily the one you are connected to.
func (a *Admin) RemoveBackendAt(beid string, zenoh string) error {
	path, err := NewPath(fmt.Sprintf("/@/router/%s/plugin/storages/backend/%s", zenoh, beid))
	if err != nil {
		return &ZError{"Invalid backend id: " + beid, 0, err}
	}
	return a.w.Remove(path)
}

//
// Storages management
//

// AddStorage adds a storage in the connected Zenoh router, using an automatically chosen backend.
//
// The storage will use the properties for initialization and "stid" as identifier.
func (a *Admin) AddStorage(stid string, properties Properties) error {
	return a.AddStorageOnBackendAt(stid, properties, "auto", a.zenohid)
}

// AddStorageAt adds a storage in the specified Zenoh router, using an automatically chosen backend.
//
// The storage will use the properties for initialization and "stid" as identifier.
func (a *Admin) AddStorageAt(stid string, properties Properties, zenoh string) error {
	return a.AddStorageOnBackendAt(stid, properties, "auto", zenoh)
}

// AddStorageOnBackend adds a storage in the connected Zenoh router, using the specified backend.
func (a *Admin) AddStorageOnBackend(stid string, properties Properties, backend string) error {
	return a.AddStorageOnBackendAt(stid, properties, backend, a.zenohid)
}

// AddStorageOnBackendAt adds a storage in the specified Zenoh router, using the specified backend.
func (a *Admin) AddStorageOnBackendAt(stid string, properties Properties, backend string, zenoh string) error {
	path, err := NewPath(fmt.Sprintf("/@/router/%s/plugin/storages/backend/%s/storage/%s", zenoh, backend, stid))
	if err != nil {
		return &ZError{"Invalid backend or storage id in path: " + path.ToString(), 0, err}
	}
	return a.w.Put(path, NewPropertiesValue(properties))
}

// GetStorage gets a storage's properties from the connected Zenoh router.
func (a *Admin) GetStorage(stid string) (Properties, error) {
	return a.GetStorageAt(stid, a.zenohid)
}

// GetStorageAt gets a storage's properties from the specified Zenoh router.
func (a *Admin) GetStorageAt(stid string, zenoh string) (Properties, error) {
	selector, err := NewSelector(fmt.Sprintf("/@/router/%s/plugin/storages/backend/*/storage/%s", zenoh, stid))
	if err != nil {
		return nil, &ZError{"Invalid storage id: " + stid, 0, err}
	}
	pvs := a.w.Get(selector)
	if len(pvs) == 0 {
		return nil, nil
	}
	return propertiesOfValue(pvs[0].Value()), nil
}

// GetStorages gets all the storages from the connected Zenoh router.
//
// It returns a map of the stgorages properties, indexed by the storages identifiers.
func (a *Admin) GetStorages() (map[string]Properties, error) {
	return a.GetStoragesFromBackendAt("*", a.zenohid)
}

// GetStoragesAt gets all the storages from the specified Zenoh router.
//
// It returns a map of the stgorages properties, indexed by the storages identifiers.
func (a *Admin) GetStoragesAt(zenoh string) (map[string]Properties, error) {
	return a.GetStoragesFromBackendAt("*", a.zenohid)
}

// GetStoragesFromBackend gets all the storages from the specified backend within the connected Zenoh router.
//
// It returns a map of the stgorages properties, indexed by the storages identifiers.
func (a *Admin) GetStoragesFromBackend(backend string) (map[string]Properties, error) {
	return a.GetStoragesFromBackendAt(backend, a.zenohid)
}

// GetStoragesFromBackendAt gets all the storages from the specified backend within the specified Zenoh router.
//
// It returns a map of the stgorages properties, indexed by the storages identifiers.
func (a *Admin) GetStoragesFromBackendAt(backend string, zenoh string) (map[string]Properties, error) {
	sel := fmt.Sprintf("/@/router/%s/plugin/storages/backend/%s/storage/*", zenoh, backend)
	selector, err := NewSelector(sel)
	if err != nil {
		return nil, &ZError{"Invalid backend id: " + backend, 0, err}
	}
	pvs := a.w.Get(selector)
	result := make(map[string]Properties)
	for _, pv := range pvs {
		stPath := pv.Path().ToString()
		stid := pv.Path().ToString()[strings.LastIndex(stPath, "/")+1:]
		result[stid] = propertiesOfValue(pv.Value())
	}
	return result, nil
}

// RemoveStorage removes a storage from the connected Zenoh router.
func (a *Admin) RemoveStorage(stid string) error {
	return a.RemoveStorageAt(stid, a.zenohid)
}

// RemoveStorageAt removes a storage from the specified Zenoh router.
func (a *Admin) RemoveStorageAt(stid string, zenoh string) error {
	selector, err := NewSelector(fmt.Sprintf("/@/router/%s/plugin/storages/backend/*/storage/%s", zenoh, stid))
	if err != nil {
		return &ZError{"Invalid storage id: " + stid, 0, err}
	}
	pvs := a.w.Get(selector)
	for _, pv := range pvs {
		p := pv.Path()
		err := a.w.Remove(p)
		if err != nil {
			return err
		}
	}
	return nil
}

func propertiesOfValue(v Value) Properties {
	pVal, ok := v.(*PropertiesValue)
	if ok {
		return pVal.p
	}
	p := make(Properties)
	p["value"] = v.ToString()
	return p
}
