package utils

import (
	"TestClient"
	"TestServer"
	"encoding/json"
	"fmt"
	"io/ioutil"
)

func ReadFromClientJSON(filePath string) TestClient.ClientConfig_FileJSON {
	data, err := ioutil.ReadFile(filePath)
	CheckError(err)
	clientConfig := &TestClient.ClientConfig_FileJSON{}
	err = json.Unmarshal(data, &clientConfig)
	CheckError(err)
	return *clientConfig
}

func ClientJSONToString(fileJSON TestClient.ClientConfig_FileJSON) string {
	return fmt.Sprintf("Connect to %s:%s, Client Identities: %s, name: %s", fileJSON.ServerAddr, fileJSON.Port, fileJSON.Identities, fileJSON.Name)
}

func ReadFromServerJSON(filePath string) TestServer.ServerConfig_FileJSON {
	data, err := ioutil.ReadFile(filePath)
	CheckError(err)
	serverConfig := &TestServer.ServerConfig_FileJSON{}
	err = json.Unmarshal(data, &serverConfig)
	CheckError(err)
	return *serverConfig
}

func ServerJSONToString(fileJSON TestServer.ServerConfig_FileJSON) string {
	return fmt.Sprintf("Bind to %s:%d, require heartbeat: %d second, max: %d clients, timeout: %d seconds, SQL Type: %s, SQL path: %s:%d[%s]",
		fileJSON.BindAddr, fileJSON.Port, fileJSON.Heartbeat, fileJSON.MaximizeClient, fileJSON.HeartbeatTimeout, fileJSON.DatabaseType, fileJSON.DatabaseAddr, fileJSON.DatabasePort, fileJSON.DatabasePath
	)
}