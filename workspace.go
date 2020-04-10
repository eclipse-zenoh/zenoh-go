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
	"sort"
	"strings"
	"sync"

	zcore "github.com/eclipse-zenoh/zenoh-go/core"
	znet "github.com/eclipse-zenoh/zenoh-go/net"
	log "github.com/sirupsen/logrus"
)

// Workspace allows to operate on Zenoh.
type Workspace struct {
	path          *Path
	session       *znet.Session
	evals         map[Path]*znet.Eval
	useSubroutine bool
}

// Put a path/value into Zenoh.
func (w *Workspace) Put(path *Path, value Value) error {
	logger.WithFields(log.Fields{
		"path":  path,
		"value": value,
	}).Debug("Put")
	p := w.toAbsolutePath(path)
	if e := w.session.WriteDataWO(p.ToString(), value.Encode(), value.Encoding(), PUT); e != nil {
		return &ZError{Msg: "Put on " + p.ToString() + " failed", Code: 0, Cause: e}
	}
	return nil
}

// Update a path/value into Zenoh.
func (w *Workspace) Update(path *Path, value Value) error {
	logger.WithFields(log.Fields{
		"path":  path,
		"value": value,
	}).Debug("Update")
	p := w.toAbsolutePath(path)
	if e := w.session.WriteDataWO(p.ToString(), value.Encode(), value.Encoding(), UPDATE); e != nil {
		return &ZError{Msg: "Put on " + path.ToString() + " failed", Code: 0, Cause: e}
	}
	return nil
}

// Remove a path/value from Zenoh.
func (w *Workspace) Remove(path *Path) error {
	logger.WithField("path", path).Debug("Remove")
	p := w.toAbsolutePath(path)
	if e := w.session.WriteDataWO(p.ToString(), nil, 0, REMOVE); e != nil {
		return &ZError{Msg: "Put on " + path.ToString() + " failed", Code: 0, Cause: e}
	}
	return nil
}

// dataset: a list of Data that can be sorted per Timestamp
type dataset []Data

func (ds dataset) Len() int {
	return len(ds)
}

func (ds dataset) Less(i, j int) bool {
	return ds[i].Timestamp().Before(ds[j].Timestamp())
}

func (ds dataset) Swap(i, j int) {
	ds[i], ds[j] = ds[j], ds[i]
}

// asSortedSet returns a sorted copy of the dataset, removing duplicate (i.e. with same timestamp)
func (ds dataset) asSortedSet() dataset {
	// sort
	sort.Sort(ds)
	// remove duplicates
	res := make([]Data, 0)
	var ts *Timestamp
	for _, data := range ds {
		if ts == nil || *data.Timestamp() != *ts {
			res = append(res, data)
			ts = data.Timestamp()
		}
	}
	return res
}

// isSelectorForSeries returns true if the selector implies time series within reply
func isSelectorForSeries(selector *Selector) bool {
	// search for starttime or stoptime property in selector
	props := strings.Split(selector.Properties(), ";")
	for _, p := range props {
		if strings.HasPrefix(p, "starttime") || strings.HasPrefix(p, "stoptime") {
			return true
		}
	}
	return false
}

// Get a selection of path/value from Zenoh.
func (w *Workspace) Get(selector *Selector) []Data {
	s := w.toAbsoluteSelector(selector)
	logger := logger.WithField("selector", s)
	logger.Debug("Get")

	qresults := make(map[Path]dataset)
	queryFinished := false

	mu := new(sync.Mutex)
	cond := sync.NewCond(mu)

	replyCb := func(reply *znet.ReplyValue) {
		switch reply.Kind() {
		case znet.ZNStorageData, znet.ZNEvalData:
			path, err := NewPath(reply.RName())
			if err != nil {
				logger.WithField("reply path", reply.RName()).
					Warn("Get received reply for an invalid path")
				return
			}
			data := reply.Data()
			info := reply.Info()
			encoding := info.Encoding()
			if reply.Kind() == znet.ZNStorageData {
				logger.WithFields(log.Fields{
					"reply path": reply.RName(),
					"len(data)":  len(data),
					"encoding":   encoding,
				}).Trace("Get => ZN_STORAGE_DATA")
			} else {
				logger.WithFields(log.Fields{
					"reply path": reply.RName(),
					"len(data)":  len(data),
					"encoding":   encoding,
				}).Trace("Get => ZN_EVAL_DATA")
			}

			decoder, ok := valueDecoders[encoding]
			if !ok {
				logger.WithFields(log.Fields{
					"reply path": reply.RName(),
					"encoding":   encoding,
				}).Warn("Get : no Decoder found for reply")
				return
			}
			value, err := decoder(data)
			if err != nil {
				logger.WithFields(log.Fields{
					"reply path": reply.RName(),
					"encoding":   encoding,
					"error":      err,
				}).Warn("Get : error decoding reply")
				return
			}
			ts := info.Tstamp()
			// @TODO: remove this when we're sure Data always come with a Timestamp.
			if ts == nil {
				ts = zcore.GenerateTimestamp()
			}
			d := Data{path, value, ts}
			l, _ := qresults[*path]
			qresults[*path] = append(l, d)

		case znet.ZNStorageFinal:
			logger.Trace("Get => ZN_STORAGE_FINAL")

		case znet.ZNEvalFinal:
			logger.Trace("Get => ZN_EVAL_FINAL")

		case znet.ZNReplyFinal:
			logger.WithField("nb replies", len(qresults)).Trace("Get => ZN_REPLY_FINAL")
			queryFinished = true
			mu.Lock()
			defer mu.Unlock()
			cond.Signal()
		}
	}

	mu.Lock()
	defer mu.Unlock()
	w.session.Query(s.Path(), s.OptionalPart(), replyCb)
	for !queryFinished {
		cond.Wait()
	}

	results := make([]Data, 0)
	if isSelectorForSeries(selector) {
		// return all data
		for _, dataset := range qresults {
			dataset = dataset.asSortedSet()
			for _, d := range dataset {
				results = append(results, d)
			}
		}
	} else {
		// return only the latest data for each path
		for _, dataset := range qresults {
			dataset = dataset.asSortedSet()
			d := dataset[len(dataset)-1]
			results = append(results, d)
		}
	}
	return results
}

// Subscribe subscribes to a selection of path/value from Zenoh.
//
// The listener will be called for each change of a path/value matching the selection.
func (w *Workspace) Subscribe(selector *Selector, listener Listener) (*SubscriptionID, error) {
	s := w.toAbsoluteSelector(selector)
	logger := logger.WithField("selector", s)
	logger.Debug("Subscribe")

	zListener := func(rname string, data []byte, info *znet.DataInfo) {
		var changes = make([]Change, 1)
		var err error
		changes[0].path, err = NewPath(rname)
		if err != nil {
			logger.WithField("notif path", rname).Warn("Subscribe received a notification for an invalid path")
			return
		}
		encoding := info.Encoding()
		decoder, ok := valueDecoders[encoding]
		if !ok {
			logger.WithFields(log.Fields{
				"notif path": rname,
				"encoding":   encoding,
			}).Warn("Subscribe received a notification with an encoding, but no Decoder found for it")
			return
		}
		changes[0].value, err = decoder(data)
		if err != nil {
			logger.WithFields(log.Fields{
				"notif path": rname,
				"encoding":   encoding,
				"error":      err,
			}).Warn("Subscribe received a notification, but Decoder failed to decode")
			return
		}

		changes[0].kind = info.Kind()
		changes[0].timestamp = info.Tstamp()

		if w.useSubroutine {
			go listener(changes)
		} else {
			listener(changes)
		}
	}

	sub, err := w.session.DeclareSubscriber(s.Path(), znet.NewSubMode(znet.ZNPushMode), zListener)
	if err != nil {
		return nil, &ZError{Msg: "Subscribe on " + s.ToString() + " failed", Code: 0, Cause: err}
	}
	return sub, nil
}

// Unsubscribe unregisters a previous subscription
func (w *Workspace) Unsubscribe(subid *SubscriptionID) error {
	err := w.session.UndeclareSubscriber(subid)
	if err != nil {
		return &ZError{Msg: "Unsubscribe failed", Code: 0, Cause: err}
	}
	return nil
}

// RegisterEval registers an "eval" function under the provided Path.
func (w *Workspace) RegisterEval(path *Path, eval Eval) error {
	p := w.toAbsolutePath(path)
	logger := logger.WithField("path", p)
	logger.Debug("RegisterEval")

	zQueryHandler := func(rname string, predicate string, repliesSender *znet.RepliesSender) {
		logger.WithFields(log.Fields{
			"rname":     rname,
			"predicate": predicate,
		}).Debug("Registered eval handling query")
		s, err := NewSelector(rname + "?" + predicate)
		if err != nil {
			logger.WithField("selector", s).Warn("Registered eval received query for an invalid selector")
			return
		}

		evalRoutine := func() {
			v := eval(p, predicateToProperties(s.Properties()))
			logger.WithFields(log.Fields{
				"rname":     rname,
				"predicate": predicate,
				"value":     v,
			}).Debug("Registered eval handling query returns")
			replies := make([]znet.Resource, 1)
			replies[0].RName = p.ToString()
			replies[0].Data = v.Encode()
			replies[0].Encoding = v.Encoding()
			replies[0].Kind = PUT
			repliesSender.SendReplies(replies)
		}
		if w.useSubroutine {
			go evalRoutine()
		} else {
			evalRoutine()
		}
	}

	e, err := w.session.DeclareEval(p.ToString(), zQueryHandler)
	if err != nil {
		return &ZError{Msg: "RegisterEval on " + p.ToString() + " failed", Code: 0, Cause: err}
	}
	w.evals[*p] = e
	return nil
}

// UnregisterEval unregisters a previously registered evaluation function.
//
// The path is the same that has been used for registration.
func (w *Workspace) UnregisterEval(path *Path) error {
	e, ok := w.evals[*path]
	if ok {
		delete(w.evals, *path)
		err := w.session.UndeclareEval(e)
		if err != nil {
			return &ZError{Msg: "UnregisterEval on " + path.ToString() + " failed", Code: 0, Cause: err}
		}
	}
	return nil
}

func (w *Workspace) toAbsolutePath(p *Path) *Path {
	if p.IsRelative() {
		return p.AddPrefix(w.path)
	}
	return p
}

func (w *Workspace) toAbsoluteSelector(s *Selector) *Selector {
	if s.IsRelative() {
		return s.AddPrefix(w.path)
	}
	return s
}

func predicateToProperties(predicate string) Properties {
	result := make(map[string]string)
	kvs := strings.Split(predicate, ";")
	for _, kv := range kvs {
		i := strings.Index(kv, "=")
		if i > 0 {
			result[kv[:i]] = kv[i+1:]
		}
	}
	return result
}
