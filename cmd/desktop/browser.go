package main

import (
	"fmt"
	"os/exec"
	"runtime"
	"time"

	"github.com/pkg/errors"
)

// Browser opens the default browser for a given url.
type Browser struct {
	OnErr     func(error)
	OnRestart func()
	Loop      bool
	failed    chan error
}

// Run a browser window, re-opening it when closed if Loop is true.
// Note: open simply uses OS default application, which may not actually be a
// web browser.
// Run blocks.
func (b *Browser) Run(u string) error {
	cmd := b.open(u)
	if cmd == nil {
		return fmt.Errorf("not implemented on this OS")
	}
	b.failed = make(chan error)
	if b.OnErr == nil {
		b.OnErr = func(error) {}
	}
	if b.OnRestart == nil {
		b.OnRestart = func() {}
	}
	go func() {
		for err := range b.failed {
			b.OnErr(err)
		}
	}()
	for {
		cmd := b.open(u)
		if err := func() error {
			if out, err := cmd.CombinedOutput(); err != nil {
				return errors.Wrap(err, string(out))
			}
			return nil
		}(); err != nil {
			b.failed <- err
		}
		if !b.Loop {
			break
		}
		time.Sleep(time.Millisecond * 200)
		b.OnRestart()
	}
	return nil
}

func (Browser) open(u string) *exec.Cmd {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", u, "-W")
	case "windows":
		cmd = exec.Command("start", u)
	default:
		cmd = exec.Command("xdg-open", u)
	}
	return cmd
}
