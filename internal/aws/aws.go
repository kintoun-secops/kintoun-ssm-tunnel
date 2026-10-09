// Package aws wraps the aws CLI so the TUI reuses the user's own profiles and credentials.
package aws

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

const Region = "ap-northeast-2"

type Instance struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

// Profiles returns the named profiles in the AWS config file. The default profile is excluded
// because it can point at another account.
func Profiles() []string {
	path := os.Getenv("AWS_CONFIG_FILE")
	if path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil
		}
		path = filepath.Join(home, ".aws", "config")
	}
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()

	var profiles []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if name, ok := strings.CutPrefix(line, "[profile "); ok {
			if name, ok = strings.CutSuffix(name, "]"); ok {
				profiles = append(profiles, strings.TrimSpace(name))
			}
		}
	}
	return profiles
}

// CheckCredentials fails when the profile has no valid temporary credentials.
func CheckCredentials(profile string) error {
	out, err := exec.Command("aws", "sts", "get-caller-identity",
		"--profile", profile, "--region", Region, "--query", "Arn", "--output", "text").CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func LoginCmd(profile string) *exec.Cmd {
	return exec.Command("aws", "login", "--profile", profile, "--region", Region)
}

// PortForwardInstances lists running instances tagged SSMPortForward=true.
func PortForwardInstances(profile string) ([]Instance, error) {
	out, err := exec.Command("aws", "ec2", "describe-instances",
		"--filters", "Name=tag:SSMPortForward,Values=true", "Name=instance-state-name,Values=running",
		"--query", "Reservations[].Instances[].{name:Tags[?Key=='Name']|[0].Value,id:InstanceId}",
		"--output", "json", "--region", Region, "--profile", profile).Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("%w: %s", err, strings.TrimSpace(string(ee.Stderr)))
		}
		return nil, err
	}
	var list []Instance
	if err := json.Unmarshal(out, &list); err != nil {
		return nil, err
	}
	return list, nil
}

func PortForwardCmd(profile, instanceID string, remote, local int) *exec.Cmd {
	params := fmt.Sprintf(`{"portNumber":["%d"],"localPortNumber":["%d"]}`, remote, local)
	return exec.Command("aws", "ssm", "start-session",
		"--target", instanceID,
		"--document-name", "AWS-StartPortForwardingSession",
		"--parameters", params,
		"--region", Region, "--profile", profile)
}

func ParsePort(s string) (int, bool) {
	n, err := strconv.Atoi(s)
	return n, err == nil && n >= 1 && n <= 65535
}
