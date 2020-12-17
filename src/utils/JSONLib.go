package utils

import (
	"TestClient"
	"encoding/json"
	"fmt"
	"io/ioutil"
)

func ReadFromClientJSON(filePath string) TestClient.ClientConfig_FileJSON {
	data, err := ioutil.ReadFile(filePath)
	CheckError(err)
	client_config := &TestClient.ClientConfig_FileJSON{}
	err = json.Unmarshal(data, &client_config)
	CheckError(err)
	return *client_config
}

func ClientJSONToString(fileJSON TestClient.ClientConfig_FileJSON) string {
	return fmt.Sprintf("Connect to %s:%s, Client Identities: %s, name: %s", fileJSON.ServerAddr, fileJSON.Port, fileJSON.Identities, fileJSON.Name)
}
