package TestServer

import (
	"TestClient"
	"encoding/json"
	"fmt"
	"net/http"
	"utils"
)

func RoutineRequestJudge(w http.ResponseWriter, r *http.Request) {

}

func RoutineRequestJudgeWork(w http.ResponseWriter, r *http.Request) {
	var recvMessage TestClient.ClientRequestJudge
	err := json.NewDecoder(r.Body).Decode(&recvMessage)
	if err != nil {
		utils.NetworkWarnings(r, "server::RoutineRequestJudgeWork", fmt.Sprintf("Format error: %s", err.Error()))
		return
	}
	matchIDKMutex.Lock()
	var serverName = matchName[recvMessage.Port]
	var expectIDK = matchIDK[recvMessage.Port]
	matchIDKMutex.Unlock()
	if expectIDK != recvMessage.IDK {
		utils.NetworkWarnings(r, fmt.Sprintf("server::RoutineRequestJudgeWork(%d)", recvMessage.Port),
			fmt.Sprintf("IDK not fit. Expected: %s, Receviced: %s", expectIDK, recvMessage.IDK))

		_ = json.NewEncoder(w).Encode(DispatchedWorkSlice{
			User:    "",
			GitRepo: "",
			GitHash: "",
			PhaseId: 0,
			WorkCnt: -1,
			Cases:   nil,
		})
		return
	}
	// TODO: update time stamp of the server
	gitHashListMux.Lock()
	if len(gitHashList) == 0 {
		gitHashListMux.Unlock()
		_ = json.NewEncoder(w).Encode(DispatchedWorkSlice{
			User:    "",
			GitRepo: "",
			GitHash: "",
			PhaseId: 0,
			WorkCnt: 0,
			Cases:   nil,
		})
		return
	}
	headHash := gitHashList[0]
	var e = judgeStatus[headHash]
	if e.Build.Status == NORunning {
		gitHashList = append(gitHashList[1:], headHash)
		var sendWork = DispatchedWorkSlice{
			User:    e.User,
			GitRepo: e.GitRepo,
			GitHash: e.GitHash,
			PhaseId: -1,
			WorkCnt: 1,
			Cases:   nil,
		}
		e.Build.Status = INRunning
		judgeStatus[headHash] = e
		gitHashListMux.Unlock()
		sendWork.Cases = make([]TestcaseFormat, 1)
		sendWork.Cases[0] = TestcaseFormat{
			IsAssertion:   false,
			SourceCode:    "",
			Assertion:     false,
			TimeLimit:     0,
			InstLimit:     0,
			MemoryLimit:   0,
			Testcase:      "",
			InputContext:  "",
			OutputContext: "",
			OutputCode:    0,
			BasicType:     0,
		}
		_ = json.NewEncoder(w).Encode(sendWork)
		return
	} else {
		var maxCnt = len(e.Phase[e.CurrentPhase].Pending)
		if maxCnt > recvMessage.RequestCount {
			maxCnt = recvMessage.RequestCount
		}
		var sendWork = DispatchedWorkSlice{
			User:    e.User,
			GitRepo: e.GitRepo,
			GitHash: e.GitHash,
			PhaseId: e.CurrentPhase,
			WorkCnt: maxCnt,
			Cases:   nil,
		}
		sendWork.Cases = make([]TestcaseFormat, maxCnt)
		var idx int
		for idx = 0; idx < maxCnt; idx++ {
			sendWork.Cases[idx] = TestcaseFormat{
				IsAssertion:   e.CurrentPhase == 0,
				SourceCode:    "",
				Assertion:     false,
				TimeLimit:     0,
				InstLimit:     0,
				MemoryLimit:   0,
				Testcase:      e.Phase[e.CurrentPhase].Pending[idx].CaseId,
				InputContext:  "",
				OutputContext: "",
				OutputCode:    0,
				BasicType:     0,
				ResultId:      JudgeIdKey.Next(),
			}
			e.Phase[e.CurrentPhase].Pending[idx].Status = INRunning
			e.Phase[e.CurrentPhase].Pending[idx].JudgerId = serverName
			e.Phase[e.CurrentPhase].Pending[idx].ResultId = sendWork.Cases[idx].ResultId
			e.Phase[e.CurrentPhase].Running = append(e.Phase[e.CurrentPhase].Running, e.Phase[e.CurrentPhase].Pending[idx])
		}
		e.Phase[e.CurrentPhase].Pending = append(e.Phase[e.CurrentPhase].Pending[maxCnt:])
		if len(e.Phase[e.CurrentPhase].Pending) == 0 {
			gitHashList = append(gitHashList[1:], headHash)
		}
		gitHashListMux.Unlock()
		// TODO: fill the data
		_ = json.NewEncoder(w).Encode(sendWork)
		return
	}
}
