package cmd

import (
	"bytes"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

type Options struct {
	Command     interface{}
	MergeStderr bool
	LiveOutput  bool
}

type CMD struct {
	Stdout     string
	Stderr     string
	ExitStatus int
	Options    Options
	Err        error
}

func New(opts Options) CMD {
	return CMD{
		Options:    opts,
		ExitStatus: -1,
	}
}

func (c *CMD) Run() *CMD {

	var out []byte
	var err error
	var command []string

	switch v := c.Options.Command.(type) {
	case string:
		command = []string{"/bin/bash", "-c", v}
	case []string:
		command = v
	default:
		c.Err = errors.New("Command must be of type string or []string")
		return c
	}

	cmd_to_log := bytes.Buffer{}
	for _, str := range command {
		if strings.Contains(str, " ") {
			cmd_to_log.WriteString(fmt.Sprintf("'%s' ", str))
		} else {
			cmd_to_log.WriteString(fmt.Sprintf("%s ", str))
		}
	}
	log.Debug(fmt.Sprintf("Running CMD: [%s]", cmd_to_log.String()))

	cmd := exec.Command(command[0], command[1:]...)

	if c.Options.LiveOutput {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		if err = cmd.Run(); err != nil {
			c.Err = err
			c.ExitStatus = getExitStatus(err)
			return c
		} else {
			c.ExitStatus = 0
		}
	} else if c.Options.MergeStderr {
		if out, err = cmd.CombinedOutput(); err != nil {
			c.Err = err
			c.ExitStatus = getExitStatus(err)
			return c
		} else {
			c.ExitStatus = 0
			c.Stdout = string(out)
		}
	} else {
		buffer := new(bytes.Buffer)
		cmd.Stderr = buffer
		if out, err = cmd.Output(); err != nil {
			c.Err = err
			c.ExitStatus = getExitStatus(err)
			return c
		} else {
			c.ExitStatus = 0
			c.Stdout = string(out)
		}
	}

	return c
}

func getExitStatus(err error) int {
	exitStatus := -1
	if exitError, ok := err.(*exec.ExitError); ok {
		exitStatus = exitError.Sys().(syscall.WaitStatus).ExitStatus()
	}
	return exitStatus
}
