package TestServer

import (
	"TestClient"
	"encoding/json"
	"fmt"
	"net/http"
	"utils"
)

func RoutineRegisterClient(w http.ResponseWriter, r *http.Request) {
	// dispatch the register client request
	var request TestClient.RegisterClientInformation_HTTPJSON
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		utils.NetworkWarnings(r, "server::RegisterClient", fmt.Sprintf("Format error: %s", err.Error()))
		_ = json.NewEncoder(w).Encode(TestClient.RegisterServerReply_HTTPJSON{
			StatusCode:        1,
			Heartbeat:         0,
			ConnPort:          0,
			IdentificationKey: "",
		})
		return
	}
	// match the request with the given keys
	if request.Identities != identificationToken{
		utils.NetworkWarnings(r, "server::RegisterClient", fmt.Sprintf("Identification Error: %s <-> %s (required)",
			request.Identities, identificationToken))
		_ = json.NewEncoder(w).Encode(TestClient.RegisterServerReply_HTTPJSON{
			StatusCode:        2,
			Heartbeat:         0,
			ConnPort:          0,
			IdentificationKey: "",
		})
		return
	}
	// identities matched, make a port now
	// NOTICE: Go into the critical region(TestServer/entry.go:portList)
	portListMutex.Lock()
	if len(portList) == 0 {
		portListMutex.Unlock()
		utils.NetworkWarnings(r, "server::RegisterClient", "The port pool is empty.")
		_ = json.NewEncoder(w).Encode(TestClient.RegisterServerReply_HTTPJSON{
			StatusCode:        100,
			Heartbeat:         0,
			ConnPort:          0,
			IdentificationKey: "",
		})
		return
	}
	availPort := portList[0]
	portList = portList[1:]
	portListMutex.Unlock()
	// NOTICE: Go out from the critical region(TestServer/entry.go:portList)

	// Dispatch the data: target port, heartbeat requirement, NUID as identification key
	// Critical region(TestServer/entry.go:identificationKey)
	identificationKeyMutex.Lock()
	idk := identificationKey.Next()
	identificationKeyMutex.Unlock()
	// Critical region(TestServer/entry.go:identificationKey) End
	// Set to the global pool

	// Critical Region(TestServer/entry.go:matchIDK, matchName)
	matchIDKMutex.Lock()
	matchIDK[availPort] = idk
	matchName[availPort] = request.Name
	matchIDKMutex.Unlock()
	// Critical Region(TestServer/entry.go:matchIDK, matchName) End
	_ = json.NewEncoder(w).Encode(TestClient.RegisterServerReply_HTTPJSON{
		StatusCode:        0,
		Heartbeat:         globalHeartBeat,
		ConnPort:          availPort,
		IdentificationKey: idk,
	})
	// TODO: start the corresponding coroutine
	go func(){
		serverMux := http.NewServeMux()
		serverMux.HandleFunc("/heartBeat", RoutineListenerHeartBeat)
		err := http.ListenAndServe(fmt.Sprintf("%s:%d", globalServerBindAddr, availPort), serverMux)
		// Here should not call checkError in order to prevent unexpected halting
		utils.Warnings(fmt.Sprintf("server on port %d", availPort), fmt.Sprintf("Runtime error: %s", err.Error()))
		utils.Warnings(fmt.Sprintf("server on port %d", availPort), "Server stops.")
	}()
}
