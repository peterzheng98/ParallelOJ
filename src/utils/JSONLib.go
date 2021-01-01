package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

func ReadFromClientJSON(filePath string) ClientConfig_FileJSON {
	data, err := ioutil.ReadFile(filePath)
	CheckError(err)
	clientConfig := &ClientConfig_FileJSON{}
	err = json.Unmarshal(data, &clientConfig)
	CheckError(err)
	return *clientConfig
}

func ClientJSONToString(fileJSON ClientConfig_FileJSON) string {
	return fmt.Sprintf("Connect to %s:%d, Client Identities: %s, name: %s", fileJSON.ServerAddr, fileJSON.Port, fileJSON.Identities, fileJSON.Name)
}

func ReadFromServerJSON(filePath string) ServerConfig_FileJSON {
	data, err := ioutil.ReadFile(filePath)
	CheckError(err)
	serverConfig := &ServerConfig_FileJSON{}
	err = json.Unmarshal(data, &serverConfig)
	CheckError(err)
	return *serverConfig
}

func ServerJSONToString(fileJSON ServerConfig_FileJSON) string {
	return fmt.Sprintf("Bind to %s:%d, require heartbeat: %d second, max: %d clients, timeout: %d seconds, SQL Type: %s, SQL path: %s:%d[%s]",
		fileJSON.BindAddr, fileJSON.Port, fileJSON.Heartbeat, fileJSON.MaximizeClient, fileJSON.HeartbeatTimeout, fileJSON.DatabaseType, fileJSON.DatabaseAddr, fileJSON.DatabasePort, fileJSON.DatabasePath)
}