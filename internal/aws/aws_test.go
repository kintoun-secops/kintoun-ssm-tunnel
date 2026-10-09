package aws

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestProfilesSkipsDefault(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config")
	cfg := "[default]\nregion = us-east-1\n\n[profile alice]\nregion = ap-northeast-2\n\n[profile bob]\n[sso-session x]\n"
	if err := os.WriteFile(path, []byte(cfg), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("AWS_CONFIG_FILE", path)

	want := []string{"alice", "bob"}
	if got := Profiles(); !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestParsePort(t *testing.T) {
	for s, ok := range map[string]bool{"80": true, "65535": true, "0": false, "65536": false, "abc": false, "": false} {
		if _, got := ParsePort(s); got != ok {
			t.Errorf("%q: got %v, want %v", s, got, ok)
		}
	}
}
