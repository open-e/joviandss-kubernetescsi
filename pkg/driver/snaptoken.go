package driver

import (
	"fmt"
	"strconv"
	"strings"

	jrest "joviandss-kubernetescsi/pkg/rest"
)

type snapshotToken struct {
	page	int64
	dc	int64
	vid	string
	sid	string
}

func NewSnapshotToken(page int64, dc int64, vid string, sid string) (t *snapshotToken) {
	t = &snapshotToken{page: page, dc: dc, vid: vid, sid: sid }
	return t
}

func NewSnapshotTokenFromStr(ts string) (t *snapshotToken, err jrest.RestError) {
	
	t = &snapshotToken{}
	err = nil

	parts := strings.Split(ts, "_")

	if len(parts) != 3 {
		return nil, jrest.GetError(jrest.RestErrorArgumentIncorrect, fmt.Sprintf("token %s have bad format", t))
	}

	if i, err := strconv.ParseInt(parts[0] , 10, 64); err != nil {
		return nil, jrest.GetError(jrest.RestErrorArgumentIncorrect, fmt.Sprintf("token %s has bad page number %+v", t, err))
	} else {
		t.page = i
	}

	if i, err := strconv.ParseInt(parts[1] , 10, 64); err != nil {
		return nil, jrest.GetError(jrest.RestErrorArgumentIncorrect, fmt.Sprintf("token %s has bad dc number %+v", t, err))
	} else {
		t.dc = i
	}

	t.vid	= parts[2]
	t.sid	= parts[3]
	return t, nil
}

func (st *snapshotToken) String() string {

	return fmt.Sprintf("%d_%d_%s_%s", st.page, st.dc, st.vid, st.sid)
}
