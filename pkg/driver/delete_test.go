package driver

import (
	"reflect"
	"strings"
	"testing"

	jrest "github.com/open-e/joviandss-kubernetescsi/pkg/rest"
)

// Replies of the appliance, as its recordings show them.
var (
	zfsCmdError = "zfslib.wrap.zfs.ZfsCmdError"

	noContent = func(fakeRequest) (int, interface{}) { return 204, nil }

	hasChildren = func(fakeRequest) (int, interface{}) {
		return 500, appError(zfsCmdError, 1000,
			"cannot destroy 'Pool-0/vp_a': volume has children\nuse '-r' to destroy the following datasets:\nPool-0/vp_a@sp_1")
	}
	zvolHasClones = func(fakeRequest) (int, interface{}) {
		return 500, appError("opene.storage.zfs.zfs.ZfsOeError", -1,
			"In order to delete a zvol, you must delete all of its clones first.")
	}
	snapshotHasClones = func(fakeRequest) (int, interface{}) {
		return 500, appError(zfsCmdError, 1000,
			"cannot destroy 'Pool-0/vp_a@sp_1': snapshot has dependent clones\nuse '-R' to destroy the following datasets:\nPool-0/sp_1")
	}
	notFound = func(fakeRequest) (int, interface{}) {
		return 500, appError("zfslib.zfsapi.resources.ZfsResourceError", 1,
			"Zfs resource: Pool-0/vp_a not found in this collection.")
	}
	somethingNew = func(fakeRequest) (int, interface{}) {
		return 500, appError("builtins.RuntimeError", -1, "something no recording shows")
	}
)

func volumeA(t *testing.T) *VolumeDesc {
	t.Helper()
	vd, err := NewVolumeDescFromVDS("vp_a")
	if err != nil {
		t.Fatal(err)
	}
	return vd
}

// noForcedUmount fails the test when a request carried force_umount true.
func noForcedUmount(t *testing.T, reqs []fakeRequest) {
	t.Helper()
	for _, r := range reqs {
		if v, sent := r.Body["force_umount"]; sent && v != false {
			t.Errorf("%s carried force_umount %v", r.line(), v)
		}
	}
}

func lines(reqs []fakeRequest) (out []string) {
	for _, r := range reqs {
		out = append(out, r.line())
	}
	return out
}

const (
	volA      = pool0 + "/volumes/vp_a"
	volASnaps = volA + "/snapshots"
	nasA      = pool0 + "/nas-volumes/vp_a"
	nasASnaps = nasA + "/snapshots"
	firstPage = "?page=0&per_page=100"
)

// An iSCSI volume without snapshots is deleted with one request, which
// carries force_umount false; a volume that is gone is no error.
func TestISCSIDeleteVolume(t *testing.T) {
	app := newFakeAppliance(t)
	d := app.iscsi()
	app.on("DELETE", volA, noContent)
	if err := d.DeleteVolume(app.ctx, "Pool-0", volumeA(t)); err != nil {
		t.Fatal(err.Error())
	}
	reqs := app.sent()
	if got := lines(reqs); !reflect.DeepEqual(got, []string{"DELETE " + volA}) {
		t.Errorf("the requests: %q", got)
	}
	if v, sent := reqs[0].Body["force_umount"]; !sent || v != false {
		t.Errorf("the delete carried force_umount %v, sent: %v; want false", v, sent)
	}

	app.on("DELETE", volA, notFound)
	if err := d.DeleteVolume(app.ctx, "Pool-0", volumeA(t)); err != nil {
		t.Errorf("a volume that is gone: %s", err.Error())
	}
}

// A zvol whose delete is refused for a leftover intermediate snapshot is
// deleted after the cleanup, for "volume has children" as for "delete its
// clones first": the snapshot's clones are read, it has none, it is
// deleted, and the volume's delete is sent again.
func TestISCSIDeleteVolumeAfterTheCleanup(t *testing.T) {
	for what, refusal := range map[string]func(fakeRequest) (int, interface{}){
		"volume has children": hasChildren, "delete its clones first": zvolHasClones,
	} {
		app := newFakeAppliance(t)
		d := app.iscsi()
		deletes := 0
		app.on("DELETE", volA, func(r fakeRequest) (int, interface{}) {
			deletes++
			if deletes == 1 {
				return refusal(r)
			}
			return 204, nil
		})
		app.on("GET", volASnaps, pages(func() []object {
			return []object{snapshotObject("vp_a", "vp_clone", jrest.SnapshotParentVolume)}
		}))
		app.on("GET", volASnaps+"/vp_clone/clones", func(fakeRequest) (int, interface{}) { return 200, data([]object{}) })
		app.on("DELETE", volASnaps+"/vp_clone", noContent)

		if err := d.DeleteVolume(app.ctx, "Pool-0", volumeA(t)); err != nil {
			t.Errorf("%s: %s", what, err.Error())
		}
		reqs := app.sent()
		want := []string{
			"DELETE " + volA,
			"GET " + volASnaps + firstPage,
			"GET " + volASnaps + "/vp_clone/clones",
			"DELETE " + volASnaps + "/vp_clone",
			"DELETE " + volA,
		}
		if got := lines(reqs); !reflect.DeepEqual(got, want) {
			t.Errorf("%s: the requests %q, want %q", what, got, want)
		}
		noForcedUmount(t, reqs)
	}
}

// A zvol that still has snapshots after the cleanup is not deleted: the
// error has the code the controller answers FailedPrecondition and names
// the CSI snapshots, the volumes cloned from it and the snapshots the
// plugin did not make. An intermediate snapshot that has a clone stays.
func TestISCSIDeleteVolumeWithSnapshots(t *testing.T) {
	app := newFakeAppliance(t)
	d := app.iscsi()
	app.on("DELETE", volA, hasChildren)
	app.on("GET", volASnaps, pages(func() []object {
		return []object{
			snapshotObject("vp_a", "sp_1", jrest.SnapshotParentVolume),
			snapshotObject("vp_a", "vp_clone", jrest.SnapshotParentVolume),
			snapshotObject("vp_a", "auto-0001", jrest.SnapshotParentVolume),
		}
	}))
	app.on("GET", volASnaps+"/vp_clone/clones", func(fakeRequest) (int, interface{}) {
		return 200, data([]object{{"name": "vp_clone", "full_name": "Pool-0/vp_clone", "is_clone": true}})
	})

	err := d.DeleteVolume(app.ctx, "Pool-0", volumeA(t))
	if err == nil || err.GetCode() != jrest.RestErrorResourceBusy {
		t.Fatalf("a volume with snapshots: error %v, want the code %d", err, jrest.RestErrorResourceBusy)
	}
	for _, name := range []string{"sp_1", "vp_clone", "auto-0001"} {
		if !strings.Contains(err.Error(), name) {
			t.Errorf("the error %q does not name %s", err.Error(), name)
		}
	}
	want := []string{"DELETE " + volA, "GET " + volASnaps + firstPage, "GET " + volASnaps + "/vp_clone/clones"}
	if got := app.seen(); !reflect.DeepEqual(got, want) {
		t.Errorf("the requests %q, want %q: no snapshot deleted, the volume's delete not sent again", got, want)
	}
}

// When the clones of an intermediate snapshot cannot be read, nothing is
// deleted.
func TestISCSICleanupDeletesNothingWithoutTheClones(t *testing.T) {
	app := newFakeAppliance(t)
	d := app.iscsi()
	app.on("DELETE", volA, hasChildren)
	app.on("GET", volASnaps, pages(func() []object {
		return []object{snapshotObject("vp_a", "vp_clone", jrest.SnapshotParentVolume)}
	}))
	app.on("GET", volASnaps+"/vp_clone/clones", somethingNew)

	if err := d.DeleteVolume(app.ctx, "Pool-0", volumeA(t)); err == nil {
		t.Fatal("the delete succeeded")
	}
	want := []string{"DELETE " + volA, "GET " + volASnaps + firstPage, "GET " + volASnaps + "/vp_clone/clones"}
	if got := app.seen(); !reflect.DeepEqual(got, want) {
		t.Errorf("the requests %q, want %q", got, want)
	}
}

// The appliance deletes a NAS volume together with its snapshots, so the
// NFS driver looks first: a NAS volume that has a snapshot, a CSI snapshot
// or one the plugin did not make, is not deleted, and no delete is sent.
func TestNFSDeleteVolumeWithSnapshotsSendsNoDelete(t *testing.T) {
	for _, snapshot := range []string{"sp_1", "auto-0001"} {
		app := newFakeAppliance(t)
		d := app.nfs()
		app.on("GET", nasASnaps, pages(func() []object {
			return []object{snapshotObject("vp_a", snapshot, jrest.SnapshotParentNASVolume)}
		}))
		err := d.DeleteVolume(app.ctx, "Pool-0", volumeA(t))
		if err == nil || err.GetCode() != jrest.RestErrorResourceBusy || !strings.Contains(err.Error(), snapshot) {
			t.Errorf("a NAS volume with the snapshot %s: error %v, want one of the code %d that names it",
				snapshot, err, jrest.RestErrorResourceBusy)
		}
		if got := app.seen(); !reflect.DeepEqual(got, []string{"GET " + nasASnaps + firstPage}) {
			t.Errorf("a NAS volume with the snapshot %s: the requests %q, want the read of its snapshots alone", snapshot, got)
		}
	}
}

// An intermediate snapshot of a NAS volume is deleted only once the
// appliance has answered that it has no clones, and the volume after it;
// with a clone, nothing is deleted.
func TestNFSDeleteVolumeAndItsIntermediateSnapshot(t *testing.T) {
	app := newFakeAppliance(t)
	d := app.nfs()
	clones := []object{}
	app.on("GET", nasASnaps, pages(func() []object {
		return []object{snapshotObject("vp_a", "vp_clone", jrest.SnapshotParentNASVolume)}
	}))
	app.on("GET", nasASnaps+"/vp_clone/clones", func(fakeRequest) (int, interface{}) { return 200, data(clones) })
	app.on("DELETE", nasASnaps+"/vp_clone", noContent)
	app.on("DELETE", nasA, noContent)

	if err := d.DeleteVolume(app.ctx, "Pool-0", volumeA(t)); err != nil {
		t.Fatal(err.Error())
	}
	want := []string{
		"GET " + nasASnaps + firstPage,
		"GET " + nasASnaps + "/vp_clone/clones",
		"DELETE " + nasASnaps + "/vp_clone",
		"DELETE " + nasA,
	}
	if got := app.seen(); !reflect.DeepEqual(got, want) {
		t.Errorf("the requests %q, want %q", got, want)
	}

	clones = []object{{"name": "vp_clone"}}
	err := d.DeleteVolume(app.ctx, "Pool-0", volumeA(t))
	if err == nil || err.GetCode() != jrest.RestErrorResourceBusy || !strings.Contains(err.Error(), "vp_clone") {
		t.Errorf("a NAS volume whose intermediate snapshot has a clone: error %v", err)
	}
	want = []string{"GET " + nasASnaps + firstPage, "GET " + nasASnaps + "/vp_clone/clones"}
	if got := app.seen(); !reflect.DeepEqual(got, want) {
		t.Errorf("the requests %q, want %q: nothing deleted", got, want)
	}

	// the clones cannot be read: nothing is deleted
	app.on("GET", nasASnaps+"/vp_clone/clones", somethingNew)
	if err := d.DeleteVolume(app.ctx, "Pool-0", volumeA(t)); err == nil {
		t.Error("the delete succeeded although the clones could not be read")
	}
	if got := app.seen(); !reflect.DeepEqual(got, want) {
		t.Errorf("the requests %q, want %q: nothing deleted", got, want)
	}
}

// A NAS volume without snapshots is deleted; one that is gone is no error;
// and when its snapshots cannot be read, no delete is sent.
func TestNFSDeleteVolume(t *testing.T) {
	app := newFakeAppliance(t)
	d := app.nfs()
	app.on("GET", nasASnaps, pages(func() []object { return nil }))
	app.on("DELETE", nasA, noContent)
	if err := d.DeleteVolume(app.ctx, "Pool-0", volumeA(t)); err != nil {
		t.Fatal(err.Error())
	}
	if got := app.seen(); !reflect.DeepEqual(got, []string{"GET " + nasASnaps + firstPage, "DELETE " + nasA}) {
		t.Errorf("the requests %q", got)
	}

	// gone: the list of its snapshots answers 404, and so does the delete
	app.on("GET", nasASnaps, func(fakeRequest) (int, interface{}) {
		return 404, appError("opene.exceptions.ItemNotFoundError", -1, "Resource Pool-0/vp_a not found.")
	})
	app.on("DELETE", nasA, notFound)
	if err := d.DeleteVolume(app.ctx, "Pool-0", volumeA(t)); err != nil {
		t.Errorf("a NAS volume that is gone: %s", err.Error())
	}

	// the snapshots cannot be read: the volume may have some
	app.on("GET", nasASnaps, somethingNew)
	app.seen()
	if err := d.DeleteVolume(app.ctx, "Pool-0", volumeA(t)); err == nil {
		t.Error("the delete succeeded although the snapshots could not be read")
	}
	if got := app.seen(); !reflect.DeepEqual(got, []string{"GET " + nasASnaps + firstPage}) {
		t.Errorf("the requests %q, want the read of the snapshots alone", got)
	}
}

// When a zvol snapshot's delete is refused for its clones, the driver
// deletes the clones that carry a snapshot's name. An error of that delete
// keeps its meaning: a clone that is gone is nothing in the way, and a
// failure the plugin does not know is not "busy".
func TestISCSIDeleteSnapshotCloneDeleteErrors(t *testing.T) {
	vd := volumeA(t)
	sd, err := NewSnapshotDescFromSDS(vd, "sp_1")
	if err != nil {
		t.Fatal(err)
	}
	snap := volASnaps + "/sp_1"
	for _, c := range []struct {
		what        string
		cloneDelete func(fakeRequest) (int, interface{})
		code        int
		cantDelete  bool
	}{
		{"the clone is gone already", notFound, jrest.RestErrorResourceBusy, false},
		{"the clone is deleted", noContent, jrest.RestErrorResourceBusy, false},
		{"the clone has snapshots of its own", hasChildren, jrest.RestErrorResourceBusy, true},
		{"a failure no recording shows", somethingNew, jrest.RestErrorFailureUnknown, false},
	} {
		app := newFakeAppliance(t)
		d := app.iscsi()
		app.on("DELETE", snap, snapshotHasClones)
		app.on("GET", snap+"/clones", func(fakeRequest) (int, interface{}) {
			return 200, data([]object{{"name": "sp_1", "full_name": "Pool-0/sp_1", "is_clone": true}})
		})
		app.on("DELETE", snap+"/clones/sp_1", c.cloneDelete)

		rerr := d.DeleteSnapshot(app.ctx, "Pool-0", vd, sd)
		if rerr == nil || rerr.GetCode() != c.code {
			t.Errorf("%s: error %v, want the code %d", c.what, rerr, c.code)
			continue
		}
		if strings.Contains(rerr.Error(), "cant be deleted") != c.cantDelete {
			t.Errorf("%s: the error is %q", c.what, rerr.Error())
		}
		reqs := app.sent()
		want := []string{"DELETE " + snap, "GET " + snap + "/clones", "DELETE " + snap + "/clones/sp_1"}
		if got := lines(reqs); !reflect.DeepEqual(got, want) {
			t.Errorf("%s: the requests %q, want %q", c.what, got, want)
		}
		noForcedUmount(t, reqs)
	}
}

// An NFS create takes a share that exists already as its own only when the
// share's path is the one it sends; a share with another path fails the
// create.
func TestNFSCreateVolumeWithAnExistingShare(t *testing.T) {
	app := newFakeAppliance(t)
	d := app.nfs()
	nvd, err := NewVolumeDescFromName("a")
	if err != nil {
		t.Fatal(err)
	}
	path := "Pool-0/vp_a/vp_a"
	app.on("POST", pool0+"/nas-volumes", func(fakeRequest) (int, interface{}) {
		return 500, appError("zfslib.zfsapi.resources.ZfsResourceError", 5, "Resource Pool-0/vp_a already exists.")
	})
	app.on("POST", "/api/v4/shares", func(fakeRequest) (int, interface{}) {
		return 409, appError("", 1, "Share vp_a already exists.")
	})
	app.on("GET", "/api/v4/shares/vp_a", func(fakeRequest) (int, interface{}) {
		return 200, data(object{"name": "vp_a", "path": path})
	})

	if err := d.CreateVolume(app.ctx, "Pool-0", nvd, 0, 0); err != nil {
		t.Fatalf("a repeated create: %s", err.Error())
	}
	reqs := app.sent()
	want := []string{"POST " + pool0 + "/nas-volumes", "POST /api/v4/shares", "GET /api/v4/shares/vp_a"}
	if got := lines(reqs); !reflect.DeepEqual(got, want) {
		t.Errorf("the requests %q, want %q", got, want)
	}
	if reqs[1].Body["path"] != path || reqs[1].Body["name"] != "vp_a" {
		t.Errorf("the share create sent the name %v and the path %v", reqs[1].Body["name"], reqs[1].Body["path"])
	}

	path = "Pool-0/another/dir"
	rerr := d.CreateVolume(app.ctx, "Pool-0", nvd, 0, 0)
	if rerr == nil || rerr.GetCode() != jrest.RestErrorResourceExists ||
		!strings.Contains(rerr.Error(), "Pool-0/another/dir") || !strings.Contains(rerr.Error(), "Pool-0/vp_a/vp_a") {
		t.Errorf("a share of that name with another path: error %v", rerr)
	}

	// the share cannot be read: the create fails with that error
	app.on("GET", "/api/v4/shares/vp_a", somethingNew)
	if rerr := d.CreateVolume(app.ctx, "Pool-0", nvd, 0, 0); rerr == nil || rerr.GetCode() != jrest.RestErrorFailureUnknown {
		t.Errorf("a share that cannot be read: error %v", rerr)
	}
}
