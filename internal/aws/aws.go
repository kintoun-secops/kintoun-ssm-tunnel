// Package aws wraps the aws CLI so the TUI reuses the user's own profiles and credentials.
package aws

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

const Region = "ap-northeast-2"

type Instance struct {
	Name  string `json:"name"`
	ID    string `json:"id"`
	State string `json:"state"`
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

// Instances lists instances tagged SSMPortForward=true with their power state.
// Stopped instances are included so the caller can start them; terminated ones are left out.
func Instances(profile string) ([]Instance, error) {
	out, err := exec.Command("aws", "ec2", "describe-instances",
		"--filters", "Name=tag:SSMPortForward,Values=true",
		"Name=instance-state-name,Values=pending,running,stopping,stopped",
		"--query", "Reservations[].Instances[].{name:Tags[?Key=='Name']|[0].Value,id:InstanceId,state:State.Name}",
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

// StartInstance powers on an instance. It returns once AWS accepts the request,
// not when the instance has reached the running state.
func StartInstance(profile, id string) error {
	return power(profile, id, "start-instances")
}

// StopInstance powers off an instance.
func StopInstance(profile, id string) error {
	return power(profile, id, "stop-instances")
}

func power(profile, id, action string) error {
	out, err := exec.Command("aws", "ec2", action, "--instance-ids", id,
		"--region", Region, "--profile", profile).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
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

// FreePort asks the OS for an unused local TCP port. There is a small window between
// closing the listener and the tunnel binding it, but it avoids fixed-port collisions.
func FreePort() (int, error) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}
