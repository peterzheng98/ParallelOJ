package TestClient

type ClientConfig_FileJSON struct {
	Mode       string `json:"mode"`
	ServerAddr string `json:"server_addr"`
	Port       int    `json:"port"`
	Identities string `json:"identities"`
	Name       string `json:"name"`
	Pwd        string `json:"pwd"`
}

type RegisterClientInformation_HTTPJSON struct {
	Identities string `json:"identities"`
	Name       string `json:"name"`
	Pwd        string `json:"pwd"`
}

type RegisterServerReply_HTTPJSON struct {
	StatusCode        int    `json:"status_code"`
	Heartbeat         int    `json:"heartbeat"`
	ConnPort          int    `json:"conn_port"`
	IdentificationKey string `json:"identification_key"`
}

type HeartBeatClient struct {
	Timestamp int64 `json:"timestamp"`
	Status    int `json:"status"`
	Additional string `json:"additional"`
}
