package models

import (
	"Talos/conf"
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	pb "Talos/internal/rpc/proto"

	"google.golang.org/grpc"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Server struct {
	Sc conf.ServerConf
	//ip2info     map[string]*hostinfo
	//mod2ip      map[string]map[string][]string
	//module2pc   map[string][]*proc.ProcC
	//baseModules []*proc.ProcC
	//hostsInfo   map[string][]*string //hostname, owner
	//conn        sqllib.Myconn
	wg sync.WaitGroup
	//commChan    chan *commChanStruct
	pb.UnimplementedMonitorConfGetterServer
}

type Modinfo struct {
	Module       string `json:"module",mapstructure:"module"`
	Cwd          string `json:"cwd"`
	Env          string `json:"env"`
	Contact      string `json:"contact"`
	Cmdline      string `json:"cmdline"`
	Script       string `json:"script"`
	Procnum      int32  `json:"procnum"`
	Logpath      string `json:"logpath"`
	Lognum       int32  `json:"lognum"`
	Logsize      int32  `json:"logsize"`
	Cmd          string `json:"cmd"`
	Restartlimit int32  `json:"restartlimit"`
}

func (s *Server) init() {
	log.Println("Init server...")
	s.Sc.LoadConf()
}

func (s *Server) loadModInfo() []*pb.MonitorConf {
	dsn := fmt.Sprintf("%v:%v@tcp(%v:%d)/%v?charset=utf8mb4&parseTime=True&loc=Local",
		s.Sc.MySQL.MysqlUser, s.Sc.MySQL.MysqlPwd,
		s.Sc.MySQL.MysqlHost, s.Sc.MySQL.MysqlPort,
		s.Sc.MySQL.MysqlDb)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(fmt.Errorf(time.Now().Format("2006-01-02 15:04:05"),
			"Fatal error connect to mysql: %s \n", err))
	}
	var cnf []*pb.MonitorConf
	db.Model(&Modinfo{}).Find(&cnf)
	return cnf
}

func (s *Server) GetConf(ctx context.Context, in *pb.MonitorConfRequest) (*pb.MonitorConfResponse, error) {
	host := in.GetHost()
	log.Printf("Receive: %v", host)
	cnfs := s.loadModInfo()

	return &pb.MonitorConfResponse{Confs: cnfs}, nil

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
	if err := gs.Serve(lis); err != nil {
		log.Fatalf("Fail to serve: %v", err)
	}
}
