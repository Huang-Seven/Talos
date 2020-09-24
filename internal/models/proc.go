package models

import (
	"Talos/tools"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/viper"
)

const baseDir = "/proc"

type Proc struct {
	Tid        int
	ProcNum    int32
	Alive      bool
	Ts1        time.Time
	Ts2        time.Time
	Sp         time.Duration
	RestartNum int32
	Modinfo    `mapstructure:",squash"`
}

type ProcT struct {
	tid, ppid, pgrp, session, tty, tpgid     int
	flags, minFlt, cminFlt, majFlt, cmajFlt  int
	utime, stime, cutime, cstime, priority   int
	nice, nlwp, alarm, startTime, vsize, rss int
	rssRlim, startCode, endCode, startStack  int
	kstkEsp, kstkEip, wchan, exitSignal      int
	processor, rtprio, sched                 int
	exe, cwd, cmd, state, cmdline, procDir   string
}

func ReadProc(path, name string) (*Proc, error) {
	log.Printf("path: %v, name: %v", path, name)
	pc := new(Proc)
	v := viper.New()
	v.AddConfigPath(path)
	v.SetConfigName(name)
	v.SetConfigType("toml")
	err := v.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf(time.Now().Format("2006-01-02 15:04:05"),
			"Fatal error config file: %s \n", err))
	}
	err = v.Unmarshal(&pc)
	if err != nil {
		log.Println(err)
	}
	return pc, err
}

func ReadProcDir(path string) []*Proc {
	log.Printf("Read Monitor conf dir: %v", path)
	dList, _ := ioutil.ReadDir(path)
	var pList []*Proc
	for _, d := range dList {
		n := d.Name()
		if strings.HasSuffix(n, ".toml") == false {
			continue
		}
		pc, err := ReadProc(path, n)
		if err != nil {
			errInfo := fmt.Sprintf("Read [%s] Error [%s]", n, err.Error())
			log.Println(errInfo)
			continue
		}
		pList = append(pList, pc)
	}
	return pList
}

type ProcTab struct {
	name2proc map[string][]*ProcT
	pid2proc  map[int]*ProcT
}

func (pc *Proc) Reset() {
	pc.RestartNum = 0
}

func (pc *Proc) CalSp() (err error) {
	err = nil
	f := "2006-01-02 15:04:05"
	if pc.Ts2.After(pc.Ts1) {
		pc.Sp = pc.Ts2.Sub(pc.Ts1)
	} else {
		errStr := fmt.Sprintf("TimeStamp2[%s] sub TimeStamp1[%s] Error", pc.Ts2.Format(f), pc.Ts1.Format(f))
		err = errors.New(errStr)
	}
	return err
}

func newproc(tid int) (*ProcT, error) {
	p := new(ProcT)
	p.tid = tid
	p.procDir = fmt.Sprintf("%s/%d", baseDir, p.tid)
	//p.readstat()
	p.readcwd()
	p.readcmdl()
	p.readexe()
	return p, nil
}

func (p *ProcT) readstat() {
	//TODO
}

func (p *ProcT) readcwd() {
	p.cwd, _ = os.Readlink(p.procDir + "/cwd")
}
func (p *ProcT) readexe() {
	p.exe, _ = os.Readlink(p.procDir + "/exe")
}
func (p *ProcT) readcmdl() {
	p.cmdline, _ = readfile(p.procDir + "/cmdline")
}

func readfile(filename string) (string, error) {
	bO, err := ioutil.ReadFile(filename)
	var bN []byte
	for _, b := range bO {
		var c byte = 0
		if b > ' ' && b < '~' {
			c = b
		} else {
			c = ' '
		}
		bN = append(bN, c)
	}
	s := ""
	if err != nil {
		//err_info := fmt.Sprintf("Read [%s] file error[%s]",filename,err.Error())
		//log.Println(err_info)
	} else {
		s = string(bN)
		s = strings.Trim(s, " ")
	}
	return s, err
}

func procList() (ProcTab, error) {
	pt := ProcTab{}
	pt.name2proc = make(map[string][]*ProcT, 300)
	pt.pid2proc = make(map[int]*ProcT, 300)
	procList, _ := ioutil.ReadDir(baseDir)
	for _, d := range procList {
		n := d.Name()
		tid, err := strconv.Atoi(n)
		//Not task dir continue
		if err != nil {
			continue
		}
		p, _ := newproc(tid)
		if p.tid == -1 {
			continue
		}
		pt.name2proc[p.cmd] = append(pt.name2proc[p.cmd], p)
		pt.name2proc[p.cmdline] = append(pt.name2proc[p.cmdline], p)
		pt.pid2proc[p.tid] = p
	}
	return pt, nil
}

func (pt *ProcTab) checkCmdline(cmdline string) []int {
	iL := make([]int, 0)
	for k, vL := range pt.name2proc {
		i := strings.Index(k, cmdline)
		if i == -1 {
			continue
		}
		for _, v := range vL {
			iL = append(iL, v.tid)
		}
	}
	return iL
}

func (pt *ProcTab) alive(pc *Proc) bool {
	tidL := pt.checkCmdline(pc.Cmdline)
	if len(tidL) == 0 {
		pc.Alive = false
	}
	pNum := 0
	for _, t := range tidL {
		_, exist := pt.pid2proc[t]
		if exist != true {
			continue
		}
		pNum += 1
	}
	pc.Alive = pNum >= int(pc.ProcNum)
	return pc.Alive
}

func (pc *Proc) Check() (f bool) {
	if pc.Script != "" {
		cmd := "/bin/bash"
		ec, _, _ := tools.RunCmd(cmd, pc.Script)
		if ec == 0 {
			f = true
		} else {
			f = false
		}
	} else {
		log.Printf("[%s]Pc Script empty!", pc.Module)
		f = false
	}
	return f
}
