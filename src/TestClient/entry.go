package TestClient

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
	"utils"
)

var ClientStatus = 1

func heartbeat(heartbeat int, addr string, port int) {
	for {
		heartbeatMessage := HeartBeatClient{
			Timestamp: utils.MakeTimestamp(),
			Status:    ClientStatus,
			Additional: "", // TODO: add current judge status
		}
		_ = utils.SendHTTPRequestJSON(addr, port, heartbeatMessage)
		time.Sleep(time.Duration(heartbeat) * time.Second)
	}
}

func makeJudge(){
	for {

	}
}


func ClientEntry(ConfigPath string){
	// load configs
	utils.Logs("client", fmt.Sprintf("Startup with client mode with file %s.", ConfigPath))
	cliConfig := utils.ReadFromClientJSON(ConfigPath)
	utils.Logs("client", "Get client configuration files.")
	utils.Logs("client", utils.ClientJSONToString(cliConfig))
	// test connections and fetch the base config
	// including heartbeat, connection port
	registerClient := RegisterClientInformation_HTTPJSON{
		Identities: cliConfig.Identities,
		Name:       cliConfig.Name,
		Pwd:        cliConfig.Pwd,
	}
	// Make HTTP request
	recv := utils.SendHTTPRequestJSON(cliConfig.ServerAddr, cliConfig.Port, registerClient)
	// Dispatch the received message
	registerReply := RegisterServerReply_HTTPJSON{}
	err := json.Unmarshal(recv, &registerReply)
	utils.CheckError(err)
	if registerReply.StatusCode != 0{
		utils.CheckError(errors.New(fmt.Sprintf("Status code %d received.", registerReply.StatusCode)))
	}
	utils.Logs("client", "Judge start.")
	// Make long connection to the heartbeat
	go heartbeat(registerReply.Heartbeat, cliConfig.ServerAddr, registerReply.ConnPort)
	go makeJudge()
}
