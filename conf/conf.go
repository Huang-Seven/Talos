package conf

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type AgentConf struct {
	LocalAddr      string `mapstructure:"local_addr"`
	LocalPort      int    `mapstructure:"local_port"`
	ConfDir        string `mapstructure:"conf_dir"`
	ServerAddr     string `mapstructure:"server_addr"`
	ServerPort     int    `mapstructure:"server_port"`
	StartDir       string `mapstructure:"start_dir"`
	CheckDir       string `mapstructure:"check_dir"`
	RunDir         string `mapstructure:"run_dir"`
	HostName       string `mapstructure:"host_name"`
	MonitorConfDir string `mapstructure:"monitor_conf_dir"`
}

type ServerConf struct {
	//LocalAddr string
	//AgentPort int
	LocalPort int
	ConfDir   string
	//RunDir  string
	MySQL MySQLConfig
}

type MySQLConfig struct {
	MysqlHost string
	MysqlPort int
	MysqlUser string
	MysqlPwd  string
	MysqlDb   string
}

type Operation struct {
	Action int // `0 actions,1 update data,2 reload pc from server`
	OperationAgent
	OperationServer
}

type OperationAgent struct {
	FrequencyMonitor, FrequenceCollect, MonitorSleep, Maintain int
}

type OperationServer struct {
}

type ReturnData struct {
	ReturnCode int
	ReturnData interface{}
}

func (sc *ServerConf) LoadConf() {
	log.Printf("Server load conf: %v", strings.TrimSuffix(sc.ConfDir, "/")+"/server.toml")
	vs := viper.New()
	vs.AddConfigPath(sc.ConfDir)
	vs.SetConfigName("server")
	vs.SetConfigType("toml")
	err := vs.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf(time.Now().Format("2006-01-02 15:04:05"),
			"Fatal error config file: %s \n", err))
	}

	sc.MySQL.MysqlUser = vs.GetString("mysql.user")
	sc.MySQL.MysqlPwd = vs.GetString("mysql.password")
	sc.MySQL.MysqlHost = vs.GetString("mysql.host")
	sc.MySQL.MysqlPort = vs.GetInt("mysql.port")
	sc.MySQL.MysqlDb = vs.GetString("mysql.db")
}

func (ac *AgentConf) LoadConf() {
	log.Printf("Agent load conf: %v", strings.TrimSuffix(ac.ConfDir, "/")+"/agent.toml")
	v := viper.New()
	v.AddConfigPath(ac.ConfDir)
	v.SetConfigName("agent")
	v.SetConfigType("toml")
	err := v.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf(time.Now().Format("2006-01-02 15:04:05"),
			"Fatal error config file: %s \n", err))
	}
	err = v.Unmarshal(ac)
	if err != nil {
		log.Println(err)
	}
}
