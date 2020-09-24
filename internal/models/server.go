package models

import (
	"Talos/conf"
	"context"
	"fmt"
	"log"
	"net"
	"strconv"
	"sync"
	"time"

	pb "Talos/internal/rpc/proto"

	"google.golang.org/grpc"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Server struct {
	Sc conf.ServerConf
	db *gorm.DB
	wg sync.WaitGroup
	pb.UnimplementedMonitorConfGetterServer
	pb.UnimplementedProcessEventServer
}

type Modinfo struct {
	Module       string `json:"module",mapstructure:"module"`
	Cwd          string `json:"cwd",mapstructure:"cwd"`
	Env          string `json:"env",mapstructure:"env"`
	Contact      string `json:"contact",mapstructure:"contact"`
	Cmdline      string `json:"cmdline",mapstructure:"cmdline"`
	Script       string `json:"script",mapstructure:"script"`
	Procnum      int32  `json:"procnum",mapstructure:"procnum"`
	Logpath      string `json:"logpath",mapstructure:"logpath"`
	Lognum       int32  `json:"lognum",mapstructure:"lognum"`
	Logsize      int32  `json:"logsize",mapstructure:"logsize"`
	Cmd          string `json:"cmd",mapstructure:"cmd"`
	Restartlimit int32  `json:"restartlimit",mapstructure:"restartlimit"`
}

type ProcessMonitor struct {
	ModuleName string    `json:"module_name"`
	Env        string    `json:"env"`
	StopTime   time.Time `json:"stop_time"`
	StartTime  time.Time `json:"start_time"`
	CostTime   int32     `json:"cost_time"`
	Host       string    `json:"host"`
	EventType  int32     `json:"event_type"`
	MailList   string    `json:"mail_list"`
}

func (ProcessMonitor) TableName() string {
	return "process_monitor"
}

func (s *Server) init() {
	log.Println("Init server...")
	s.Sc.LoadConf()
	s.getDB()
}

func (s *Server) getDB() {
	log.Println("Get mysql conn...")
	dsn := fmt.Sprintf("%v:%v@tcp(%v:%d)/%v?charset=utf8mb4&parseTime=True&loc=Local",
		s.Sc.MySQL.MysqlUser, s.Sc.MySQL.MysqlPwd,
		s.Sc.MySQL.MysqlHost, s.Sc.MySQL.MysqlPort,
		s.Sc.MySQL.MysqlDb)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(fmt.Errorf(time.Now().Format("2006-01-02 15:04:05"),
			"Fatal error connect to mysql: %s \n", err))
	}
	s.db = db
}

func (s *Server) loadModInfo() []*pb.MonitorConf {
	var cnf []*pb.MonitorConf
	s.db.Model(&Modinfo{}).Find(&cnf)
	return cnf
}

func (s *Server) GetConf(ctx context.Context, in *pb.MonitorConfRequest) (*pb.MonitorConfResponse, error) {
	host := in.GetHost()
	log.Printf("Receive: %v", host)
	cnfs := s.loadModInfo()
	return &pb.MonitorConfResponse{Confs: cnfs}, nil
}

func (s *Server) ProcessEventHandler(ctx context.Context, in *pb.ProcessEventRequest) (*pb.ProcessEventResponse, error) {
	stopT, err := time.Parse(tfmt, in.StopTime)
	if err != nil {
		log.Println(stopT, err)
	}
	startT, _ := time.Parse(tfmt, in.StartTime)
	ct, _ := strconv.ParseInt(in.CostTime, 10, 32)
	pm := ProcessMonitor{
		ModuleName: in.ModuleName,
		Host:       in.Host,
		Env:        in.Env,
		StopTime:   stopT,
		StartTime:  startT,
		CostTime:   int32(ct),
		EventType:  int32(in.EventType),
		MailList:   in.MailList}
	_ = s.db.Create(&pm)
	return &pb.ProcessEventResponse{StatusCode: 200, Message: "success"}, nil
}

func (s *Server) Run() {
	s.init()
	log.Printf("Listen port %v", s.Sc.LocalPort)
	ls := fmt.Sprintf(":%d", s.Sc.LocalPort)
	lis, err := net.Listen("tcp", ls)
	if err != nil {
		log.Fatalf("Fail to listen: %v", err)
	}
	gs := grpc.NewServer()
	pb.RegisterMonitorConfGetterServer(gs, s)
	pb.RegisterProcessEventServer(gs, s)
	if err := gs.Serve(lis); err != nil {
		log.Fatalf("Fail to serve: %v", err)
	}
}
