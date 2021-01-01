package TestClient

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
	"utils"
)

var ClientStatus = 1
var trust_key = ""
var CliRequestCount = 1

// Danger Zone!!
var UNTRUST = false

func heartbeat(heartbeat int, addr string, port int, idk string) {
	for {
		heartbeatMessage := HeartBeatClient{
			Timestamp:  utils.MakeTimestamp(),
			Status:     ClientStatus,
			Additional: "", // TODO: add current judge status
			Port:       port,
			IDK:        idk,
			Operation:  0,
		}
		// TODO: check the server is alive
		heartbeatMess := ServerHeartBeatReply_HTTPJSON{}
		heartbeatBytes := utils.SendHTTPRequestJSON(addr, port, heartbeatMessage)
		err := json.Unmarshal(heartbeatBytes, &heartbeatMess)
		if err != nil {
			utils.Warnings("client", "Runtime error: heartbeat package error")
			utils.CheckError(err)
		}
		if heartbeatMess.Status != 0 {
			utils.CheckError(errors.New("Fail to make heartbeat package"))
		}
		if !UNTRUST && heartbeatMess.TrustKey != trust_key {
			utils.Warnings("client", "Runtime error: UNTRUSTABLE SERVER!!")
			utils.Warnings("client", "Received key not equal to local.")
			utils.CheckError(errors.New("Untrustable server: Receive non-equal key"))
		} else if heartbeatMess.TrustKey != trust_key {
			utils.Warnings("client", "Runtime error: UNTRUSTABLE SERVER!!")
			utils.Warnings("client", "Received key not equal to local. Configuration set to start up in untrust mode.")
		}
		time.Sleep(time.Duration(heartbeat) * time.Second)
	}
}

func makeJudge(addr string, port int, idk string) {
	for {
		receiveMessage := ClientRequestJudge{
			Timestamp:    utils.MakeTimestamp(),
			Status:       ClientStatus,
			Additional:   "",
			Port:         port,
			IDK:          idk,
			Operation:    1,
			RequestCount: CliRequestCount,
		}

		_ = utils.SendHTTPRequestJSON(addr, port, receiveMessage)
	}
}

func ClientEntry(ConfigPath string) {
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
	if registerReply.StatusCode != 0 {
		utils.CheckError(errors.New(fmt.Sprintf("Status code %d received.", registerReply.StatusCode)))
		// TODO: For no port exception(ErrCode: 100), wait for an interval and retry
		return
	}
	utils.Logs("client", "Judge start.")
	// Make long connection to the heartbeat
	go heartbeat(registerReply.Heartbeat, cliConfig.ServerAddr, registerReply.ConnPort, registerReply.IdentificationKey)
	go makeJudge(cliConfig.ServerAddr, registerReply.ConnPort, registerReply.IdentificationKey)
}
