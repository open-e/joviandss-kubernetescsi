package controller

import (
	"strings"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// The iscsi.iqn setting is the beginning of every target name the plugin
// builds, <prefix>:<64 hex digits>. It is checked once against the rule for
// target names: lower case, digits, ".", ":" and "-", and 220 characters at
// most for the whole name.
func TestIQNPrefix(t *testing.T) {
	for _, prefix := range []string{
		"iqn.csi.2019-04", // the default
		"iqn.2026-11.com.example:storage",
		"a",
		"iqn." + strings.Repeat("a", 151), // 155 + 1 + 64 = 220
	} {
		if err := checkIQNPrefix(prefix); err != nil {
			t.Errorf("the prefix %q is refused: %v", prefix, err)
		}
	}
	for what, prefix := range map[string]string{
		"upper case":             "iqn.CSI.2019-04",
		"an underscore":          "iqn.csi_2019",
		"a space":                "iqn.csi 2019",
		"a slash":                "iqn.csi/2019",
		"a dot at the end":       "iqn.csi.",
		"a colon at the end":     "iqn.csi:",
		"a dash at the start":    "-iqn.csi",
		"nothing":                "",
		"too long with the hash": "iqn." + strings.Repeat("a", 152),
	} {
		err := checkIQNPrefix(prefix)
		if err == nil || status.Code(err) != codes.InvalidArgument || !strings.Contains(err.Error(), "iscsi.iqn") {
			t.Errorf("%s: the prefix %q: error %v, want one that names iscsi.iqn", what, prefix, err)
		}
	}
}
