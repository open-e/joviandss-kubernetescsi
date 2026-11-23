package driver

import (
	"math/rand"
	"sort"
	"strings"
	"testing"
)

// allStrings returns every string over alphabet of at most n bytes.
func allStrings(alphabet string, n int) []string {
	out := []string{""}
	last := []string{""}
	for i := 0; i < n; i++ {
		var next []string
		for _, s := range last {
			for _, c := range alphabet {
				next = append(next, s+string(c))
			}
		}
		out = append(out, next...)
		last = next
	}
	return out
}

// bound(last, next) is at or after last and before next, and no shorter
// string is.
func TestBoundIsTheShortestBetweenTwoKeys(t *testing.T) {
	all := allStrings("abc", 4)
	sort.Strings(all)
	for i, last := range all {
		if last == "" {
			continue
		}
		for _, next := range all[i+1:] {
			b := bound(last, next)
			if !(last <= b && b < next) {
				t.Fatalf("bound(%q, %q) = %q is not at or after the first and before the second", last, next, b)
			}
			for _, s := range all {
				if len(s) >= len(b) {
					continue
				}
				if last <= s && s < next {
					t.Fatalf("bound(%q, %q) = %q, and %q is shorter", last, next, b, s)
				}
			}
		}
	}
	// keys as the plugin makes them
	rnd := rand.New(rand.NewSource(1))
	const alphabet = "!%-0123456789=ABCXYZ_abcxyz}"
	for n := 0; n < 20000; n++ {
		mk := func() string {
			b := make([]byte, 1+rnd.Intn(6))
			for i := range b {
				b[i] = alphabet[rnd.Intn(len(alphabet))]
			}
			return string(b)
		}
		last, next := mk(), mk()
		if last == next {
			continue
		}
		if last > next {
			last, next = next, last
		}
		b := bound(last, next)
		if !(last <= b && b < next) {
			t.Fatalf("bound(%q, %q) = %q", last, next, b)
		}
		for i := 0; i < len(b); i++ {
			if b[i] < '!' || b[i] > '~' {
				t.Fatalf("bound(%q, %q) = %q holds the byte %#x", last, next, b, b[i])
			}
		}
	}
}

// A token read back resumes where the token written did, in both of its
// forms, and is ASCII of at most 128 bytes.
func TestTokenRoundTrip(t *testing.T) {
	long := strings.Repeat("vb_ORSXG5DJNZTQ", 9)[:125]
	for _, list := range []ListID{VolumesList(), SnapshotsList(), VolumeSnapshotsList("vp_pvc-1")} {
		for _, c := range []struct {
			what       string
			last, next string
			hashes     bool
		}{
			{"short keys", "vp_a", "vp_b", false},
			{"one key the beginning of the other", "vp_a", "vp_a-1", false},
			{"two names in a key", "vp_a!sp_1", "vp_a!sp_2", false},
			{"keys of 131 bytes that differ at the end", long + "AAAAA1", long + "AAAAA2", true},
			{"two keys of 263 bytes that differ at the end", long + long + "000001", long + long + "000002", true},
			{"a long key and a short one", long + "AAAAA1", "vp_z", false},
		} {
			first, _ := NewCSIListingTokenFromTokenString("", list)
			tok := first.nextCall(c.last, c.next)
			ts := tok.Token()
			if len(ts) > tokenMax || len(ts) == 0 {
				t.Errorf("%s: a token of %d bytes", c.what, len(ts))
			}
			if (tok.prefix != "") != c.hashes {
				t.Errorf("%s: the token %q carries hashes: %v, want %v", c.what, ts, tok.prefix != "", c.hashes)
			}
			back, err := NewCSIListingTokenFromTokenString(ts, list)
			if err != nil {
				t.Errorf("%s: the token %q is refused by the list it was issued for: %s", c.what, ts, err.Error())
				continue
			}
			if *back != *tok {
				t.Errorf("%s: the token %q reads back as %+v, was %+v", c.what, ts, *back, *tok)
			}
			keys := []string{"a", c.last, c.next, "vp_zz"}
			sort.Strings(keys)
			at := back.start(keys)
			if at >= len(keys) || keys[at] != c.next {
				t.Errorf("%s: the next call starts at %d of %q, want at %q", c.what, at, keys, c.next)
			}
		}
	}
}

// A token that carries hashes starts at the next entry, or after the last
// one returned, whichever is still there; when both are gone it starts
// after the beginning the two shared, so that no entry is missed.
func TestTokenWithHashesWhenItsEntriesAreGone(t *testing.T) {
	shared := strings.Repeat("vb_ORSXG5DJNZTQ", 9)[:125]
	keys := []string{"vp_a"}
	for _, tail := range []string{"000001", "000002", "000003", "000004", "000005"} {
		keys = append(keys, shared+tail)
	}
	keys = append(keys, "vp_z")
	sort.Strings(keys)
	first, _ := NewCSIListingTokenFromTokenString("", VolumesList())
	tok := first.nextCall(shared+"000002", shared+"000003")
	if tok.prefix == "" {
		t.Fatal("the token carries no hashes")
	}
	without := func(gone ...string) (left []string) {
		for _, k := range keys {
			keep := true
			for _, g := range gone {
				if k == g {
					keep = false
				}
			}
			if keep {
				left = append(left, k)
			}
		}
		return left
	}
	for _, c := range []struct {
		what string
		keys []string
		want string
	}{
		{"both are there", keys, shared + "000003"},
		{"the next one is gone", without(shared + "000003"), shared + "000004"},
		{"the last one returned is gone", without(shared + "000002"), shared + "000003"},
		// the entries that share the beginning come again: none is missed
		{"both are gone", without(shared+"000002", shared+"000003"), shared + "000001"},
	} {
		at := tok.start(c.keys)
		if at >= len(c.keys) || c.keys[at] != c.want {
			t.Errorf("%s: the call starts at %d, want at ...%s", c.what, at, c.want[len(c.want)-6:])
		}
	}
}

// A starting_token that does not parse, has another version or belongs to
// another list is refused, and no input makes the parse panic.
func TestTokenRefused(t *testing.T) {
	volumes := VolumesList()
	first, _ := NewCSIListingTokenFromTokenString("", volumes)
	good := first.nextCall("vp_a", "vp_b").Token()
	if _, err := NewCSIListingTokenFromTokenString(good, volumes); err != nil {
		t.Fatalf("the token %q is refused by its own list", good)
	}
	ofSnapshots, _ := NewCSIListingTokenFromTokenString("", SnapshotsList())
	ofVolumeA, _ := NewCSIListingTokenFromTokenString("", VolumeSnapshotsList("vp_a"))
	refused := map[string]string{
		"a token of the format before this one": "0_4037200794235010051_dnBfYQ==",
		"a page and a number":                   "3_17",
		"another version":                       "2" + good[1:],
		"the list of all snapshots":             ofSnapshots.nextCall("vp_a!sp_1", "vp_a!sp_2").Token(),
		"the list of one volume's snapshots":    ofVolumeA.nextCall("sp_1", "sp_2").Token(),
		"no bound":                              "1v.",
		"the head alone":                        "1v",
		"another mark after the head":           "1v:vp_a",
		"hashes and no beginning":               "1v~" + strings.Repeat("h", 2*hashLen),
		"hashes cut short":                      "1v~abc",
		"more than 128 bytes":                   "1v." + strings.Repeat("a", 126),
		"a space":                               "1v.vp a",
		"a control character":                   "1v.vp\x00a",
		"a byte that is not ASCII":              "1v.vp_ä",
		"text":                                  "next page please",
		"one character":                         "1",
	}
	for what, ts := range refused {
		tok, err := NewCSIListingTokenFromTokenString(ts, volumes)
		if err == nil || tok != nil {
			t.Errorf("%s: the token %q is taken", what, ts)
		}
	}
	// a token of one volume's snapshots is refused for another volume
	a := ofVolumeA.nextCall("sp_1", "sp_2").Token()
	if _, err := NewCSIListingTokenFromTokenString(a, VolumeSnapshotsList("vp_b")); err == nil {
		t.Errorf("the token %q of one volume's snapshots is taken for another volume", a)
	}
	if _, err := NewCSIListingTokenFromTokenString(a, VolumeSnapshotsList("vp_a")); err != nil {
		t.Errorf("the token %q is refused for its own volume", a)
	}

	// no input makes the parse, or the start it gives, panic
	rnd := rand.New(rand.NewSource(3))
	keys := []string{"vp_a", "vp_b", "vp_c"}
	for n := 0; n < 50000; n++ {
		b := make([]byte, rnd.Intn(140))
		for i := range b {
			switch rnd.Intn(4) {
			case 0:
				b[i] = byte(rnd.Intn(256))
			default:
				b[i] = "1vsw.~_-abcXYZ019"[rnd.Intn(17)]
			}
		}
		for _, list := range []ListID{volumes, SnapshotsList(), VolumeSnapshotsList("vp_a")} {
			if tok, err := NewCSIListingTokenFromTokenString(string(b), list); err == nil {
				if at := tok.start(keys); at < 0 || at > len(keys) {
					t.Fatalf("the token %q starts at %d", b, at)
				}
			}
		}
	}
}
