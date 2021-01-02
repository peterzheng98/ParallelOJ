package TestClient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/go-git/go-git"
	_ "github.com/go-git/go-git"
	"github.com/go-git/go-git/plumbing"
	_ "github.com/go-git/go-git/storage/memory"
	"io"
	"time"
	"utils"
)

var ClientStatus = 1
var trust_key = ""
var CliRequestCount = 1

var localClonePath = ""

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


		reply := utils.SendHTTPRequestJSON(addr, port, receiveMessage)
		replyMess := DispatchedWorkSlice{}
		err := json.Unmarshal(reply, &replyMess)
		if err != nil {
			utils.Warnings("client", "Runtime error: workload package error")
			utils.CheckError(err)
		}

		// Todo: 1. Check if there is no work at all

		imageMakeName := fmt.Sprintf("%s:%s", replyMess.User, replyMess.GitHash[0:8])
		var imageFound bool
		// 2. Check the image exists
		if replyMess.PhaseId != -1{
			ctx := context.Background()
			cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
			if err != nil{
				utils.CheckError(err)
			}
			imageList, err := cli.ImageList(ctx, types.ImageListOptions{
				All: false,
			})

			imageFound = false
			utils.CheckError(err)
			for _, image := range imageList{
				// Search all tags in the images
				for _, subTag := range image.RepoTags{
					if subTag == imageMakeName{
						imageFound = true
						break
					}
				}
				if imageFound {
					break
				}
			}
		}
		var buildSuccess = true
		if (replyMess.PhaseId == -1) || (!imageFound){
			r, err := git.PlainClone(localClonePath, false, &git.CloneOptions{
				URL: replyMess.GitRepo,
			})
			w, err := r.Worktree()
			if err != nil {
				utils.Warnings(fmt.Sprintf("TestClient:[Clone user %s, git address: %s, target hash: %s]", replyMess.User, replyMess.GitRepo, replyMess.GitHash), err.Error())
				// TODO: reply: bad package
				buildSuccess = false
			}
			err = w.Checkout(&git.CheckoutOptions{
				Hash: plumbing.NewHash(replyMess.GitHash),
			})
			if err != nil {
				utils.Warnings(fmt.Sprintf("TestClient:[Build user %s, git address: %s, target hash: %s]", replyMess.User, replyMess.GitRepo, replyMess.GitHash), err.Error())
				// TODO: reply: bad package
				buildSuccess = false
			}
			// TODO: make image
			ctx := context.Background()
			cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
			if err != nil {
				utils.CheckError(err)
			}
			cli.ImageBuild(ctx, io.ByteReader(), types.ImageBuildOptions{
				Tags:           nil,
				SuppressOutput: false,
				RemoteContext:  "",
				NoCache:        false,
				Remove:         false,
				ForceRemove:    false,
				PullParent:     false,
				Isolation:      "",
				CPUSetCPUs:     "",
				CPUSetMems:     "",
				CPUShares:      0,
				CPUQuota:       0,
				CPUPeriod:      0,
				Memory:         0,
				MemorySwap:     0,
				CgroupParent:   "",
				NetworkMode:    "",
				ShmSize:        0,
				Dockerfile:     "",
				Ulimits:        nil,
				BuildArgs:      nil,
				AuthConfigs:    nil,
				Context:        nil,
				Labels:         nil,
				Squash:         false,
				CacheFrom:      nil,
				SecurityOpt:    nil,
				ExtraHosts:     nil,
				Target:         "",
				SessionID:      "",
				Platform:       "",
				Version:        "",
				BuildID:        "",
				Outputs:        nil,
			})
		}
		if !buildSuccess{
			// todo: send bad package
		} else {

		}



		time.Sleep(time.Duration(1) * time.Second)
	}
}

func ClientEntry(ConfigPath string) {
	// load configs
	utils.Logs("client", fmt.Sprintf("Startup with client mode with file %s.", ConfigPath))
	cliConfig := utils.ReadFromClientJSON(ConfigPath)
	localClonePath = cliConfig.PathPrefix
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
