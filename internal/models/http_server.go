package models

import (
	"Talos/conf"
	"encoding/json"
	"log"
	"net/http"
)

func operate(opChan chan<- *conf.Operation) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Body == nil {
			http.Error(w, "Request body empty", 400)
			return
		}
		defer r.Body.Close()
		var co conf.Operation
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&co)
		if err != nil {
			http.Error(w, "Parsing body error", 400)
			log.Println("Parsing body error:", err.Error())
			log.Println("Body:", r.Body)
			return
		}
		opChan <- &co
		var re conf.ReturnData
		re.ReturnCode = 0
		_ = json.NewEncoder(w).Encode(re)
	}
}
