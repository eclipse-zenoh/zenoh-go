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
	"regexp"
	"strconv"
	"strings"

	zcore "github.com/eclipse-zenoh/zenoh-go/core"
	znet "github.com/eclipse-zenoh/zenoh-go/net"
)

// ZError reports an error that occurred in zenoh.
type ZError = zcore.ZError

// Timestamp is a Zenoh Timestamp
type Timestamp = zcore.Timestamp

// Properties is a (string,string) map
type Properties map[string]string

// Listener defines the callback function that has to be registered for subscriptions
type Listener func([]Change)

// SubscriptionID identifies a Zenoh subscription
type SubscriptionID = znet.Subscriber

// Eval defines the callback function that has to be registered for evals
type Eval func(path *Path, props Properties) Value

////////////////
//    Path    //
////////////////

// Path is a set of strings separated by '/' , as in a filesystem path.
// A Path cannot contain any '*' character.
//
// Examples of paths:
//   "/demo/example/test"
//   "/com/adlink/building/fr/floor/1/office/2"
//
// A path can be absolute (i.e. starting with a `'/'`) or relative to a Workspace.
type Path struct {
	path string
}

// NewPath returns a new Path from the string p, if it's a valid path specification.
// Otherwise, it returns an error.
func NewPath(p string) (*Path, error) {
	if len(p) == 0 {
		return nil, &ZError{Msg: "Invalid path (empty String)", Code: 0, Cause: nil}
	}

	for i, c := range p {
		if c == '?' || c == '#' || c == '[' || c == ']' || c == '*' {
			return nil, &ZError{
				Msg:  "Invalid path: " + p + " (forbidden character at index " + strconv.Itoa(i) + ")",
				Code: 0, Cause: nil}
		}
	}
	result := removeUselessSlashes(p)
	return &Path{result}, nil
}

// ToString returns the Path as a string
func (p *Path) ToString() string {
	return p.path
}

// Length returns length of the path string
func (p *Path) Length() int {
	return len(p.path)
}

// IsRelative returns true if the Path is not absolute (i.e. it doesn't start with '/')
func (p *Path) IsRelative() bool {
	return p.Length() == 0 || p.path[0] != '/'
}

// AddPrefix returns a new Path made from the concatenation of the prefix and this path.
func (p *Path) AddPrefix(prefix *Path) *Path {
	result, _ := NewPath(prefix.path + "/" + p.path)
	return result
}

var slashesRegexp = regexp.MustCompile("/+")

func removeUselessSlashes(s string) string {
	result := slashesRegexp.ReplaceAllString(s, "/")
	return strings.TrimSuffix(result, "/")
}

////////////////
//  Selector  //
////////////////

// Selector is a string which is the conjunction of an path expression identifying
// a set of keys and some optional parts allowing to refine the set of Paths
// and associated Values.
//
// Structure of a selector:
//
//    /s1/s2/../sn?x>1&y<2&..&z=4(p1=v1;p2=v2;..;pn=vn)#a;x;y;..;z
//    |          | |            | |                  |  |        |
//    |-- expr --| |-- filter --| |--- properties ---|  |fragment|
//
// where:
//
// - expr: is a path expression. I.e. a string similar to a Path but with character '*'  allowed.
// A single '*' matches any set of characters in a path, except '/'.
// While `"**"` matches any set of characters in a path, including '/'.
// A path expression can be absolute (i.e. starting with a '/') or relative to a Workspace.
//
// - filter: a list of predicates separated by '&' allowing to perform filtering on the Value
// associated with the matching keys.
// Each predicate has the form "`field``operator``value`" where:
// -- `field` is the name of a field in the value (is applicable and is existing. otherwise the predicate is false).
// -- `operator` is one of a comparison operators: `<` , `>` , `<=`  , `>=`  , `=`  , `!=`.
// -- `value` is the the value to compare the field's value with.
//
// - fragment: a list of fields names allowing to return a sub-part of each value.
// This feature only applies to structured values using a "self-describing" encoding, such as JSON or XML.
// It allows to select only some fields within the structure. A new structure with only the selected fields
// will be used in place of the original value.
//
// NOTE: the filters and fragments are not yet supported in current zenoh version.
type Selector struct {
	path         string
	predicate    string
	properties   string
	fragment     string
	optionalPart string
	toString     string
}

const (
	regexPath       string = "[^\\[\\]?#]+"
	regexPredicate  string = "[^\\[\\]\\(\\)#]+"
	regexProperties string = ".*"
	regexFragment   string = ".*"
)

var pattern = regexp.MustCompile(
	fmt.Sprintf("(%s)(\\?(%s)?(\\((%s)\\))?)?(#(%s))?", regexPath, regexPredicate, regexProperties, regexFragment))

// NewSelector returns a new Selector from the string s, if it's a valid path specification.
// Otherwise, it returns an error.
func NewSelector(s string) (*Selector, error) {
	if len(s) == 0 {
		return nil, &ZError{Msg: "Invalid selector (empty String)", Code: 0, Cause: nil}
	}

	if !pattern.MatchString(s) {
		return nil, &ZError{Msg: "Invalid selector (not matching regex)", Code: 0, Cause: nil}
	}

	groups := pattern.FindStringSubmatch(s)
	path := groups[1]
	predicate := groups[3]
	properties := groups[5]
	fragment := groups[7]

	return newSelector(path, predicate, properties, fragment), nil
}

func newSelector(path string, predicate string, properties string, fragment string) *Selector {
	propertiesPart := ""
	if len(properties) > 0 {
		propertiesPart = "(" + properties + ")"
	}
	fragmentPart := ""
	if len(fragment) > 0 {
		fragmentPart = "#" + fragment
	}
	optionalPart := fmt.Sprintf("%s%s%s", predicate, propertiesPart, fragmentPart)
	toString := path
	if len(optionalPart) > 0 {
		toString += "?" + optionalPart
	}

	return &Selector{path, predicate, properties, fragment, optionalPart, toString}
}

// Path returns the path part of the Selector
func (s *Selector) Path() string {
	return s.path
}

// Predicate returns the predicate part of the Selector
func (s *Selector) Predicate() string {
	return s.predicate
}

// Properties returns the properties part of the Selector
func (s *Selector) Properties() string {
	return s.properties
}

// Fragment returns the fragment part of the Selector
func (s *Selector) Fragment() string {
	return s.fragment
}

// OptionalPart returns the optional part of the Selector
// (i.e. the part starting from the '?' character to the end of string)
func (s *Selector) OptionalPart() string {
	return s.optionalPart
}

// ToString returns the Selector as a string
func (s *Selector) ToString() string {
	return s.toString
}

// IsRelative returns true if the Path is not absolute (i.e. it doesn't start with '/')
func (s *Selector) IsRelative() bool {
	return len(s.path) == 0 || s.path[0] != '/'
}

// AddPrefix returns a new Selector made from the concatenation of the prefix and this path.
func (s *Selector) AddPrefix(prefix *Path) *Selector {
	return newSelector(prefix.path+s.path, s.predicate, s.properties, s.fragment)
}

///////////////
//    Data   //
///////////////

// Data is a zenoh data returned by a Workspace.get(selector) query.
//
// The Data objects are comparable according to their Timestamp.
// Note that zenoh makes sure that each published path/value
// has a unique timestamp accross the system.
type Data struct {
	path   *Path
	value  Value
	tstamp *Timestamp
}

// Path returns the path of the Data
func (e *Data) Path() *Path {
	return e.path
}

// Value returns the value of the Data
func (e *Data) Value() Value {
	return e.value
}

// Timestamp returns the timestamp of the Data
func (e *Data) Timestamp() *Timestamp {
	return e.tstamp
}

////////////////
//   Change   //
////////////////

// ChangeKind is a kind of change
type ChangeKind = uint8

const (
	// PUT represents a change made by a put on Zenoh
	PUT ChangeKind = 0x00
	// UPDATE represents a change made by an update on Zenoh
	UPDATE ChangeKind = 0x01
	// REMOVE represents a change made by a remove on Zenoh
	REMOVE ChangeKind = 0x02
)

// Change represents the notification of a change for a resource in zenoh.
//
// The Listener function that is registered in Workspace.subscribe(selector, listener)
// will receive a list of Changes.
type Change struct {
	path      *Path
	kind      ChangeKind
	timestamp *Timestamp
	value     Value
}

// Path returns the path impacted by the change
func (c *Change) Path() *Path {
	return c.path
}

// Kind returns the kind of change
func (c *Change) Kind() ChangeKind {
	return c.kind
}

// Timestamp returns the time of change (as registered in Zenoh)
func (c *Change) Timestamp() *Timestamp {
	return c.timestamp
}

// Value returns the value that changed
func (c *Change) Value() Value {
	return c.value
}

////////////////
//  Encoding  //
////////////////

// Encoding is a description of the Value format, allowing zenoh to know
// how to encode/decode the value to/from a bytes buffer.
type Encoding = uint8

// Known encodings:
const (
	// RAW: The value has a RAW encoding (i.e. it's a bytes buffer).
	RAW Encoding = 0x00

	// STRING: The value is an UTF-8 string.
	STRING Encoding = 0x02

	// PROPERTIES: The value if a list of keys/values, encoded as an UTF-8 string.
	// The keys/values are separated by ';' character, and each key is separated
	// from its associated value (if any) with a '=' character.
	PROPERTIES Encoding = 0x03

	// JSON The value is a JSON structure in an UTF-8 string.
	JSON Encoding = 0x04

	// INT The value is an integer as an UTF-8 string.
	INT Encoding = 0x06

	// FLOAT The value is a float as an UTF-8 string.
	FLOAT Encoding = 0x07
)

var valueDecoders = map[Encoding]ValueDecoder{}

// RegisterValueDecoder registers a ValueDecoder function with it's Encoding
func RegisterValueDecoder(encoding Encoding, decoder ValueDecoder) error {
	if valueDecoders[encoding] != nil {
		return &ZError{Msg: "Already registered ValueDecoder for Encoding " + strconv.Itoa(int(encoding)),
			Code: 0, Cause: nil}
	}
	valueDecoders[encoding] = decoder
	return nil
}

func init() {
	RegisterValueDecoder(RAW, rawDecoder)
	RegisterValueDecoder(STRING, stringDecoder)
	RegisterValueDecoder(PROPERTIES, propertiesDecoder)
	RegisterValueDecoder(JSON, stringDecoder)
	RegisterValueDecoder(INT, intDecoder)
	RegisterValueDecoder(FLOAT, floatDecoder)
}

////////////////
//   Value    //
////////////////

// Value is the interface of a value that, associated to a Path, can be published into zenoh
// via Workspace.put(Path, Value), or retrieved via Workspace.get(Selector) or
// via a subscription (Workspace.subscribe(Selector, Listener)).
type Value interface {
	Encoding() Encoding
	Encode() []byte
	ToString() string
}

// ValueDecoder is a decoder for a Value
type ValueDecoder func([]byte) (Value, error)

///////////////////
//   RAW Value   //
///////////////////

// RawValue is a RAW value (i.e. a bytes buffer)
type RawValue struct {
	buf []byte
}

// NewRawValue returns a new RawValue
func NewRawValue(buf []byte) *RawValue {
	return &RawValue{buf}
}

// Encoding returns the encoding flag for a RawValue
func (v *RawValue) Encoding() Encoding {
	return RAW
}

// Encode returns the value encoded as a []byte
func (v *RawValue) Encode() []byte {
	return v.buf
}

// ToString returns the value as a string
func (v *RawValue) ToString() string {
	return fmt.Sprintf("[x %d]", v.buf)
}

func rawDecoder(buf []byte) (Value, error) {
	return &RawValue{buf}, nil
}

//////////////////////
//   STRING Value   //
//////////////////////

// StringValue is a STRING value (i.e. just a string)
type StringValue struct {
	s string
}

// NewStringValue returns a new StringValue
func NewStringValue(s string) *StringValue {
	return &StringValue{s}
}

// Encoding returns the encoding flag for a StringValue
func (v *StringValue) Encoding() Encoding {
	return STRING
}

// Encode returns the value encoded as a []byte
func (v *StringValue) Encode() []byte {
	return []byte(v.s)
}

// ToString returns the value as a string
func (v *StringValue) ToString() string {
	return v.s
}

func stringEncoder(s string) []byte {
	return []byte(s)
}

func stringDecoder(buf []byte) (Value, error) {
	return &StringValue{string(buf)}, nil
}

//////////////////////////
//   PROPERTIES Value   //
//////////////////////////

// PropertiesValue is a PROPERTIES value (i.e. a map[string]string)
type PropertiesValue struct {
	p Properties
}

// NewPropertiesValue returns a new PropertiesValue
func NewPropertiesValue(p Properties) *PropertiesValue {
	return &PropertiesValue{p}
}

// Encoding returns the encoding flag for a PropertiesValue
func (v *PropertiesValue) Encoding() Encoding {
	return PROPERTIES
}

// Encode returns the value encoded as a []byte
func (v *PropertiesValue) Encode() []byte {
	return []byte(v.ToString())
}

const (
	propSep = ";"
	kvSep   = "="
)

// ToString returns the value as a string
func (v *PropertiesValue) ToString() string {
	builder := new(strings.Builder)
	i := 0
	for key, val := range v.p {
		builder.WriteString(key)
		builder.WriteString(kvSep)
		builder.WriteString(val)
		i++
		if i < len(v.p) {
			builder.WriteString(propSep)
		}
	}
	return builder.String()
}

func propertiesOfString(s string) Properties {
	p := make(Properties)
	if len(s) > 0 {
		for _, kv := range strings.Split(s, propSep) {
			i := strings.Index(kv, kvSep)
			if i < 0 {
				p[kv] = ""
			} else {
				p[kv[:i]] = kv[i+1:]
			}
		}
	}
	return p
}

func propertiesDecoder(buf []byte) (Value, error) {
	return &PropertiesValue{propertiesOfString(string(buf))}, nil
}

//////////////////////
//    INT Value     //
//////////////////////

// IntValue is a INT value (i.e. an int64)
type IntValue struct {
	i int64
}

// NewIntValue returns a new IntValue
func NewIntValue(i int64) *IntValue {
	return &IntValue{i}
}

// Encoding returns the encoding flag for an IntValue
func (v *IntValue) Encoding() Encoding {
	return INT
}

// Encode returns the value encoded as a []byte
func (v *IntValue) Encode() []byte {
	return intEncoder(v.i)
}

// ToString returns the value as a string
func (v *IntValue) ToString() string {
	return strconv.FormatInt(v.i, 10)
}

func intEncoder(i int64) []byte {
	return []byte(strconv.FormatInt(i, 10))
}

func intDecoder(buf []byte) (Value, error) {
	i, err := strconv.ParseInt(string(buf), 10, 64)
	if err != nil {
		return nil, &ZError{Msg: "Failed to decode INT value", Code: 0, Cause: err}
	}
	return &IntValue{i}, nil
}

//////////////////////
//    FLOAT Value     //
//////////////////////

// FloatValue is a FLOAT value (i.e. a float64)
type FloatValue struct {
	f float64
}

// NewFloatValue returns a new FloatValue
func NewFloatValue(f float64) *FloatValue {
	return &FloatValue{f}
}

// Encoding returns the encoding flag for an FloatValue
func (v *FloatValue) Encoding() Encoding {
	return FLOAT
}

// Encode returns the value encoded as a []byte
func (v *FloatValue) Encode() []byte {
	return []byte(v.ToString())
}

// ToString returns the value as a string
func (v *FloatValue) ToString() string {
	return strconv.FormatFloat(v.f, 'g', -1, 64)
}

func floatEncoder(f float64) []byte {
	return []byte(strconv.FormatFloat(f, 'g', -1, 64))
}

func floatDecoder(buf []byte) (Value, error) {
	f, err := strconv.ParseFloat(string(buf), 64)
	if err != nil {
		return nil, &ZError{Msg: "Failed to decode FLOAT value", Code: 0, Cause: err}
	}
	return &FloatValue{f}, nil
}
