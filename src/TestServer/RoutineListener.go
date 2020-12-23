package TestServer

import (
	"TestClient"
	"encoding/json"
	"fmt"
	"net/http"
	"utils"
)

func (ports *Ports) RoutineListenerHeartBeat(w http.ResponseWriter, r *http.Request) {
	var recvMess TestClient.HeartBeatClient
	err := json.NewDecoder(r.Body).Decode(&recvMess)
	if err != nil {
		utils.NetworkWarnings(r,
			fmt.Sprintf("server::RoutineListenerHeartBeat(%d)", ports.Port),
			fmt.Sprintf("Format error: %s", err.Error()))
		_ = json.NewEncoder(w).Encode(TestClient.ServerHeartBeatReply_HTTPJSON{
			Status:   1,
			TrustKey: "",
		})
		return
	}
	// Critical Region(TestServer/entry.go:matchIDK, matchName)
	matchIDKMutex.Lock()
	idk := matchIDK[ports.Port]
	matchIDKMutex.Unlock()
	// Critical Region(TestServer/entry.go:matchIDK, matchName) End
	if recvMess.IDK != idk {
		utils.NetworkWarnings(r,
			fmt.Sprintf("server::RoutineListenerHeartBeat(%d)", ports.Port),
			fmt.Sprintf("IDK Mismatch: Expected: %s, Receive: %s", idk, recvMess.IDK))
		_ = json.NewEncoder(w).Encode(TestClient.ServerHeartBeatReply_HTTPJSON{
			Status:   2,
			TrustKey: "",
		})
		return
	}
	// Critical Region(TestServer/entry.go:availTimeStamp)
	availTimeStampMutex.Lock()
	availTimeStamp[ports.Port] = recvMess.Timestamp
	availTimeStampMutex.Unlock()
	// Critical Region(TestServer/entry.go:availTimeStamp) End
	_ = json.NewEncoder(w).Encode(TestClient.ServerHeartBeatReply_HTTPJSON{
		Status:   0,
		TrustKey: trust_key,
	})
}
