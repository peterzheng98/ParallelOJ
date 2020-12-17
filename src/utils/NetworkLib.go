package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func SendHTTPRequestJSON(addr string, port int, v interface{}) []byte {
	bytesData, err := json.Marshal(v)
	CheckError(err)
	url := fmt.Sprintf("http://%s:%d", addr, port)
	reader := bytes.NewReader(bytesData)
	request, err := http.NewRequest("POST", url, reader)
	CheckError(err)
	request.Header.Set("Content-Type", "application/json;charset=UTF-8")
	client := http.Client{}
	resp, err := client.Do(request)
	CheckError(err)
	respBytes, err := ioutil.ReadAll(resp.Body)
	CheckError(err)
	return respBytes
}
