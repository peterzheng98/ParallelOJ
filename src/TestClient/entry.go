package TestClient

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/codeskyblue/go-sh"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/go-git/go-git"
	_ "github.com/go-git/go-git"
	"github.com/go-git/go-git/plumbing"
	_ "github.com/go-git/go-git/storage/memory"
	"io/ioutil"
	"time"
	"utils"
)

var ClientStatus = 1
var trust_key = ""
var CliRequestCount = 1

var localClonePath = ""
var base_image = ""
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
		// todo: add log output when judging

		imageMakeTag := fmt.Sprintf("%s:%s", replyMess.User, replyMess.GitHash[0:8])

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
					if subTag == imageMakeTag {
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
		var buildFailMessage = ""

		if (replyMess.PhaseId == -1) || (!imageFound){
			parentFolderName := fmt.Sprintf("%s_%s", replyMess.User, replyMess.GitHash[0:8])
			DockerbuildPath := fmt.Sprintf("%s/%s", localClonePath, parentFolderName)
			clonePath := fmt.Sprintf("%s/%s/src", localClonePath, parentFolderName)
			DockerfilePath := fmt.Sprintf("%s/%s/Dockerfile", localClonePath, parentFolderName)
			JudgeSemanticPath := fmt.Sprintf("%s/%s/JudgeSemantic.bash", localClonePath, parentFolderName)
			JudgeCodegenPath := fmt.Sprintf("%s/%s/JudgeCodegen.bash", localClonePath, parentFolderName)
			JudgeOptimizePath := fmt.Sprintf("%s/%s/JudgeOptimize.bash", localClonePath, parentFolderName)
			r, err := git.PlainClone(clonePath, false, &git.CloneOptions{
				URL: replyMess.GitRepo,
			})
			w, err := r.Worktree()
			if err != nil {
				utils.Warnings(fmt.Sprintf("TestClient:[Clone user %s, git address: %s, target hash: %s]", replyMess.User, replyMess.GitRepo, replyMess.GitHash), err.Error())
				// reply: bad package
				buildSuccess = false
				buildFailMessage = fmt.Sprintf("Internal Error when fetching repo: %s", replyMess.GitRepo)
			}
			err = w.Checkout(&git.CheckoutOptions{
				Hash: plumbing.NewHash(replyMess.GitHash),
			})
			if err != nil {
				utils.Warnings(fmt.Sprintf("TestClient:[Build user %s, git address: %s, target hash: %s]", replyMess.User, replyMess.GitRepo, replyMess.GitHash), err.Error())
				// reply: bad package
				buildSuccess = false
				buildFailMessage = fmt.Sprintf("Internal Error when checking out the repo to hash: %s", replyMess.GitHash)
			}
			// make image
			JudgeSemanticContent := []byte("cp /mounted/input.mx /src/ && bash /src/semantic.bash")
			_ = ioutil.WriteFile(JudgeSemanticPath, JudgeSemanticContent, 0644)
			JudgeCodegenContent := []byte("cp /mounted/input.mx /src/ && bash codegen.bash && cp /src/output.mx /mounted/")
			_ = ioutil.WriteFile(JudgeCodegenPath, JudgeCodegenContent, 0644)
			JudgeOptimizeContent := []byte("cp /mounted/input.mx /src/ && bash optimize.bash && cp /src/output.mx /mounted/")
			_ = ioutil.WriteFile(JudgeOptimizePath, JudgeOptimizeContent, 0644)
			contents := []byte(fmt.Sprintf("FROM %s\nWORKDIR /src\nCOPY src .\nCOPY *.bash /\nRUN /bin/bash build.bash", base_image))
			err = ioutil.WriteFile(DockerfilePath, contents, 0644)
			// reply: bad package
			if err != nil{
				buildSuccess = false
				buildFailMessage = "Internal Error when creating Dockerfile and corresponding Bash"
			}
			if replyMess.PhaseId == -1{
				output, err := sh.Command("docker", "build", "-t", fmt.Sprintf("%s", imageMakeTag), ".", sh.Dir(DockerbuildPath)).CombinedOutput()
				// TODO: if err != nil: reply bad package
				//utils.CheckError(err)
				buildResult := base64.StdEncoding.EncodeToString(output)
				uploadData := UploadWorkSlice{
					User:         replyMess.User,
					GitRepo:      replyMess.GitRepo,
					GitHash:      replyMess.GitHash,
					PhaseId:      replyMess.PhaseId,
					WorkCnt:      replyMess.WorkCnt,
					Cases:        nil,
					PortsInfo:    Ports{port, idk},
					BuildResult:  buildResult,
					BuildVerdict: 1, // Failed
				}
				if err != nil {
					uploadData.BuildVerdict = 1
					_ = utils.SendHTTPRequestJSON(addr, port, uploadData)
					continue
				} else {
					uploadData.BuildVerdict = 0
					_ = utils.SendHTTPRequestJSON(addr, port, uploadData)
					continue
				}
			} else {
				_, err := sh.Command("docker", "build", "-t", fmt.Sprintf("%s", imageMakeTag), ".", sh.Dir(DockerbuildPath)).CombinedOutput()
				utils.CheckError(err)
				// TODO: There should be error, Uhh?
			}

		}
		if !buildSuccess{
			// todo: send bad package
			buildFailMessage = base64.StdEncoding.EncodeToString([]byte(buildFailMessage))
			uploadData := UploadWorkSlice{
				User:         replyMess.User,
				GitRepo:      replyMess.GitRepo,
				GitHash:      replyMess.GitHash,
				PhaseId:      replyMess.PhaseId,
				WorkCnt:      replyMess.WorkCnt,
				Cases:        nil,
				PortsInfo:    Ports{port, idk},
				BuildResult:  buildFailMessage,
				BuildVerdict: 1, // Failed
			}
			_ = utils.SendHTTPRequestJSON(addr, port, uploadData)
			continue
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
	base_image = cliConfig.BaseImageName
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
