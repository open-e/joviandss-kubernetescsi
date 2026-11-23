package driver

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"sort"

	jrest "github.com/open-e/joviandss-kubernetescsi/pkg/rest"
)

const (
	tokenVersion = "1" // a token of another version is refused
	tokenMax     = 128 // bytes: CSI's limit for a string field
	hashLen      = 11  // base64url of the first 64 bits of a SHA-256
)

// shortHash names a string in hashLen ASCII characters.
func shortHash(s string) string {
	sum := sha256.Sum256([]byte(s))
	return base64.RawURLEncoding.EncodeToString(sum[:8])
}

// ListID names the list a token is issued for: the volumes, all
// snapshots, or the snapshots of one volume, for which a hash of the
// volume's name stands, so that the name does not take room in the token.
type ListID struct{ tag string }

// Tag returns the name of the list as a token carries it.
func (l ListID) Tag() string { return l.tag }

func VolumesList() ListID                      { return ListID{"v"} }
func SnapshotsList() ListID                    { return ListID{"s"} }
func VolumeSnapshotsList(volume string) ListID { return ListID{"w" + shortHash(volume)[:8]} }

// CSIListingToken is where a CSI list call resumes: after bound, or, when
// the bound did not fit, by the hashes of the last key returned and of the
// next, with prefix, the bound's first bytes. It names no appliance page.
type CSIListingToken struct {
	list               ListID
	bound              string
	prefix             string
	hashLast, hashNext string // with prefix only
}

// bound returns the shortest string b with last <= b < next, for
// last < next: the next call starts after it. It holds bytes of the two
// keys, or one of them raised by one.
func bound(last, next string) string {
	i := 0
	for i < len(last) && i < len(next) && last[i] == next[i] {
		i++
	}
	switch {
	case i == len(last): // last is the beginning of next
		return last
	case last[i]+1 < next[i]: // a byte between the two fits
		return last[:i] + string([]byte{last[i] + 1})
	case i+1 < len(next): // next[i] follows last[i]: the beginning of next
		return next[:i+1]
	}
	// next ends on the byte that follows last[i]: the bound keeps last[i],
	// and is made greater than last at the first later byte that has room
	for j := i + 1; j < len(last); j++ {
		if last[j] < 0xFF {
			return last[:j] + string([]byte{last[j] + 1})
		}
	}
	return last
}

// nextCall is the token of the call after one that returned up to the key
// last, when next is the key of the entry that follows it.
func (t *CSIListingToken) nextCall(last, next string) *CSIListingToken {
	n := &CSIListingToken{list: t.list, bound: bound(last, next)}
	room := tokenMax - len(tokenVersion) - len(t.list.tag) - 1
	if len(n.bound) > room {
		// the two keys share a beginning too long for the token
		n.prefix = n.bound[:room-2*hashLen]
		n.hashLast, n.hashNext = shortHash(last), shortHash(next)
		n.bound = ""
	}
	return n
}

// start is the place, in keys sorted in byte order, of the first entry
// the call returns.
func (t *CSIListingToken) start(keys []string) int {
	after := func(b string) int {
		return sort.Search(len(keys), func(i int) bool { return keys[i] > b })
	}
	switch {
	case t == nil:
		return 0 // no token: the whole list
	case t.bound != "":
		return after(t.bound)
	case t.prefix == "":
		return 0 // the first call of a walk
	}
	for i, k := range keys {
		if shortHash(k) == t.hashNext {
			return i
		}
	}
	for i := len(keys) - 1; i >= 0; i-- {
		if shortHash(keys[i]) == t.hashLast {
			return i + 1
		}
	}
	// both entries are gone: start after the beginning they shared, so that
	// no entry is missed; the entries that share it come again
	return after(t.prefix)
}

// Token writes the next_token, ASCII of at most tokenMax bytes: the
// version, the list, then "." and the bound, or "~", the two hashes and
// the prefix.
func (t *CSIListingToken) Token() string {
	head := tokenVersion + t.list.tag
	if t.prefix != "" {
		return head + "~" + t.hashLast + t.hashNext + t.prefix
	}
	return head + "." + t.bound
}

// NewCSIListingTokenFromTokenString reads the starting_token of a call on
// list, "" for the first call of a walk. A token that does not parse, has
// another version or was issued for another list is refused. No input makes
// it panic.
func NewCSIListingTokenFromTokenString(ts string, list ListID) (*CSIListingToken, jrest.RestError) {
	t := &CSIListingToken{list: list}
	if ts == "" {
		return t, nil
	}
	refused := jrest.GetError(jrest.RestErrorArgumentIncorrect,
		fmt.Sprintf("token %q was not issued for this list", ts))
	head := tokenVersion + list.tag
	if len(ts) > tokenMax || len(ts) < len(head)+2 || ts[:len(head)] != head {
		return nil, refused
	}
	for i := 0; i < len(ts); i++ {
		if ts[i] < '!' || ts[i] > '~' { // ASCII, no space, no control
			return nil, refused
		}
	}
	rest := ts[len(head)+1:]
	switch ts[len(head)] {
	case '.':
		t.bound = rest
	case '~':
		if len(rest) <= 2*hashLen {
			return nil, refused
		}
		t.hashLast, t.hashNext, t.prefix = rest[:hashLen], rest[hashLen:2*hashLen], rest[2*hashLen:]
	default:
		return nil, refused
	}
	return t, nil
}
