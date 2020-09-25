package tools

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

func execShell(oneTask Task, oneTaskSta *TaskStatus) {
	oneTaskSta.Tid = oneTask.Tid
	oneTaskSta.Begin = time.Now()

	taskArr := strings.Split(oneTask.Cmd, " ")
	c := taskArr[0]
	// test if absolute path
	if strings.HasPrefix(c, "/") == false {
		fC, err := exec.LookPath(c)
		if err != nil {
			log.Printf("Looking cmd[%s] error[%s]\n", c, err.Error())
		} else {
			log.Printf("Replace cmd[%s] with[%s]\n", c, fC)
			taskArr[0] = fC
		}
	}
	Cmd := exec.Cmd{Path: taskArr[0],
		Args: taskArr,
		Dir:  oneTask.Path}

	var stdout, stderr bytes.Buffer
	Cmd.Stdout = &stdout
	Cmd.Stderr = &stderr
	if err := Cmd.Run(); err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			if s, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				oneTaskSta.Ret = s.ExitStatus()
			}
		}
		log.Printf("Cmd run error [%s]", err.Error())
		oneTaskSta.Status = 203
		errInfo := fmt.Sprintf("%s|%s", err.Error(), stderr.String())
		oneTaskSta.Err = errInfo
		oneTaskSta.End = time.Now()
		//one_task_sta.Ret = s.ExitStatus()
		return
	}
	oneTaskSta.Info = stdout.String()
	oneTaskSta.Err = stderr.String()

	if err := Cmd.Wait(); err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			if Status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				oneTaskSta.Ret = Status.ExitStatus()
			} else {
				oneTaskSta.Status = 207
				oneTaskSta.End = time.Now()
				return
			}
		} else {
			oneTaskSta.Status = 0
		}
	} else {
		oneTaskSta.Ret = 0
	}
	oneTaskSta.Status = 200
	oneTaskSta.End = time.Now()
	return
}
