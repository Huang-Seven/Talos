package models

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Proc struct {
	Tid     int
	ProcNum int
	MailTo  string
	Core    bool
	Alive   bool
	Ts1     time.Time
	Ts2     time.Time
	Sp      time.Duration
	Modinfo
}

func ReadProc(path, name string) (*Proc, error) {
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
	err = v.Unmarshal(pc)
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
