package TestServer

import (
	"TestClient"
	"encoding/json"
	"fmt"
	"net/http"
	"utils"
)

func RoutineSubmitJudgeWork(w http.ResponseWriter, r *http.Request) {
	var recvMessage utils.UploadWorkSlice
	err := json.NewDecoder(r.Body).Decode(&recvMessage)
	if err != nil {
		utils.NetworkWarnings(r, "server::RoutineSubmitJudgeWork", fmt.Sprintf("Format error: %s", err.Error()))
		return
	}
	matchIDKMutex.Lock()
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
	gitHashListMux.Lock()

	judgeElement := judgeStatus[recvMessage.GitHash]
	for _, element := range recvMessage.Cases{
		targetStatus := judgeElement.Phase[judgeElement.CurrentPhase].Running[element.ResultId]
		targetStatus.Testcase = TestcaseFormat{
			IsAssertion:   element.IsAssertion,
			ResultId:      element.ResultId,
			SourceCode:    element.SourceCode,
			Assertion:     element.Assertion,
			TimeLimit:     element.TimeLimit,
			InstLimit:     element.InstLimit,
			MemoryLimit:   element.MemoryLimit,
			Testcase:      element.Testcase,
			InputContext:  element.InputContext,
			OutputContext: element.OutputContext,
			OutputCode:    element.OutputCode,
			BasicType:     element.BasicType,
			Verdict:       element.Verdict,
			StdOutMessage: element.StdOutMessage,
			StdErrMessage: element.StdErrMessage,
			Runtime:       element.Runtime,
			InstsCount:    element.InstsCount,
			RunningStdOut: element.RunningStdOut,
			RavelMessage:  element.RavelMessage,
			ErrorMessage:  element.ErrorMessage,
		}
		if element.Verdict == VERDICT_CORRECT{
			judgeElement.Phase[judgeElement.CurrentPhase].Success = append(judgeElement.Phase[judgeElement.CurrentPhase].Success, targetStatus)
		} else {
			judgeElement.Phase[judgeElement.CurrentPhase].Fail = append(judgeElement.Phase[judgeElement.CurrentPhase].Fail, targetStatus)
		}
		delete(judgeElement.Phase[judgeElement.CurrentPhase].Running, element.ResultId)
	}
	// Check for the stage information
	if len(judgeElement.Phase[judgeElement.CurrentPhase].Running) == 0 && len(judgeElement.Phase[judgeElement.CurrentPhase].Pending) == 0 {
		if (len(judgeElement.Phase[judgeElement.CurrentPhase].Success) != 0) && (len(judgeElement.Phase[judgeElement.CurrentPhase].Fail) == 0) {
			judgeElement.CurrentPhase = judgeElement.CurrentPhase + 1

			if judgeElement.CurrentPhase == 3{
				// Finish all the points, remove from the list
				idx := utils.FindIdx(gitHashList, judgeElement.GitHash)
				gitHashList = append(append([]string{}, gitHashList[0:idx]...), gitHashList[idx +1:]...)
				delete(judgeStatus, judgeElement.GitHash)
				gitHashListMux.Unlock()
				// TODO: Database: add finish job to database
				return
			}

			stageTestcaseMutex[judgeElement.CurrentPhase].Lock()

			for _, ele := range stageTestcase[judgeElement.CurrentPhase] {
				judgeElement.Phase[judgeElement.CurrentPhase].Pending = append(judgeElement.Phase[judgeElement.CurrentPhase].Pending, JudgeStatus{
					Phase:    judgeElement.CurrentPhase,
					Status:   0,
					CaseId:   "",
					ResultId: JudgeIdKey.Next(),
					JudgerId: "",
					Testcase: ele,
				})
			}

			stageTestcaseMutex[judgeElement.CurrentPhase].Unlock()
			judgeStatus[judgeElement.GitHash] = judgeElement

		} else if (len(judgeElement.Phase[judgeElement.CurrentPhase].Success) != 0) && (len(judgeElement.Phase[judgeElement.CurrentPhase].Fail) != 0) {
			// remove from the list
			idx := utils.FindIdx(gitHashList, judgeElement.GitHash)
			gitHashList = append(append([]string{}, gitHashList[0:idx]...), gitHashList[idx+1:]...)
			delete(judgeStatus, judgeElement.GitHash)
			gitHashListMux.Unlock()
			return
			// TODO: add to database: failed to judge
		}
	}
	gitHashListMux.Unlock()
	return

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
	// Critical Region(TestServer/entry.go:availTimeStamp)
	availTimeStampMutex.Lock()
	availTimeStamp[recvMessage.Port] = recvMessage.Timestamp
	availTimeStampMutex.Unlock()
	// Critical Region(TestServer/entry.go:availTimeStamp) End
	// Critical Region: gitHashList(JudgePool)
	gitHashListMux.Lock()
	if len(gitHashList) == 0 {
		gitHashListMux.Unlock()
		// Critical Region: gitHashList(JudgePool) End
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
		// Critical Region: gitHashList(JudgePool) End
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
		// TODO: Detect whether the working queue has no element
		if maxCnt == 0 {
			gitHashListMux.Unlock()
			// Critical Region: gitHashList(JudgePool) End
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
			sendWork.Cases[idx] = e.Phase[e.CurrentPhase].Pending[idx].Testcase
			e.Phase[e.CurrentPhase].Pending[idx].Status = INRunning
			e.Phase[e.CurrentPhase].Pending[idx].JudgerId = serverName
			e.Phase[e.CurrentPhase].Pending[idx].ResultId = sendWork.Cases[idx].ResultId
			e.Phase[e.CurrentPhase].Running[sendWork.Cases[idx].ResultId] = e.Phase[e.CurrentPhase].Pending[idx]
		}
		e.Phase[e.CurrentPhase].Pending = append(e.Phase[e.CurrentPhase].Pending[maxCnt:])
		if len(e.Phase[e.CurrentPhase].Pending) == 0 {
			gitHashList = append(gitHashList[1:], headHash)
		}
		gitHashListMux.Unlock()
		// Critical Region: gitHashList(JudgePool) End
		// TODO: fill the data

		_ = json.NewEncoder(w).Encode(sendWork)
		return
	}
}
