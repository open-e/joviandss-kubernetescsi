package driver

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	jrest "github.com/open-e/joviandss-kubernetescsi/pkg/rest"
)

type CSIListingToken struct {
	// BasedID is not universal, it depends on entry that we a listing through
	// for snapshot of particular volume that is sds
	// for volumes that is vds
	// for all snapshots that is vds_sds
	// and as name sugests it is all based64
	basedid string
	token   string
	page    int64
	dc      int64
}

func (t *CSIListingToken) Token() string {
	return t.token
}

func (t *CSIListingToken) BasedID() string {
	return t.basedid
}

func (t *CSIListingToken) DropBasedID() {
	t.basedid = ""
	t.token = fmt.Sprintf("%d_%d", t.page, t.dc)
}

func (t *CSIListingToken) Page() int64 {
	return t.page
}

func (t *CSIListingToken) DC() int64 {
	return t.dc
}

func (t *CSIListingToken) PageUp() {
	t.page += 1
}

//	Page()	int64
//	DC()	int64
//	ID()	string
//	String() string
//	StrToken() string
//	PageUP()

func NewCSIListingTokenFromBasedID(bid string, page int64, dc int64) (token CSIListingToken, err jrest.RestError) {

	token.basedid = bid

	if dc == 0 {
		token.dc = rand.Int63()
	}

	token.page = page
	token.token = fmt.Sprintf("%d_%d_%s", token.page, token.dc, token.basedid)
	return token, nil
}

func NewCSIListingToken() (token CSIListingToken) {

	token.dc = rand.Int63()
	token.page = 0
	token.basedid = ""
	token.token = fmt.Sprintf("%d_%d", token.page, token.dc)
	return token
}

func NewCSIListingTokenFromTokenString(ts string) (t *CSIListingToken, err jrest.RestError) {

	var ct CSIListingToken
	err = nil

	if len(ts) == 0 {
		ct.page = 0
		ct.dc = rand.Int63()
		ct.basedid = ""
		ct.token = ""
		return &ct, nil
	}

	parts := strings.Split(ts, "_")

	if len(parts) < 2 {
		return nil, jrest.GetError(jrest.RestErrorArgumentIncorrect, fmt.Sprintf("token %+v have bad format", t))
	}

	if i, err := strconv.ParseInt(parts[0], 10, 64); err != nil {
		return nil, jrest.GetError(jrest.RestErrorArgumentIncorrect, fmt.Sprintf("token %+v has bad page number %+v", t, err))
	} else {
		t.page = i
	}

	if i, err := strconv.ParseInt(parts[1], 10, 64); err != nil {
		return nil, jrest.GetError(jrest.RestErrorArgumentIncorrect, fmt.Sprintf("token %+v has bad dc number %+v", t, err))
	} else {
		t.dc = i
	}

	if len(parts) >= 3 {
		t.basedid = strings.Join(parts[2:], "_")
	}
	return &ct, nil
}

type snapshotToken struct {
	page int64
	dc   int64
	vid  string
	sid  string
}

func (st *snapshotToken) Page() int64 {
	return st.page
}

func (st *snapshotToken) DC() int64 {
	return st.dc
}

func (st *snapshotToken) ID() string {
	return fmt.Sprintf("%s_%s", st.vid, st.sid)
}

func (st *snapshotToken) String() string {

	return fmt.Sprintf("%d_%d_%s_%s", st.page, st.dc, st.vid, st.sid)
}

func NewSnapshotToken(page int64, dc int64, vid string, sid string) (t *snapshotToken) {
	t = &snapshotToken{page: page, dc: dc, vid: vid, sid: sid}
	return t
}

func NewSnapshotTokenFromStr(ts string) (t *snapshotToken, err jrest.RestError) {

	t = &snapshotToken{}
	err = nil

	parts := strings.Split(ts, "_")

	if len(parts) != 3 {
		return nil, jrest.GetError(jrest.RestErrorArgumentIncorrect, fmt.Sprintf("token %s have bad format", t))
	}

	if i, err := strconv.ParseInt(parts[0], 10, 64); err != nil {
		return nil, jrest.GetError(jrest.RestErrorArgumentIncorrect, fmt.Sprintf("token %s has bad page number %+v", t, err))
	} else {
		t.page = i
	}

	if i, err := strconv.ParseInt(parts[1], 10, 64); err != nil {
		return nil, jrest.GetError(jrest.RestErrorArgumentIncorrect, fmt.Sprintf("token %s has bad dc number %+v", t, err))
	} else {
		t.dc = i
	}

	t.vid = parts[2]
	t.sid = parts[3]
	return t, nil
}
