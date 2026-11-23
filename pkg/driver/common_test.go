package driver

import (
	"context"
	"fmt"
	"math/rand"
	"reflect"
	"sort"
	"strings"
	"testing"

	jrest "github.com/open-e/joviandss-kubernetescsi/pkg/rest"
)

// entry is an entry of a list in these tests: a name, the name of its
// parent for a list of all snapshots, and what tells two objects of one
// name apart.
type entry struct {
	name   string
	parent string
	guid   string
}

func entryName(e entry) string   { return e.name }
func sameEntry(a, b entry) bool  { return a.guid == b.guid }
func volumeKey(e entry) string   { return keyPart(e.name) }
func snapshotKey(e entry) string { return snapshotsListKey(e.parent, e.name) }
func everyEntry(entry) bool      { return true }
func keysOf(es []entry, key func(entry) string) (keys []string) {
	for _, e := range es {
		keys = append(keys, key(e))
	}
	return keys
}

// pagesOf serves a list page by page, as an appliance does, and counts the
// requests.
type pagesOf struct {
	list     []entry
	requests []string
	fail     jrest.RestError
	failAt   int
}

func (p *pagesOf) getPage(ctx context.Context, page, perPage int) ([]entry, jrest.RestError) {
	p.requests = append(p.requests, fmt.Sprintf("page=%d&per_page=%d", page, perPage))
	if p.fail != nil && page == p.failAt {
		return nil, p.fail
	}
	from := page * perPage
	if from >= len(p.list) {
		return nil, nil
	}
	to := from + perPage
	if to > len(p.list) {
		to = len(p.list)
	}
	return append([]entry(nil), p.list[from:to]...), nil
}

func numbered(n int) []entry {
	list := make([]entry, n)
	for i := range list {
		list[i] = entry{name: fmt.Sprintf("vp_%05d", i), guid: fmt.Sprintf("g%d", i)}
	}
	return list
}

// A list is read from page 0 with per_page=100, and the read ends on the
// first page that is not full, an empty one included.
func TestReadListEndsOnAPageThatIsNotFull(t *testing.T) {
	for _, c := range []struct{ entries, requests int }{
		{0, 1}, {1, 1}, {99, 1}, {100, 2}, {101, 2}, {200, 3}, {237, 3},
	} {
		p := &pagesOf{list: numbered(c.entries)}
		read, err := readList(context.Background(), p.getPage)
		if err != nil {
			t.Fatal(err.Error())
		}
		if len(read) != c.entries || len(p.requests) != c.requests {
			t.Errorf("a list of %d entries: %d read in %d requests, want %d requests", c.entries, len(read), len(p.requests), c.requests)
		}
		for i, r := range p.requests {
			if want := fmt.Sprintf("page=%d&per_page=100", i); r != want {
				t.Errorf("a list of %d entries: request %d is %s, want %s", c.entries, i, r, want)
			}
		}
	}

	// an error of a page fails the read: a list is never returned in part
	p := &pagesOf{list: numbered(250), fail: jrest.GetError(jrest.RestErrorRequestTimeout, "no reply"), failAt: 1}
	read, err := readList(context.Background(), p.getPage)
	if err == nil || read != nil || err.GetCode() != jrest.RestErrorRequestTimeout {
		t.Errorf("a read whose second page fails: %d entries, error %v", len(read), err)
	}
}

// Of two entries of one name that are one object, the one read later
// stays; two that differ both stay.
func TestMergePages(t *testing.T) {
	a1 := entry{name: "a", guid: "1"}
	a1later := entry{name: "a", guid: "1", parent: "later"}
	a2 := entry{name: "a", guid: "2"}
	b := entry{name: "b", guid: "3"}
	for _, c := range []struct {
		what string
		read []entry
		want []entry
	}{
		{"nothing twice", []entry{a1, b}, []entry{a1, b}},
		{"one object twice: the later entry stays", []entry{a1, b, a1later}, []entry{b, a1later}},
		{"two objects of one name: both stay", []entry{a1, b, a2}, []entry{a1, b, a2}},
		{"an object twice and another of its name", []entry{a1, a2, a1later}, []entry{a2, a1later}},
		{"an object three times", []entry{a1, a1, a1later}, []entry{a1later}},
		{"nothing read", nil, []entry{}},
	} {
		got := mergePages(c.read, entryName, sameEntry)
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("%s: %v, want %v", c.what, got, c.want)
		}
	}

	// a list without a field that tells two objects apart: a name that
	// comes twice is dropped once
	got := mergePages([]entry{a1, a2, b}, entryName, func(a, b entry) bool { return true })
	if !reflect.DeepEqual(got, []entry{a2, b}) {
		t.Errorf("a name twice in a list of names only: %v", got)
	}

	// an optional value: both missing, or both the same
	yes, no := true, false
	for _, c := range []struct {
		a, b *bool
		want bool
	}{{nil, nil, true}, {&yes, &yes, true}, {&yes, &no, false}, {nil, &no, false}, {&yes, nil, false}} {
		if sameOptional(c.a, c.b) != c.want {
			t.Errorf("sameOptional(%v, %v) is not %v", c.a, c.b, c.want)
		}
	}
}

// walk follows next_token from the first call to the last, on a list that
// current may alter before each call, and returns what the calls returned,
// and whether a token of the walk carried hashes in place of a bound. It
// fails the test on a token that is not ASCII of at most 128 bytes.
func walk(t *testing.T, list ListID, key func(entry) string, limit int, current func(call int, got []entry) []entry) (got []entry, calls int, hashed bool) {
	t.Helper()
	token := ""
	for {
		from, err := NewCSIListingTokenFromTokenString(token, list)
		if err != nil {
			t.Fatalf("the token %q of this walk is refused: %s", token, err.Error())
		}
		out, next := cutAfter(append([]entry(nil), current(calls, got)...), key, from, limit)
		calls++
		got = append(got, out...)
		if limit > 0 && len(out) > limit {
			t.Fatalf("a call for %d entries returned %d", limit, len(out))
		}
		if next == nil {
			return got, calls, hashed
		}
		hashed = hashed || next.prefix != ""
		if len(out) == 0 {
			t.Fatalf("a call returned no entry and a token")
		}
		token = next.Token()
		if len(token) > tokenMax {
			t.Fatalf("a token of %d bytes: %q", len(token), token)
		}
		for i := 0; i < len(token); i++ {
			if token[i] < '!' || token[i] > '~' {
				t.Fatalf("a token that is not printable ASCII: %q", token)
			}
		}
		if calls > 100000 {
			t.Fatal("the walk does not end")
		}
	}
}

// pluginName makes a name as the plugin makes them, n bytes long, from a
// shared beginning and a number.
func pluginName(prefix string, n int, i int) string {
	tail := fmt.Sprintf("%06d", i)
	if len(prefix)+len(tail) >= n {
		return prefix[:n-len(tail)] + tail
	}
	return prefix + strings.Repeat("A", n-len(prefix)-len(tail)) + tail
}

// The lists of the walk tests: short names, names of the longest size an
// appliance takes that differ only at their end, and both in two-name keys.
func walkLists() map[string][]entry {
	rnd := rand.New(rand.NewSource(7))
	lists := map[string][]entry{}
	var short, long, mixed []entry
	for i := 0; i < 57; i++ {
		short = append(short, entry{name: fmt.Sprintf("vp_pvc-%04x-%d", rnd.Intn(1<<16), i), guid: fmt.Sprint(i)})
		long = append(long, entry{name: pluginName("vb_"+strings.Repeat("ORSXG5DJNZTQ", 12), 131, i), guid: fmt.Sprint(i)})
	}
	for i := 0; i < 40; i++ {
		parent := pluginName("vp_"+strings.Repeat("k", 100), 131, i%5)
		mixed = append(mixed, entry{name: pluginName("sp_"+strings.Repeat("s", 110), 131, i), parent: parent, guid: fmt.Sprint(i)})
		mixed = append(mixed, entry{name: fmt.Sprintf("sp_snap-%d", i%7), parent: fmt.Sprintf("vp_vol-%d", i), guid: fmt.Sprint(1000 + i)})
	}
	rnd.Shuffle(len(short), func(i, j int) { short[i], short[j] = short[j], short[i] })
	rnd.Shuffle(len(long), func(i, j int) { long[i], long[j] = long[j], long[i] })
	rnd.Shuffle(len(mixed), func(i, j int) { mixed[i], mixed[j] = mixed[j], mixed[i] })
	lists["short names"] = short
	lists["names of 131 bytes that differ at the end"] = long
	lists["snapshots of many volumes"] = mixed
	lists["one entry"] = short[:1]
	lists["no entry"] = nil
	return lists
}

// distinct reports whether sorted keys are all different.
func distinct(keys []string) bool {
	for i := 1; i < len(keys); i++ {
		if keys[i] == keys[i-1] {
			return false
		}
	}
	return true
}

// Calls that start with no token and follow next_token return every entry
// of a list that does not change, each exactly once and in key order, for
// any max_entries; every token is ASCII of at most 128 bytes; and
// next_token is empty exactly when no entry follows.
func TestWalkOfAListThatDoesNotChange(t *testing.T) {
	for what, list := range walkLists() {
		for _, key := range []func(entry) string{volumeKey, snapshotKey} {
			want := append([]entry(nil), list...)
			sort.SliceStable(want, func(i, j int) bool { return key(want[i]) < key(want[j]) })
			if !distinct(keysOf(want, key)) {
				continue // two entries of one key: TestCutBetweenTwoEntriesOfOneKey
			}
			for _, limit := range []int{0, 1, 2, 3, 7, 10, 56, 57, 58, 1000} {
				got, calls, _ := walk(t, VolumesList(), key, limit, func(int, []entry) []entry { return list })
				if !reflect.DeepEqual(keysOf(got, key), keysOf(want, key)) {
					t.Errorf("%s, max_entries %d: the walk returned %d entries, want the %d of the list, each once and in key order",
						what, limit, len(got), len(want))
				}
				wantCalls := 1
				if limit > 0 && len(list) > 0 {
					// no call returns fewer than max_entries while entries follow,
					// and no call is made after the last entry
					wantCalls = (len(list) + limit - 1) / limit
				}
				if calls != wantCalls {
					t.Errorf("%s, max_entries %d: %d calls, want %d", what, limit, calls, wantCalls)
				}
			}
		}
	}
}

// An entry that exists through the whole walk is returned once while other
// entries are created or deleted between the calls, the two the token was
// made from among them. Where the token carries hashes in place of a bound
// and both of its entries are gone, the entries that share their beginning
// come again: none is missed.
func TestWalkOfAListThatChanges(t *testing.T) {
	for what, list := range walkLists() {
		if len(list) < 10 {
			continue
		}
		for _, key := range []func(entry) string{volumeKey, snapshotKey} {
			sorted := append([]entry(nil), list...)
			sort.SliceStable(sorted, func(i, j int) bool { return key(sorted[i]) < key(sorted[j]) })
			if !distinct(keysOf(sorted, key)) {
				continue
			}
			place := map[string]int{}
			for i, e := range sorted {
				place[key(e)] = i
			}
			for _, limit := range []int{1, 2, 3, 7} {
				for seed := int64(0); seed < 20; seed++ {
					rnd := rand.New(rand.NewSource(seed))
					// every third entry stays through the walk; the others come and go
					stable := func(i int) bool { return i%3 == 0 }
					bothGone := 0
					current := func(call int, got []entry) []entry {
						// the last entry returned and the one after it: the two
						// the token of this call was made from
						last, next := -1, -1
						if len(got) > 0 {
							last = place[key(got[len(got)-1])]
							next = last + 1
						}
						remove := call%2 == 1 && bothGone < 3
						if remove && last >= 0 && next < len(sorted) && !stable(last) && !stable(next) {
							bothGone++
						}
						var now []entry
						for i, e := range sorted {
							switch {
							case stable(i):
								now = append(now, e)
							case i == last || i == next:
								if !remove {
									now = append(now, e)
								}
							case rnd.Intn(3) > 0:
								now = append(now, e)
							}
						}
						return now
					}
					got, _, hashed := walk(t, VolumesList(), key, limit, current)
					seen := map[string]int{}
					for _, e := range got {
						seen[key(e)]++
					}
					for i, e := range sorted {
						if !stable(i) {
							continue
						}
						n := seen[key(e)]
						if n == 0 || (n > 1 && !hashed) {
							t.Fatalf("%s, max_entries %d, seed %d: an entry that stayed through the walk was returned %d times",
								what, limit, seed, n)
						}
					}
					if !hashed && !sort.StringsAreSorted(keysOf(got, key)) {
						t.Fatalf("%s, max_entries %d, seed %d: the walk is not in key order", what, limit, seed)
					}
				}
			}
		}
	}
}

// Two entries of one key that both stay keep the order of the read. When
// the cut falls between them the second is not returned, since the next
// call starts after that key.
func TestCutBetweenTwoEntriesOfOneKey(t *testing.T) {
	a := entry{name: "sp_a", guid: "1"}
	b1 := entry{name: "sp_b", guid: "2"}
	b2 := entry{name: "sp_b", guid: "3"}
	c := entry{name: "sp_c", guid: "4"}
	list := []entry{c, b1, a, b2}

	got, _, _ := walk(t, VolumesList(), volumeKey, 0, func(int, []entry) []entry { return list })
	if !reflect.DeepEqual(got, []entry{a, b1, b2, c}) {
		t.Errorf("no limit: %v", got)
	}
	got, _, _ = walk(t, VolumesList(), volumeKey, 3, func(int, []entry) []entry { return list })
	if !reflect.DeepEqual(got, []entry{a, b1, b2, c}) {
		t.Errorf("a cut after both: %v", got)
	}
	got, _, _ = walk(t, VolumesList(), volumeKey, 2, func(int, []entry) []entry { return list })
	if !reflect.DeepEqual(got, []entry{a, b1, c}) {
		t.Errorf("a cut between the two: %v, want the first of them and no second", got)
	}

	// the two are the last of the list: no entry follows, so no token
	from, _ := NewCSIListingTokenFromTokenString("", VolumesList())
	out, next := cutAfter([]entry{a, b1, b2}, volumeKey, from, 2)
	if len(out) != 2 || next != nil {
		t.Errorf("a cut between the last two entries, of one key: %d entries, token %v", len(out), next)
	}
}

// listEntries reads the whole list once on every call, whatever the token,
// drops what the reply does not carry before the cut, and so returns
// max_entries entries whenever that many of its own follow.
func TestListEntriesCountsItsOwnEntries(t *testing.T) {
	var list []entry
	for i := 0; i < 250; i++ {
		e := entry{name: fmt.Sprintf("vp_%05d", i), guid: fmt.Sprint(i)}
		if i%5 != 0 {
			e.name = fmt.Sprintf("other-%05d", i) // not the plugin's: four of every five
		}
		list = append(list, e)
	}
	// one entry was returned by two pages: the appliance's list changed
	// while it was read
	list = append(list, entry{name: "vp_00000", guid: "0"})
	own := func(e entry) bool { return strings.HasPrefix(e.name, "vp_") }

	var got []entry
	token, calls := "", 0
	for {
		p := &pagesOf{list: list}
		from, err := NewCSIListingTokenFromTokenString(token, VolumesList())
		if err != nil {
			t.Fatal(err.Error())
		}
		out, next, err := listEntries(context.Background(), p.getPage, entryName, sameEntry, own, volumeKey, from, 10)
		if err != nil {
			t.Fatal(err.Error())
		}
		calls++
		if want := []string{"page=0&per_page=100", "page=1&per_page=100", "page=2&per_page=100"}; !reflect.DeepEqual(p.requests, want) {
			t.Fatalf("call %d read %v, want the whole list once, from page 0: %v", calls, p.requests, want)
		}
		got = append(got, out...)
		if next == nil {
			break
		}
		if len(out) != 10 {
			t.Fatalf("call %d returned %d entries and a token, want 10", calls, len(out))
		}
		token = next.Token()
	}
	if len(got) != 50 || calls != 5 {
		t.Errorf("%d entries in %d calls, want the plugin's 50 in 5", len(got), calls)
	}
	for i, e := range got {
		if e.name != fmt.Sprintf("vp_%05d", i*5) {
			t.Fatalf("entry %d is %s", i, e.name)
		}
	}

	// an error of the read is the error of the call
	p := &pagesOf{list: list, fail: jrest.GetError(jrest.RestErrorResourceDNEPool, "no pool"), failAt: 0}
	from, _ := NewCSIListingTokenFromTokenString("", VolumesList())
	out, next, err := listEntries(context.Background(), p.getPage, entryName, sameEntry, own, volumeKey, from, 10)
	if err == nil || err.GetCode() != jrest.RestErrorResourceDNEPool || out != nil || next != nil {
		t.Errorf("a read that fails: %d entries, token %v, error %v", len(out), next, err)
	}
}

// A name made by hand can hold any byte; a key holds printable ASCII only,
// keeps two names apart, and keeps a name of the plugin as it is.
func TestKeyPart(t *testing.T) {
	for _, name := range []string{"vp_pvc-1", "sb_ORSXG5A=", "vs_0123ABCDEF", "v_x"} {
		if keyPart(name) != name {
			t.Errorf("the plugin's name %q is written %q", name, keyPart(name))
		}
	}
	seen := map[string]string{}
	for _, name := range []string{"sp_a b", "sp_a%20b", "sp_a!b", "sp_a~b", "sp_ä", "sp_a\nb", "sp_a\x00b", "sp_a", "sp_a%"} {
		k := keyPart(name)
		for i := 0; i < len(k); i++ {
			if k[i] <= '!' || k[i] >= '~' {
				t.Errorf("the key part %q of %q holds the byte %#x", k, name, k[i])
			}
		}
		if other, twice := seen[k]; twice {
			t.Errorf("%q and %q have one key part, %q", name, other, k)
		}
		seen[k] = name
	}
	if snapshotsListKey("vp_a", "sp_b!c") == snapshotsListKey("vp_a!sp_b", "c") {
		t.Error("two entries of different parents have one key")
	}
	if !(snapshotsListKey("vp_a", "sp_z") < snapshotsListKey("vp_a-1", "sp_a")) {
		t.Error("the order is not by parent first")
	}
}
