package common

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// Every key of the endpoint, iscsi and nfs sections has a yaml tag, with
// the name its json tag has: the configuration is read from yaml, and
// without a tag a key is found only by the lower-cased name of its field.
func TestConfigurationKeysHaveYAMLTags(t *testing.T) {
	for _, section := range []interface{}{RestEndpointCfg{}, ISCSIEndpointCfg{}, NFSEndpointCfg{}} {
		typ := reflect.TypeOf(section)
		for i := 0; i < typ.NumField(); i++ {
			f := typ.Field(i)
			yamlName := f.Tag.Get("yaml")
			jsonName := strings.Split(f.Tag.Get("json"), ",")[0]
			if yamlName == "" || yamlName != jsonName {
				t.Errorf("%s.%s: yaml tag %q, json tag %q", typ.Name(), f.Name, yamlName, jsonName)
			}
		}
	}
}

// exampleOf returns the first block of a guide that opens with a line of
// ``` alone, without a language: its example configuration.
func exampleOf(t *testing.T, guide string) string {
	t.Helper()
	raw, err := os.ReadFile(guide)
	if err != nil {
		t.Fatal(err)
	}
	var block []string
	in, wanted := false, false
	for _, line := range strings.Split(string(raw), "\n") {
		fence := strings.HasPrefix(strings.TrimSpace(line), "```")
		switch {
		case fence && in && wanted:
			return strings.Join(block, "\n") + "\n"
		case fence && in:
			in = false
		case fence:
			in, wanted = true, strings.TrimSpace(line) == "```"
		case in && wanted:
			block = append(block, line)
		}
	}
	t.Fatalf("%s holds no example", guide)
	return ""
}

// A configuration copied from either guide, or from the example file, gives
// the driver its pool and its endpoint, as the code reads them.
func TestGuideExamplesAreReadAsWritten(t *testing.T) {
	root := filepath.Join("..", "..")
	sources := map[string]string{
		"docs/configuration.md":     exampleOf(t, filepath.Join(root, "docs", "configuration.md")),
		"docs/configuration-nfs.md": exampleOf(t, filepath.Join(root, "docs", "configuration-nfs.md")),
	}
	raw, err := os.ReadFile(filepath.Join(root, "deploy", "cfg", "cfg.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	sources["deploy/cfg/cfg.yaml"] = string(raw)

	for name, text := range sources {
		for _, wrong := range []string{"logpath", "iddletimeout", "  pool:"} {
			if strings.Contains(text, wrong) {
				t.Errorf("%s holds %q, which the code does not read", name, strings.TrimSpace(wrong))
			}
		}
		path := filepath.Join(t.TempDir(), "cfg.yaml")
		if err := os.WriteFile(path, []byte(text), 0o600); err != nil {
			t.Fatal(err)
		}
		var cfg JovianDSSCfg
		if err := SetupConfig(path, &cfg); err != nil {
			t.Errorf("%s: %v", name, err)
			continue
		}
		e := cfg.RestEndpointCfg
		if cfg.Pool != "Pool-0" || cfg.LLevel != "Info" || cfg.LDest != "" {
			t.Errorf("%s: pool %q, loglevel %q, logfile %q", name, cfg.Pool, cfg.LLevel, cfg.LDest)
		}
		if len(e.Addrs) != 1 || e.Addrs[0] != "192.168.0.100" || e.Port != 82 || e.User != "admin" || e.Pass != "admin" {
			t.Errorf("%s: the endpoint is read as %+v", name, e.Addrs)
		}
		if e.IdleTimeOut != "30s" || e.Tries != 3 || e.RetryDelay != "1s" || e.Prot != "" {
			t.Errorf("%s: idletimeout %q, tries %d, retrydelay %q, prot %q", name, e.IdleTimeOut, e.Tries, e.RetryDelay, e.Prot)
		}
		if strings.Contains(text, "iscsi:") {
			if cfg.ISCSIEndpointCfg == nil || cfg.ISCSIEndpointCfg.Iqn != "iqn.csi.2019-04" ||
				cfg.ISCSIEndpointCfg.Port != 3260 || len(cfg.ISCSIEndpointCfg.Addrs) != 1 {
				t.Errorf("%s: the iscsi section is read as %+v", name, cfg.ISCSIEndpointCfg)
			}
		}
		if strings.Contains(text, "nfs:") {
			if cfg.NFSEndpointCfg == nil || len(cfg.NFSEndpointCfg.Addrs) != 1 {
				t.Errorf("%s: the nfs section is read as %+v", name, cfg.NFSEndpointCfg)
			}
		}
	}
}
