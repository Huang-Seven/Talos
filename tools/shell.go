package tools

import (
	"bytes"
	"log"
	"os/exec"
	"syscall"
)

func RunCmd(cmd string, args ...string) (ec int, so, se string) {
	var ob, eb bytes.Buffer
	c := exec.Command(cmd, args...)
	c.Stdout = &ob
	c.Stderr = &eb

	err := c.Run()
	so = ob.String()
	se = eb.String()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			ws := exitError.Sys().(syscall.WaitStatus)
			ec = ws.ExitStatus()
		} else {
			log.Printf("Can't get exit code for failed program:%v,%v", cmd, args)
		}
		if se == "" {
			se = err.Error()
		}
		ec = -255
	} else {
		ws := c.ProcessState.Sys().(syscall.WaitStatus)
		ec = ws.ExitStatus()
	}
	return ec, so, se
}
