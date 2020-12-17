package TestServer

import (
	"fmt"
	"github.com/nats-io/nuid"
	"net/http"
	"sync"
	"utils"
)

// the identification token is set when the server is starting.
// no need for mutex
var identificationToken = "123"

var (
	portListMutex sync.Mutex
	portList []int
)
var (
	identificationKey = nuid.New()
	identificationKeyMutex sync.Mutex
)

var (
	matchIDK map[int]string
	matchName map[int]string
	matchIDKMutex sync.Mutex
)

var globalHeartBeat = 0
var globalServerBindAddr = ""
func ServerEntry(ConfigPath string){
	utils.Logs("server", fmt.Sprintf("Startup with server mode with file %s.", ConfigPath))
	servConfig := utils.ReadFromServerJSON(ConfigPath)
	utils.Logs("server", "Get server configuration file")
	utils.Logs("server", utils.ServerJSONToString(servConfig))
	identificationToken = servConfig.ServerIdentificationToken
	globalHeartBeat = servConfig.Heartbeat
	globalServerBindAddr = servConfig.BindAddr
	// make port list, here no need for mutex
	for i := 0; i < servConfig.MaximizeClient; i++ {
		portList = append(portList, i + servConfig.StartPort)
	}
	utils.Logs("server", fmt.Sprintf("make port: from %d to %d", portList[0], portList[len(portList) - 1]))
	// bind to the path: http://addr:port/
	// Notice that for different port, we should bind to different multiplexing handler for the same URI
	// e.g.: http://addr:port1/ and http://addr:port2/ should use different handler function

	go func() {
		serverMuxA := http.NewServeMux()
		serverMuxA.HandleFunc("/", RoutineRegisterClient)
		err := http.ListenAndServe(fmt.Sprintf("%s:%d", servConfig.BindAddr, servConfig.Port), serverMuxA)
		utils.CheckError(err)
	}()
}
