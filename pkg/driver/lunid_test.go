package driver

import (
	"strings"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// The ID of a volume that comes in through a CSI request is checked against
// the rules the appliance has for a name, and refused with InvalidArgument
// when it breaks them.
func TestVolumeIDFromACSIRequest(t *testing.T) {
	for _, id := range []string{
		"vp_pvc-0a1b2c3d-4e5f-6789-abcd-ef0123456789",
		"vb_ORSXG5DJNZTQ----",
		"vs_9F86D081884C7D659A2FEAA0C55AD015A3BF4F1B2B0B822CD15D6C15B0F00A08",
		"v_legacy",
		"vp_" + strings.Repeat("a", 128), // 131 characters, the longest
	} {
		vd, err := NewVolumeDescFromCSIID(id)
		if err != nil || vd.VDS() != id || vd.CSIID() != id {
			t.Errorf("the ID %q is refused: %v", id, err)
		}
	}
	for what, id := range map[string]string{
		"an empty ID":           "",
		"a dot":                 ".",
		"two dots":              "..",
		"a slash":               "vp_a/b",
		"a parent":              "vp_a/../b",
		"a space":               "vp_a b",
		"an escaped slash":      "vp_a%2Fb",
		"a query":               "vp_a?x=1",
		"a fragment":            "vp_a#b",
		"a dot in the name":     "vp_a.b",
		"an equals sign":        "vb_ORSXG5A=",
		"a letter not in ASCII": "vp_ä",
		"a line break":          "vp_a\nb",
		"132 characters":        "vp_" + strings.Repeat("a", 129),
		"1000 characters":       "vp_" + strings.Repeat("a", 997),
	} {
		vd, err := NewVolumeDescFromCSIID(id)
		if err == nil || vd != nil || status.Code(err) != codes.InvalidArgument {
			t.Errorf("%s: the ID %q is taken, error %v", what, id, err)
		}
	}
	// a name the appliance takes, in a form the plugin does not make
	if _, err := NewVolumeDescFromCSIID("backup-volume"); status.Code(err) != codes.InvalidArgument {
		t.Errorf("an ID without the plugin's prefix: %v", err)
	}
}

// The ID of a snapshot that comes in through a CSI request names a snapshot
// and its volume; one that does not is refused with InvalidArgument.
func TestSnapshotIDFromACSIRequest(t *testing.T) {
	vd, err := NewVolumeDescFromVDS("vp_vol")
	if err != nil {
		t.Fatal(err)
	}
	made := NewSnapshotDescFromName(vd, "snap-1")
	sd, err := NewSnapshotDescFromCSIID(made.CSIID())
	if err != nil || sd.SDS() != "sp_snap-1" || sd.GetVD().VDS() != "vp_vol" {
		t.Fatalf("the ID %q of a snapshot the plugin made: %v", made.CSIID(), err)
	}
	for what, id := range map[string]string{
		"an empty ID":                      "",
		"a dot":                            ".",
		"two dots":                         "..",
		"no volume part":                   "sp_snap-1",
		"a volume part that is not base64": "sp_snap-1_!!!",
		"a volume part that is no volume":  "sp_snap-1_YmFja3Vw", // "backup"
		"text":                             "next page please",
	} {
		sd, err := NewSnapshotDescFromCSIID(id)
		if err == nil || sd != nil || status.Code(err) != codes.InvalidArgument {
			t.Errorf("%s: the ID %q is taken, error %v", what, id, err)
		}
	}
}

// No name the plugin generates for a volume, and so no CSI ID of a volume
// it creates, is longer than 128 bytes; every one of them is a name the
// appliance takes, and reads back as the ID of the same volume.
func TestGeneratedVolumeNames(t *testing.T) {
	longest := 0
	for n := 1; n <= 300; n++ {
		for what, name := range map[string]string{
			"plain":      strings.Repeat("a", n),
			"with a dot": strings.Repeat("a", n-1) + ".",
			"not ASCII":  strings.Repeat("ä", (n+1)/2),
		} {
			vd, err := NewVolumeDescFromName(name)
			if err != nil {
				t.Fatalf("a %s name of %d bytes: %v", what, len(name), err)
			}
			vds := vd.VDS()
			if len(vds) > longest {
				longest = len(vds)
			}
			if len(vds) > 128 {
				t.Fatalf("a %s name of %d bytes gives the appliance name %q of %d bytes", what, len(name), vds, len(vds))
			}
			if err := checkApplianceName(vds); err != nil {
				t.Fatalf("a %s name of %d bytes gives the appliance name %q: %v", what, len(name), vds, err)
			}
			if back, err := NewVolumeDescFromCSIID(vd.CSIID()); err != nil || back.VDS() != vds {
				t.Fatalf("the ID %q of a %s name of %d bytes does not read back: %v", vd.CSIID(), what, len(name), err)
			}
		}
	}
	if longest != 128 {
		t.Errorf("the longest generated name has %d bytes, want 128", longest)
	}

	for _, c := range []struct {
		name string
		form string
	}{
		{"pvc-0a1b2c3d-4e5f-6789-abcd-ef0123456789", "vp_"},
		{strings.Repeat("a", 125), "vp_"}, // the longest plain name
		{strings.Repeat("a", 126), "vs_"},
		{strings.Repeat(".", 75), "vb_"}, // its base32 form has 120 characters
		{strings.Repeat(".", 76), "vs_"}, // and that of 76 bytes has 128
	} {
		vd, _ := NewVolumeDescFromName(c.name)
		if !strings.HasPrefix(vd.VDS(), c.form) {
			t.Errorf("a name of %d bytes takes the form %s, want %s", len(c.name), vd.VDS()[:3], c.form)
		}
	}
}
