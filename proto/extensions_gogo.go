// Copyright (c) 2013, Vastech SA (PTY) LTD. All rights reserved.
// http://code.google.com/p/gogoprotobuf/gogoproto
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are
// met:
//
//     * Redistributions of source code must retain the above copyright
// notice, this list of conditions and the following disclaimer.
//     * Redistributions in binary form must reproduce the above
// copyright notice, this list of conditions and the following disclaimer
// in the documentation and/or other materials provided with the
// distribution.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
// "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
// LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
// A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
// OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
// SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
// LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
// DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
// THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

package proto

import (
	"bytes"
	"fmt"
	"reflect"
	"sort"
	"strings"
)

func GetBoolExtension(pb extendableProto, extension *ExtensionDesc, ifnotset bool) bool {
	if reflect.ValueOf(pb).IsNil() {
		return ifnotset
	}
	value, err := GetExtension(pb, extension)
	if err != nil {
		return ifnotset
	}
	if value == nil {
		return ifnotset
	}
	if value.(*bool) == nil {
		return ifnotset
	}
	return *(value.(*bool))
}

func (this *Extension) Equal(that *Extension) bool {
	return bytes.Equal(this.enc, that.enc)
}

func SizeOfExtensionMap(m map[int32]Extension) (n int) {
	return sizeExtensionMap(m)
}

type sortableMapElem struct {
	field int32
	ext   Extension
}

func newSortableExtensionsFromMap(m map[int32]Extension) sortableExtensions {
	s := make(sortableExtensions, 0, len(m))
	for k, v := range m {
		s = append(s, &sortableMapElem{field: k, ext: v})
	}
	return s
}

type sortableExtensions []*sortableMapElem

func (this sortableExtensions) Len() int { return len(this) }

func (this sortableExtensions) Swap(i, j int) { this[i], this[j] = this[j], this[i] }

func (this sortableExtensions) Less(i, j int) bool { return this[i].field < this[j].field }

func (this sortableExtensions) String() string {
	sort.Sort(this)
	ss := make([]string, len(this))
	for i := range this {
		ss[i] = fmt.Sprintf("%d: %v", this[i].field, this[i].ext)
	}
	return "map[" + strings.Join(ss, ",") + "]"
}

func StringFromExtensionsMap(m map[int32]Extension) string {
	return newSortableExtensionsFromMap(m).String()
}

func EncodeExtensionMap(m map[int32]Extension, data []byte) (n int, err error) {
	if err := encodeExtensionMap(m); err != nil {
		return 0, err
	}
	keys := make([]int, 0, len(m))
	for k := range m {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	for _, k := range keys {
		n += copy(data[n:], m[int32(k)].enc)
	}
	return n, nil
}

func GetRawExtension(m map[int32]Extension, id int32) ([]byte, error) {
	if m[id].value == nil || m[id].desc == nil {
		return m[id].enc, nil
	}
	if err := encodeExtensionMap(m); err != nil {
		return nil, err
	}
	return m[id].enc, nil
}

func BytesToExtensionsMap(buf []byte) (map[int32]Extension, error) {
	m := make(map[int32]Extension)
	i := 0
	for i < len(buf) {
		v, n := DecodeVarint(buf[i:])
		if n <= 0 {
			return nil, fmt.Errorf("unable to decode varint")
		}
		i += n
		l, n := DecodeVarint(buf[i:])
		if n <= 0 {
			return nil, fmt.Errorf("unable to decode varint")
		}
		i += n
		m[int32(v)] = Extension{enc: buf[i : i+int(l)]}
	}
	return m, nil
}

func NewExtension(e []byte) Extension {
	ee := Extension{enc: make([]byte, len(e))}
	copy(ee.enc, e)
	return ee
}

func (this Extension) GoString() string {
	return fmt.Sprintf("proto.NewExtension(%#v)", this.enc)
}
