package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"runtime"

	"github.com/kintoun-secops/kintoun-ssm-tunnel/internal/aws"
	"github.com/kintoun-secops/kintoun-ssm-tunnel/internal/install"
	"github.com/kintoun-secops/kintoun-ssm-tunnel/internal/ui"
)

func main() {
	missing := false
	for _, bin := range []string{"aws", "session-manager-plugin"} {
		if _, err := exec.LookPath(bin); err != nil {
			fmt.Fprintln(os.Stderr, install.Hint(bin, runtime.GOOS))
			missing = true
		}
	}
	if missing {
		fmt.Fprintln(os.Stderr, "\n설치 후 터미널을 새로 열고 다시 실행하세요.")
		os.Exit(1)
	}

	var state ui.State
	for {
		res, next, err := ui.Run(state)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		state = next

		switch res.Action {
		case ui.Quit:
			return
		case ui.Login:
			if err := runInteractive(aws.LoginCmd(res.Profile)); err != nil {
				fmt.Fprintf(os.Stderr, "aws login 실패: %v\n", err)
			}
		case ui.Connect:
			fmt.Printf("\n%s (%s) 터널을 엽니다.\n접속 주소: %s\n종료하려면 Ctrl+C 를 누르세요.\n\n",
				res.Instance.Name, res.Instance.ID, res.URL)
			if err := runInteractive(aws.PortForwardCmd(res.Profile, res.Instance.ID, res.Remote, res.Local)); err != nil {
				fmt.Fprintf(os.Stderr, "세션 종료: %v\n", err)
			}
		}
	}
}

// runInteractive hands the terminal to cmd. Ctrl+C reaches the child; this process ignores it
// so the wrapper keeps running and returns to the menu once the child exits.
func runInteractive(cmd *exec.Cmd) error {
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)
	defer signal.Stop(sigs)
	go func() {
		for range sigs {
		}
	}()

	return cmd.Run()
}
