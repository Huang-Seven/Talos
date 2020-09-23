package models

import (
	"Talos/conf"
	pb "Talos/internal/rpc/proto"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"google.golang.org/grpc"
)

type Agent struct {
	version     string
	Ac          conf.AgentConf
	pcList      []*Proc
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
	a.pcList = ReadProcDir(a.Ac.MonitorConfDir)
	a.pingServer()
	_ = os.MkdirAll(a.Ac.MonitorConfDir, 0755)

	a.opChan = make(chan *conf.Operation, 6)
	http.HandleFunc("/operate", operate(a.opChan))
}

func (a *Agent) pingServer() {
	log.Printf("Server addr: %v:%d", a.Ac.ServerAddr, a.Ac.ServerPort)
}

func (a *Agent) geneConf() {
	address := fmt.Sprintf("%v:%d", a.Ac.ServerAddr, a.Ac.ServerPort)
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("Fail to connect: %v, err: %v", address, err)
	}
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
		log.Printf("Found conf: %v", i)
	}
}

func (a *Agent) operate() {
	go func() {
		log.Println("Starting Waiting Message...")
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
				if c.FrequenceCollect != 0 {
					a.FrequenceCollect = c.FrequenceCollect
					log.Println("Change collect")
				}
				if c.MonitorSleep != 0 {
					a.MonitorSleep = c.MonitorSleep
					log.Println("Stop Monitor Process for ", a.MonitorSleep, " Seconds")
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
			case 2: //gene conf
				{
					log.Println("Re-genetate conf")
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
	log.Fatal(http.ListenAndServe(fmt.Sprintf("127.0.0.1:%d", a.Ac.LocalPort), nil))
}

func (a *Agent) Run() {
	a.init()
	a.operate()
	log.Println("Starting HttpServer...")
	defer a.wg.Done()
	go a.runHttp()
	a.wg.Add(1)
	log.Println("Run and service...")
	a.wg.Wait()
}
