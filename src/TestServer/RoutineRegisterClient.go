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
	
}
