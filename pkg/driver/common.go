/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
/*
Copyright (c) 2024 Open-E, Inc.
All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License"); you may
not use this file except in compliance with the License. You may obtain
a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
License for the specific language governing permissions and limitations
under the License.
*/

package driver

import (
	"sort"

	"golang.org/x/net/context"

	jrest "github.com/open-e/joviandss-kubernetescsi/pkg/rest"
)

// How a CSI list call is answered.
//
// The appliance pages its lists by number, and a page number names other
// entries as soon as the list changes, while a name keeps its place in the
// order. So every call reads the whole appliance list, from page 0, sorts
// it by a key made of names, and returns the entries that follow the place
// its token gives. Nothing kept between two CSI calls names an appliance
// page.

// perPage is the page size of every list request.
const perPage = 100

// readList reads one appliance list from page 0 to its last page.
// getPage returns one page as the appliance sent it: the stop rule
// counts a page before any entry is dropped from it.
func readList[T any](ctx context.Context,
	getPage func(ctx context.Context, page, perPage int) ([]T, jrest.RestError),
) ([]T, jrest.RestError) {
	var read []T
	for page := 0; ; page++ {
		ents, err := getPage(ctx, page, perPage)
		if err != nil {
			return nil, err
		}
		read = append(read, ents...)
		if len(ents) < perPage {
			// an empty page, or the last one, which is not full
			return read, nil
		}
	}
}

// sameOptional reports whether two optional values agree: both are
// missing, or both hold the same value.
func sameOptional[V comparable](a, b *V) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}

// mergePages takes the entries of one read, in the order of the read. A
// list that changes while it is read can return a name twice. Of two
// entries of one name that same counts as one object, the one read first is
// dropped, since the later one holds the newer state; two that differ both
// stay: two snapshots of one name on two volumes, or an object deleted and
// made again under its name during the read.
func mergePages[T any](read []T, name func(T) string, same func(a, b T) bool) []T {
	dropped := make([]bool, len(read))
	staying := make(map[string][]int) // name -> its entries that stay
	for i, e := range read {
		var differ []int
		for _, j := range staying[name(e)] {
			if same(read[j], e) {
				dropped[j] = true // the later entry holds the newer state
			} else {
				differ = append(differ, j)
			}
		}
		staying[name(e)] = append(differ, i)
	}
	merged := make([]T, 0, len(read))
	for i, e := range read {
		if !dropped[i] {
			merged = append(merged, e)
		}
	}
	return merged
}

// The key of an entry is its place in the order of its list, one string
// compared byte by byte: the name for volumes and for the snapshots of one
// volume; for a list of all snapshots, the parent's name, "!" and the
// snapshot's name. "!" sorts before every character a name of the plugin
// holds, so the order is by parent, then by name.
func snapshotsListKey(parent, name string) string { return keyPart(parent) + "!" + keyPart(name) }

// keyPart writes a name into a key. A name the plugin made holds letters,
// digits, "-", "_" and "=" and is written as it is. Any other byte, which
// only a name made by hand can hold, is written as "%" and two hex digits,
// so that a key, and the token made from it, is printable ASCII whatever
// the appliance holds.
func keyPart(name string) string {
	plain := true
	for i := 0; i < len(name); i++ {
		if !keyByte(name[i]) {
			plain = false
			break
		}
	}
	if plain {
		return name
	}
	const hex = "0123456789ABCDEF"
	out := make([]byte, 0, len(name)+8)
	for i := 0; i < len(name); i++ {
		if c := name[i]; keyByte(c) {
			out = append(out, c)
		} else {
			out = append(out, '%', hex[c>>4], hex[c&15])
		}
	}
	return string(out)
}

// keyByte reports whether a byte of a name goes into a key as it is:
// printable ASCII but "!", which parts the two names of a key, "%", which
// starts an escaped byte, and "~", the last printable character, after
// which a bound has no room.
func keyByte(c byte) bool {
	return c > '!' && c < '~' && c != '%'
}

// cutAfter sorts the entries by key and returns the first limit of them
// from the place the token gives, all of them when limit is 0; limit is
// never negative here. When entries follow those it returns, next is the
// token of the call after it, else nil.
func cutAfter[T any](ents []T, key func(T) string, from *CSIListingToken, limit int) (out []T, next *CSIListingToken) {
	// stable: two entries of one key keep the order of the read
	sort.SliceStable(ents, func(i, j int) bool { return key(ents[i]) < key(ents[j]) })
	keys := make([]string, len(ents))
	for i, e := range ents {
		keys[i] = key(e)
	}
	rest := ents[from.start(keys):]
	if limit == 0 || len(rest) <= limit {
		return rest, nil
	}
	out = rest[:limit]
	last := key(out[limit-1])
	// When the cut falls between two entries of one key, the second is not
	// returned: the next call starts after that key.
	for _, e := range rest[limit:] {
		if k := key(e); k > last {
			return out, from.nextCall(last, k)
		}
	}
	return out, nil
}

// listEntries answers one CSI list call. It reads the whole list once,
// whatever the token. own drops what the reply does not carry, a name
// the plugin did not make (with its warning) and the other driver's
// snapshots, before the cut, so that limit counts entries of the reply.
func listEntries[T any](ctx context.Context,
	getPage func(ctx context.Context, page, perPage int) ([]T, jrest.RestError),
	name func(T) string, same func(a, b T) bool, own func(T) bool,
	key func(T) string, from *CSIListingToken, limit int,
) (out []T, next *CSIListingToken, err jrest.RestError) {
	read, err := readList(ctx, getPage)
	if err != nil {
		return nil, nil, err
	}
	ents := make([]T, 0, len(read))
	for _, e := range mergePages(read, name, same) {
		if own(e) {
			ents = append(ents, e)
		}
	}
	out, next = cutAfter(ents, key, from, limit)
	return out, next, nil
}
