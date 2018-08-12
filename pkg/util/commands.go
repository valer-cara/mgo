package util

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

// Shamelessly extracted from https://github.com/roboll/helmfile/blob/master/helmexec/runner.go
func Exec(cmd string, args []string, env []string) ([]byte, error) {
	dir, err := ioutil.TempDir("/tmp", "mygitopscmd-"+cmd+"-exec")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(dir)

	preparedCmd := exec.Command(cmd, args...)
	preparedCmd.Dir = dir
	preparedCmd.Env = env

	return preparedCmd.CombinedOutput()
}

func RunCommands(env []string, commands ...*exec.Cmd) error {
	for _, cmd := range commands {
		var out bytes.Buffer

		cmd.Env = env

		// Both, maybe one contains valuable info
		cmd.Stdout = &out
		cmd.Stderr = &out

		err := cmd.Run()
		if err != nil {
			return errors.New(fmt.Sprintf("[failed] %s: %s", strings.Join(cmd.Args, " "), out.String()))
		}
	}
	return nil
}

func CallFunctions(functions ...func() error) error {
	for _, f := range functions {
		if err := f(); err != nil {
			return err
		}
	}
	return nil
}
