package controller

import (
	"context"
	"io"
	"testing"

	"github.com/container-storage-interface/spec/lib/go/csi"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	jdrvr "github.com/open-e/joviandss-kubernetescsi/pkg/driver"
	jrest "github.com/open-e/joviandss-kubernetescsi/pkg/rest"
)

// listDriver stands for a driver in the list calls of the controller: it
// keeps the arguments of the last list call and answers with what the test
// gives it. A call of any other method of the driver panics.
type listDriver struct {
	jdrvr.CSIDriver

	called string // the list method called last
	maxret int
	token  *jdrvr.CSIListingToken
	volume string // the volume of a list of one volume's snapshots

	entries interface{}
	next    *jdrvr.CSIListingToken
	err     jrest.RestError
}

func (d *listDriver) ListAllVolumes(ctx context.Context, pool string, maxret int, token *jdrvr.CSIListingToken) (interface{}, *jdrvr.CSIListingToken, jrest.RestError) {
	d.called, d.maxret, d.token = "ListAllVolumes", maxret, token
	return d.entries, d.next, d.err
}

func (d *listDriver) ListAllSnapshots(ctx context.Context, pool string, maxret int, token *jdrvr.CSIListingToken) (interface{}, *jdrvr.CSIListingToken, jrest.RestError) {
	d.called, d.maxret, d.token = "ListAllSnapshots", maxret, token
	return d.entries, d.next, d.err
}

func (d *listDriver) ListVolumeSnapshots(ctx context.Context, pool string, vid *jdrvr.VolumeDesc, maxret int, token *jdrvr.CSIListingToken) (interface{}, *jdrvr.CSIListingToken, jrest.RestError) {
	d.called, d.maxret, d.token, d.volume = "ListVolumeSnapshots", maxret, token, vid.VDS()
	return d.entries, d.next, d.err
}

func listController() (*ControllerPlugin, *listDriver) {
	l := log.New()
	l.SetOutput(io.Discard)
	d := &listDriver{}
	return &ControllerPlugin{l: l, d: d, pool: "Pool-0"}, d
}

// tokenOf makes a token as a call on list issues it.
func tokenOf(t *testing.T, ts string, list jdrvr.ListID) *jdrvr.CSIListingToken {
	t.Helper()
	token, err := jdrvr.NewCSIListingTokenFromTokenString(ts, list)
	if err != nil {
		t.Fatalf("the token %q: %s", ts, err.Error())
	}
	return token
}

// The tokens the controller must refuse on the volume list: text, the
// format before this one, another version and another list.
var foreignTokens = []string{"next page please", "0_4037200794235010051_dnBfYQ==", "3_17", "2v.vp_a", "1s.vp_a!sp_1", "1v", "1v.", "1v.vp a"}

// ListVolumes answers a negative max_entries with InvalidArgument and a
// starting_token it did not issue for this list with Aborted, without
// asking the appliance, and without a panic; next_token is set exactly
// when the driver gives a token.
func TestListVolumesArguments(t *testing.T) {
	cp, d := listController()

	_, err := cp.ListVolumes(context.Background(), &csi.ListVolumesRequest{MaxEntries: -1})
	if status.Code(err) != codes.InvalidArgument || d.called != "" {
		t.Errorf("max_entries -1: %v, driver called: %q", err, d.called)
	}
	for _, ts := range foreignTokens {
		_, err := cp.ListVolumes(context.Background(), &csi.ListVolumesRequest{StartingToken: ts})
		if status.Code(err) != codes.Aborted || d.called != "" {
			t.Errorf("starting_token %q: %v, driver called: %q", ts, err, d.called)
		}
	}

	d.entries = []jrest.ResourceVolume{{Name: "vp_a"}, {Name: "vp_b"}}
	d.next = tokenOf(t, "1v.vp_b", jdrvr.VolumesList())
	rsp, err := cp.ListVolumes(context.Background(), &csi.ListVolumesRequest{MaxEntries: 2, StartingToken: "1v.vp_0"})
	if err != nil {
		t.Fatal(err)
	}
	if d.called != "ListAllVolumes" || d.maxret != 2 || d.token == nil || d.token.Token() != "1v.vp_0" {
		t.Errorf("the driver got %s, max %d, token %v", d.called, d.maxret, d.token)
	}
	if len(rsp.Entries) != 2 || rsp.Entries[0].Volume.VolumeId != "vp_a" || rsp.NextToken != "1v.vp_b" {
		t.Errorf("the reply: %d entries, next_token %q", len(rsp.Entries), rsp.NextToken)
	}

	// no entry follows: no next_token
	d.next = nil
	rsp, err = cp.ListVolumes(context.Background(), &csi.ListVolumesRequest{})
	if err != nil || rsp.NextToken != "" || d.maxret != 0 {
		t.Errorf("the last call of a walk: next_token %q, error %v", rsp.GetNextToken(), err)
	}
}

// ListSnapshots does the same, for the list of all snapshots and for the
// list of one volume's snapshots, whose token no other volume takes.
func TestListSnapshotsArguments(t *testing.T) {
	cp, d := listController()

	for _, req := range []*csi.ListSnapshotsRequest{
		{MaxEntries: -1},
		{MaxEntries: -1, SourceVolumeId: "vp_a"},
	} {
		_, err := cp.ListSnapshots(context.Background(), req)
		if status.Code(err) != codes.InvalidArgument || d.called != "" {
			t.Errorf("max_entries -1: %v, driver called: %q", err, d.called)
		}
	}
	ofA := "1" + jdrvr.VolumeSnapshotsList("vp_a").Tag() + ".sp_1"
	for _, c := range []struct {
		what string
		req  *csi.ListSnapshotsRequest
	}{
		{"text", &csi.ListSnapshotsRequest{StartingToken: "next page please"}},
		{"the format before this one", &csi.ListSnapshotsRequest{StartingToken: "0_4037200794235010051_dnBfYQ=="}},
		{"a token of the volume list", &csi.ListSnapshotsRequest{StartingToken: "1v.vp_a"}},
		{"a token of one volume's snapshots", &csi.ListSnapshotsRequest{StartingToken: ofA}},
		{"a token of all snapshots, for one volume", &csi.ListSnapshotsRequest{StartingToken: "1s.vp_a!sp_1", SourceVolumeId: "vp_a"}},
		{"a token of another volume's snapshots", &csi.ListSnapshotsRequest{StartingToken: ofA, SourceVolumeId: "vp_b"}},
	} {
		_, err := cp.ListSnapshots(context.Background(), c.req)
		if status.Code(err) != codes.Aborted || d.called != "" {
			t.Errorf("%s: %v, driver called: %q", c.what, err, d.called)
		}
	}

	// all snapshots
	d.entries = []jrest.ResourceSnapshotShort{{Name: "sp_1", Parent: jrest.ResourceSnapshotParent{Name: "vp_a"}}}
	d.next = tokenOf(t, "1s.vp_a!sp_1", jdrvr.SnapshotsList())
	rsp, err := cp.ListSnapshots(context.Background(), &csi.ListSnapshotsRequest{MaxEntries: 1, StartingToken: "1s.vp_a"})
	if err != nil {
		t.Fatal(err)
	}
	if d.called != "ListAllSnapshots" || d.maxret != 1 || d.token.Token() != "1s.vp_a" {
		t.Errorf("the driver got %s, max %d, token %q", d.called, d.maxret, d.token.Token())
	}
	if len(rsp.Entries) != 1 || rsp.Entries[0].Snapshot.SourceVolumeId != "vp_a" || rsp.NextToken != "1s.vp_a!sp_1" {
		t.Errorf("the reply: %d entries, next_token %q", len(rsp.Entries), rsp.NextToken)
	}

	// the snapshots of one volume
	d.called, d.next = "", nil
	rsp, err = cp.ListSnapshots(context.Background(), &csi.ListSnapshotsRequest{StartingToken: ofA, SourceVolumeId: "vp_a"})
	if err != nil {
		t.Fatal(err)
	}
	if d.called != "ListVolumeSnapshots" || d.volume != "vp_a" || d.token.Token() != ofA || rsp.NextToken != "" {
		t.Errorf("the driver got %s for %s, token %q; next_token %q", d.called, d.volume, d.token.Token(), rsp.NextToken)
	}
}
