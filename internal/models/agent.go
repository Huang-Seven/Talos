package models

import (
	"Talos/conf"
	pb "Talos/internal/rpc/proto"
	"Talos/tools"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/spf13/viper"

	"google.golang.org/grpc"
)

const tfmt = "2006-01-02 15:04:05"

type locks struct {
	flag int //0 read 1 write 2 delete
	tsp  *tools.TaskStatus
	cb   chan bool
}

type Agent struct {
	version     string
	Ac          conf.AgentConf
	pcList      []*Proc
	taskin      chan *tools.Task
	taskout     chan *tools.TaskStatus
	taskBuf     map[int]tools.TaskStatus
	lockChan    chan *locks
	opChan      chan *conf.Operation
	startTime   time.Time
	wg          sync.WaitGroup
	hostCharger string
	maintain    bool
	conf.OperationAgent
}

func (a *Agent) init() {
	log.Println("Init agent...")
	a.Ac.LoadConf()
	a.pingServer()
	if !tools.IsDir(a.Ac.MonitorConfDir) {
		log.Println("Monitor conf_dir not found, create it.")
		err := os.MkdirAll(a.Ac.MonitorConfDir, 0755)
		if err != nil {
			log.Printf("Create conf_dir fail: %v", err)
		} else {
			a.geneConf()
		}
	} else {
		a.pcList = ReadProcDir(a.Ac.MonitorConfDir)
	}
	a.FrequencyMonitor = 5
	if a.Ac.LocalAddr == "" {
		ip, err := tools.GetClientIp()
		if err != nil {
			log.Printf("Get ip fail: %v", err)
		} else {
			a.Ac.LocalAddr = ip
			log.Printf("Get local addr: %v", ip)
		}
	}
	a.opChan = make(chan *conf.Operation, 6)
	a.taskin = make(chan *tools.Task, 10)
	a.taskout = make(chan *tools.TaskStatus, 10)
	a.taskBuf = make(map[int]tools.TaskStatus, 100)
	a.lockChan = make(chan *locks, 100)
	http.HandleFunc("/operate", operate(a.opChan))
}

func (a *Agent) pingServer() {
	log.Printf("Server addr: %v:%d", a.Ac.ServerAddr, a.Ac.ServerPort)
}

func newConf(path string, m *pb.MonitorConf) {
	v := viper.New()
	v.SetConfigFile(path + "/" + m.Module + ".toml")

	v.Set("module", m.Module)
	v.Set("cwd", m.Cwd)
	v.Set("env", m.Env)
	v.Set("contact", m.Contact)
	v.Set("cmdline", m.Cmdline)
	v.Set("script", m.Script)
	v.Set("Procnum", m.Procnum)
	v.Set("logpath", m.Logpath)
	v.Set("lognum", m.Lognum)
	v.Set("logsize", m.Logsize)
	v.Set("cmd", m.Cmd)
	v.Set("restartlimit", m.Restartlimit)

	err := v.WriteConfig()
	if err != nil {
		log.Println("Error: write config failed: ", err)
	}
}

func (a *Agent) getRpcConn() (conn *grpc.ClientConn) {
	address := fmt.Sprintf("%v:%d", a.Ac.ServerAddr, a.Ac.ServerPort)
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("Fail to connect: %v, err: %v", address, err)
	}
	return
}

func (a *Agent) geneConf() {
	conn := a.getRpcConn()
	defer conn.Close()
	c := pb.NewMonitorConfGetterClient(conn)
	cntx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.GetConf(cntx, &pb.MonitorConfRequest{Host: "127.0.0.1"})
	if err != nil {
		log.Fatalf("Could not to get conf: %v", err)
	}
	confs := r.GetConfs()
	for _, i := range confs {
		go newConf(strings.TrimSuffix(a.Ac.MonitorConfDir, "/"), i)
	}
	a.pcList = ReadProcDir(a.Ac.MonitorConfDir)
}

func (a *Agent) operate() {
	go func() {
		log.Println("Starting waiting message...")
		defer a.wg.Done()
		a.wg.Add(1)
		for c := range a.opChan {
			log.Println("Got a message")
			switch c.Action {
			case 0:
				if c.FrequencyMonitor != 0 {
					a.FrequencyMonitor = c.FrequencyMonitor
					log.Println("Change monitor")
				}
				if c.MonitorSleep != 0 {
					a.MonitorSleep = c.MonitorSleep
					log.Printf("Stop monitor process for %d seconds", a.MonitorSleep)
				}
				if c.Maintain == 1 {
					a.maintain = false
					log.Println("Maintain off")
				} else if c.Maintain == 2 {
					a.maintain = true
					log.Println("Maintain on")
				}
			case 1: //reload
				{
					log.Println("Reload conf")
				}
			case 2: //gene monitor conf
				{
					log.Println("Regenerate conf")
					a.geneConf()
				}
			default:
				{
					log.Println("Unknow action:", c.Action)
				}
			}
			log.Println("Agent get message done")
		}
	}()
}

func (a *Agent) runHttp() {
	log.Fatal(http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", a.Ac.LocalPort), nil))
}

func (a *Agent) MonitorProcess() {
	go func() {
		log.Println("Starting monitor process...")
		defer a.wg.Done()
		a.wg.Add(1)
		for {
			pt, _ := procList()
			for _, pc := range a.pcList {
				af := false
				if strings.Trim(pc.Script, " ") == "" {
					af = pt.alive(pc)
				} else {
					err := os.Chdir(a.Ac.CheckDir)
					if err != nil {
						log.Println("Change dir", err)
						continue
					} else {
						af = pc.Check()
						_ = os.Chdir(a.Ac.StartDir)
					}
				}
				if af == false {
					procDown(a, pc)
				} else {
					procUp(a, pc)
				}
			}
			sleep(1)
			if a.MonitorSleep != 0 {
				sleep(a.MonitorSleep)
				a.MonitorSleep = 0
			}
		}
	}()
}

func (a *Agent) RunTask() {
	wi := 10
	for i := 0; i < wi; i++ {
		go func(i int) {
			log.Printf("Starting task worker%d...", i)
			defer a.wg.Done()
			a.wg.Add(1)
			for t := range a.taskin {
				var ts tools.TaskStatus
				ts.Status = 999
				if t.Path == "" {
					t.Path = a.Ac.RunDir
					log.Printf("Tid[%d] lost path using default[%s]", t.Tid, t.Path)
				}
				log.Printf("Worker[%d]: run cmd[%s],path[%s],Tid[%d]", i, t.Cmd, t.Path, t.Tid)
				t.RunTask(&ts)
				log.Printf("Worker[%d]: task[%d] done\n", i, t.Tid)
				a.taskout <- &ts
			}
		}(i)
	}
}

func (a *Agent) getMail(pc *Proc, downUp int) []string {
	var contentList []string
	contentList = append(contentList, fmt.Sprintf("负责人:%s", pc.Contact))
	contentList = append(contentList, fmt.Sprintf("服务名称:%s", pc.Module))
	var s string
	if downUp == 0 {
		s = "进程终止"
	} else {
		s = "进程恢复"
	}
	contentList = append(contentList, fmt.Sprintf("报警原因: %s", s))
	contentList = append(contentList, fmt.Sprintf("机器地址: %s", a.Ac.LocalAddr))
	contentList = append(contentList, fmt.Sprintf("累次报警次数: %d", pc.RestartNum))
	contentList = append(contentList, fmt.Sprintf("首次报警时间: %s", pc.Ts1.Format(tfmt)))
	contentList = append(contentList, fmt.Sprintf("末次报警时间: %s", pc.Ts2.Format(tfmt)))
	contentList = append(contentList, fmt.Sprintf("报警持续时间: %s", pc.Sp.String()))
	if pc.RestartNum >= pc.Restartlimit && downUp == 0 {
		contentList = append(contentList, "未设置重启或超出重试次数，请人工检查")
	}
	return contentList
}

func sendmail(title, content, to string) {
	//TODO
	log.Printf("SendMail-> title: %v content: %v to: %v", title, content, to)
}

//rpc server 保存事件信息
func (a *Agent) postProcess(pc *Proc) {
	et := int64(1) //event type
	conn := a.getRpcConn()
	defer conn.Close()
	c := pb.NewProcessEventClient(conn)
	cntx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.ProcessEventHandler(cntx, &pb.ProcessEventRequest{
		ModuleName: pc.Module,
		Env:        pc.Env,
		StopTime:   pc.Ts1.Format(tfmt),
		StartTime:  pc.Ts2.Format(tfmt),
		CostTime:   pc.Sp.String(),
		Host:       a.Ac.LocalAddr,
		EventType:  et,
		MailList:   pc.Contact})
	if err != nil {
		log.Fatalf("Could not to get conf: %v", err)
	}
	log.Printf("PostProcess info -> status_code: %d, message: %v", r.GetStatusCode(), r.GetMessage())
}

func procDown(a *Agent, pc *Proc) {
	log.Printf("Process: %v is down", pc.Module)
	t := tools.Task{Tid: tools.RandTid(),
		Cmd:  pc.Cmd,
		Path: pc.Cwd}
	n := time.Now()
	if pc.RestartNum == 0 {
		pc.Ts1 = n
	}
	pc.Ts2 = n
	_ = pc.CalSp()
	if pc.RestartNum < pc.Restartlimit {
		a.taskin <- &t
	}
	pc.RestartNum += 1
	title := fmt.Sprintf("[%s][(%s)服务异常]", pc.Module, pc.Env)
	contentList := a.getMail(pc, 0)
	content := strings.Join(contentList, "\n")
	mt := pc.Contact
	if a.maintain {
		title += "(该机器维护中)"
		mt = "devops@300.cn"
	}
	sendmail(title, content, mt)
	sleep(5)
}

func procUp(a *Agent, pc *Proc) {
	if pc.RestartNum != 0 {
		pc.Ts2 = time.Now()
		_ = pc.CalSp()
		//保存事件记录
		a.postProcess(pc)
		title := fmt.Sprintf("[%s][(%s)服务恢复]", pc.Module, pc.Env)
		contentList := a.getMail(pc, 1)
		content := strings.Join(contentList, "\n")
		mt := pc.Contact
		if a.maintain {
			title += "(该机器维护中)"
			mt = "devops@300.cn"
		}
		sendmail(title, content, mt)
		pc.Reset()
	}
}

func sleep(i int) {
	time.Sleep(time.Duration(i) * time.Second)
}

func (a *Agent) Run() {
	a.init()
	a.operate()
	a.RunTask()
	a.MonitorProcess()
	defer a.wg.Done()
	log.Println("Starting http server...")
	go a.runHttp()
	a.wg.Add(1)
	log.Println("Run and service...")
	a.wg.Wait()
}
