// Package install explains how to get the CLI tools this program shells out to.
package install

import "fmt"

const (
	awsCLIGuide = "https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html"
	pluginDocs  = "https://docs.aws.amazon.com/systems-manager/latest/userguide/"
)

// Hint returns a message telling the user where the official install guide for bin is.
func Hint(bin, goos string) string {
	switch bin {
	case "aws":
		return fmt.Sprintf("aws CLI v2 가 필요합니다.\n  설치 안내: %s", awsCLIGuide)
	case "session-manager-plugin":
		return fmt.Sprintf("Session Manager 플러그인이 필요합니다.\n  설치 안내: %s", pluginGuide(goos))
	}
	return fmt.Sprintf("%s 를 찾을 수 없습니다.", bin)
}

func pluginGuide(goos string) string {
	switch goos {
	case "windows":
		return pluginDocs + "install-plugin-windows.html"
	case "darwin":
		return pluginDocs + "install-plugin-macos-overview.html"
	}
	return pluginDocs + "install-plugin-linux-overview.html"
}
