package install

import (
	"strings"
	"testing"
)

func TestHintPointsToOfficialGuide(t *testing.T) {
	cases := []struct {
		bin, goos, want string
	}{
		{"aws", "windows", "cli/latest/userguide/getting-started-install.html"},
		{"aws", "linux", "cli/latest/userguide/getting-started-install.html"},
		{"session-manager-plugin", "windows", "install-plugin-windows.html"},
		{"session-manager-plugin", "darwin", "install-plugin-macos-overview.html"},
		{"session-manager-plugin", "linux", "install-plugin-linux-overview.html"},
	}
	for _, c := range cases {
		if got := Hint(c.bin, c.goos); !strings.Contains(got, c.want) {
			t.Errorf("%s on %s: %q does not contain %q", c.bin, c.goos, got, c.want)
		}
	}
}
