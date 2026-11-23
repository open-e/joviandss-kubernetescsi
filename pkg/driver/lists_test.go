package driver

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	jrest "github.com/open-e/joviandss-kubernetescsi/pkg/rest"
)

const pool0 = "/api/v4/pools/Pool-0"

// threePages is what a read of a list of 201 to 299 entries sends: the
// whole list, from page 0, with per_page=100 and no _dc.
func threePages(path string) []string {
	return []string{
		"GET " + path + "?page=0&per_page=100",
		"GET " + path + "?page=1&per_page=100",
		"GET " + path + "?page=2&per_page=100",
	}
}

// ListAllVolumes reads the whole volume list once on every call, whatever
// the token, leaves out the volumes the plugin did not make, with a
// warning, and returns max_entries of its own whenever that many follow.
func TestListAllVolumes(t *testing.T) {
	app := newFakeAppliance(t)
	var list []object
	for i := 0; i < 250; i++ {
		name := fmt.Sprintf("vp_pvc-%04d", i)
		if i%10 == 3 {
			name = fmt.Sprintf("backup-%04d", i) // made by hand
		}
		list = append(list, volumeObject(name))
	}
	app.on("GET", pool0+"/volumes", pages(func() []object { return list }))

	for _, d := range []CSIDriver{app.iscsi()} {
		var names []string
		ts, calls := "", 0
		for {
			token, err := NewCSIListingTokenFromTokenString(ts, VolumesList())
			if err != nil {
				t.Fatal(err.Error())
			}
			app.warned = nil
			app.seen()
			vols, next, err := d.ListAllVolumes(app.ctx, "Pool-0", 60, token)
			if err != nil {
				t.Fatal(err.Error())
			}
			calls++
			if got := app.seen(); !reflect.DeepEqual(got, threePages(pool0+"/volumes")) {
				t.Fatalf("call %d sent %q", calls, got)
			}
			if len(app.warned) != 25 || !strings.Contains(app.warned[0], "backup-0003") {
				t.Errorf("call %d: %d warnings about names the plugin did not make, want 25", calls, len(app.warned))
			}
			page := vols.([]jrest.ResourceVolume)
			for _, v := range page {
				names = append(names, v.Name)
			}
			if next == nil {
				break
			}
			if len(page) != 60 {
				t.Fatalf("call %d returned %d volumes and a token, want 60", calls, len(page))
			}
			ts = next.Token()
		}
		if len(names) != 225 || calls != 4 {
			t.Errorf("%d volumes in %d calls, want the plugin's 225 in 4", len(names), calls)
		}
		for i := 1; i < len(names); i++ {
			if names[i-1] >= names[i] {
				t.Fatalf("the volumes are not in the order of their names: %s before %s", names[i-1], names[i])
			}
		}
	}
}

// The NFS driver lists the NAS volumes the same way.
func TestListAllNASVolumes(t *testing.T) {
	app := newFakeAppliance(t)
	list := []object{nasVolumeObject("vp_b"), nasVolumeObject("by-hand"), nasVolumeObject("vp_a"), nasVolumeObject("vp_c")}
	app.on("GET", pool0+"/nas-volumes", pages(func() []object { return list }))
	d := app.nfs()

	token, _ := NewCSIListingTokenFromTokenString("", VolumesList())
	vols, next, err := d.ListAllVolumes(app.ctx, "Pool-0", 2, token)
	if err != nil {
		t.Fatal(err.Error())
	}
	if got := app.seen(); !reflect.DeepEqual(got, []string{"GET " + pool0 + "/nas-volumes?page=0&per_page=100"}) {
		t.Errorf("the requests of a list of four: %q", got)
	}
	page := vols.([]jrest.ResourceNASVolume)
	if len(page) != 2 || page[0].Name != "vp_a" || page[1].Name != "vp_b" || next == nil {
		t.Fatalf("the first call: %d volumes, token %v", len(page), next)
	}
	token, err = NewCSIListingTokenFromTokenString(next.Token(), VolumesList())
	if err != nil {
		t.Fatal(err.Error())
	}
	vols, next, err = d.ListAllVolumes(app.ctx, "Pool-0", 2, token)
	page = vols.([]jrest.ResourceNASVolume)
	if err != nil || len(page) != 1 || page[0].Name != "vp_c" || next != nil {
		t.Errorf("the second call: %d volumes, token %v, error %v", len(page), next, err)
	}
}

// Both lists of all snapshots return the snapshots of zvols and of NAS
// volumes together. Each driver drops those of the other kind, without a
// warning, and keeps a parent of another kind; the order is by the parent's
// name and then by the snapshot's.
func TestListAllSnapshotsDropsTheOtherKind(t *testing.T) {
	app := newFakeAppliance(t)
	list := []object{
		snapshotObject("vp_b", "sp_1", jrest.SnapshotParentVolume),
		snapshotObject("vp_n", "sp_1", jrest.SnapshotParentNASVolume),
		snapshotObject("vp_a", "sp_2", jrest.SnapshotParentVolume),
		snapshotObject("vp_a", "sp_1", jrest.SnapshotParentVolume),
		snapshotObject("vp_n", "sp_0", jrest.SnapshotParentNASVolume),
		snapshotObject("vp_x", "sp_9", 7),                               // a kind no recording shows
		snapshotObject("vp_a", "vp_clone", jrest.SnapshotParentVolume),  // an intermediate snapshot
		snapshotObject("by-hand", "sp_1", jrest.SnapshotParentVolume),   // a parent the plugin did not make
		snapshotObject("vp_a", "auto-0001", jrest.SnapshotParentVolume), // a snapshot the plugin did not make
		snapshotObject("vp_n", "auto-0002", jrest.SnapshotParentNASVolume),
	}
	app.on("GET", pool0+"/volumes/snapshots", pages(func() []object { return list }))
	app.on("GET", pool0+"/nas-volumes/snapshots", pages(func() []object { return list }))

	token, _ := NewCSIListingTokenFromTokenString("", SnapshotsList())
	snaps, next, err := app.iscsi().ListAllSnapshots(app.ctx, "Pool-0", 0, token)
	if err != nil || next != nil {
		t.Fatalf("the iSCSI list: token %v, error %v", next, err)
	}
	var got []string
	for _, s := range snaps.([]jrest.ResourceSnapshotShort) {
		got = append(got, s.Parent.Name+"@"+s.Name)
	}
	if want := []string{"vp_a@sp_1", "vp_a@sp_2", "vp_b@sp_1", "vp_x@sp_9"}; !reflect.DeepEqual(got, want) {
		t.Errorf("the iSCSI list of all snapshots: %q, want %q", got, want)
	}
	if got := app.seen(); !reflect.DeepEqual(got, []string{"GET " + pool0 + "/volumes/snapshots?page=0&per_page=100"}) {
		t.Errorf("the iSCSI list sent %q", got)
	}
	// three names the plugin did not make among the zvol snapshots, and
	// none of the NAS snapshots, which are dropped before the name check
	if len(app.warned) != 3 {
		t.Errorf("the iSCSI list warned %d times, want 3: %q", len(app.warned), app.warned)
	}

	app.warned = nil
	snaps, next, err = app.nfs().ListAllSnapshots(app.ctx, "Pool-0", 0, token)
	if err != nil || next != nil {
		t.Fatalf("the NFS list: token %v, error %v", next, err)
	}
	got = nil
	for _, s := range snaps.([]jrest.ResourceNASSnapshotShort) {
		got = append(got, s.Parent.Name+"@"+s.Name)
	}
	if want := []string{"vp_n@sp_0", "vp_n@sp_1", "vp_x@sp_9"}; !reflect.DeepEqual(got, want) {
		t.Errorf("the NFS list of all snapshots: %q, want %q", got, want)
	}
	if got := app.seen(); !reflect.DeepEqual(got, []string{"GET " + pool0 + "/nas-volumes/snapshots?page=0&per_page=100"}) {
		t.Errorf("the NFS list sent %q", got)
	}
	if len(app.warned) != 1 {
		t.Errorf("the NFS list warned %d times, want 1: %q", len(app.warned), app.warned)
	}
}

// The list of one volume's snapshots asks the path of that volume's
// snapshots, resumes after the last name returned, and gives each snapshot
// a place of its own: two snapshots of one volume never share a token.
func TestListVolumeSnapshots(t *testing.T) {
	app := newFakeAppliance(t)
	vd, err := NewVolumeDescFromVDS("vp_a")
	if err != nil {
		t.Fatal(err)
	}
	var list []object
	for i := 0; i < 7; i++ {
		list = append(list, snapshotObject("vp_a", fmt.Sprintf("sp_%d", 6-i), jrest.SnapshotParentVolume))
	}
	list = append(list, snapshotObject("vp_a", "vp_clone", jrest.SnapshotParentVolume))
	app.on("GET", pool0+"/volumes/vp_a/snapshots", pages(func() []object { return list }))
	d := app.iscsi()

	var got, tokens []string
	ts := ""
	for {
		token, err := NewCSIListingTokenFromTokenString(ts, VolumeSnapshotsList(vd.VDS()))
		if err != nil {
			t.Fatal(err.Error())
		}
		snaps, next, err := d.ListVolumeSnapshots(app.ctx, "Pool-0", vd, 3, token)
		if err != nil {
			t.Fatal(err.Error())
		}
		for _, s := range snaps.([]jrest.ResourceSnapshotShort) {
			got = append(got, s.Name)
		}
		if next == nil {
			break
		}
		ts = next.Token()
		tokens = append(tokens, ts)
	}
	if want := []string{"sp_0", "sp_1", "sp_2", "sp_3", "sp_4", "sp_5", "sp_6"}; !reflect.DeepEqual(got, want) {
		t.Errorf("the snapshots of the volume: %q, want %q", got, want)
	}
	if len(tokens) != 2 || tokens[0] == tokens[1] {
		t.Errorf("the tokens of the walk: %q, want two that differ", tokens)
	}
	for _, line := range app.seen() {
		if line != "GET "+pool0+"/volumes/vp_a/snapshots?page=0&per_page=100" {
			t.Errorf("the list sent %s", line)
		}
	}

	// a volume that is gone: the error of the appliance is the error of the call
	app.on("GET", pool0+"/volumes/vp_a/snapshots", func(fakeRequest) (int, interface{}) {
		return 404, appError("opene.exceptions.ItemNotFoundError", -1, "Resource Pool-0/vp_a not found.")
	})
	token, _ := NewCSIListingTokenFromTokenString("", VolumeSnapshotsList(vd.VDS()))
	snaps, next, rerr := d.ListVolumeSnapshots(app.ctx, "Pool-0", vd, 3, token)
	if rerr == nil || rerr.GetCode() != jrest.RestErrorResourceDNE || snaps != nil || next != nil {
		t.Errorf("the snapshots of a volume that is gone: %v, token %v, error %v", snaps, next, rerr)
	}
}

// A name that two pages of one read returned stays once when both entries
// are one object, the entry read later, and twice when they are two.
func TestListMergesTheEntriesOfOneName(t *testing.T) {
	app := newFakeAppliance(t)
	older := volumeObject("vp_a")
	older["volsize"] = "1073741824"
	newer := volumeObject("vp_a")
	newer["volsize"] = "2147483648" // the same volume, grown during the read
	remade := volumeObject("vp_b")
	remade["guid"] = "another" // deleted and made again under its name
	list := []object{older, volumeObject("vp_b"), newer, remade}
	app.on("GET", pool0+"/volumes", pages(func() []object { return list }))

	token, _ := NewCSIListingTokenFromTokenString("", VolumesList())
	vols, _, err := app.iscsi().ListAllVolumes(app.ctx, "Pool-0", 0, token)
	if err != nil {
		t.Fatal(err.Error())
	}
	var got []string
	for _, v := range vols.([]jrest.ResourceVolume) {
		got = append(got, fmt.Sprintf("%s %d %s", v.Name, v.GetSize(), v.Guid))
	}
	want := []string{"vp_a 2147483648 guid-vp_a", "vp_b 1073741824 guid-vp_b", "vp_b 1073741824 another"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("the volumes: %q, want %q", got, want)
	}
}
