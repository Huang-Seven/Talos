package tools

import "time"

type Task struct {
	Path string
	Cmd  string
	Tid  int
}

type TaskStatus struct {
	Tid    int
	Begin  time.Time
	End    time.Time
	Status int //101:正在运行, 200:正常结束 2xx:异常结束 301:结束并且已经检查
	Ret    int //退出码
	Info   string
	Err    string
}

func (task Task) RunTask(status *TaskStatus) {
	execShell(task, status)
}
