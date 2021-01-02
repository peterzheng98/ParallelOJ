package TestClient

type ClientConfig_FileJSON struct {
	Mode          string `json:"mode"`
	ServerAddr    string `json:"server_addr"`
	Port          int    `json:"port"`
	Identities    string `json:"identities"`
	Name          string `json:"name"`
	Pwd           string `json:"pwd"`
	TrustKey      string `json:"trust_key"`
	PathPrefix    string `json:"path_prefix"`
	BaseImageName string `json:"base_image_name"`
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
	Timestamp  int64  `json:"timestamp"`
	Status     int    `json:"status"`
	Additional string `json:"additional"`
	Port       int    `json:"port"`
	IDK        string `json:"idk"`
	Operation  int    `json:"operation"`
}

type ClientRequestJudge struct {
	Timestamp    int64  `json:"timestamp"`
	Status       int    `json:"status"`
	Additional   string `json:"additional"`
	Port         int    `json:"port"`
	IDK          string `json:"idk"`
	Operation    int    `json:"operation"`
	RequestCount int    `json:"request_count"`
}

type ServerHeartBeatReply_HTTPJSON struct {
	Status   int    `json:"status"`
	TrustKey string `json:"trust_key"`
}

type Ports struct {
	Port              int    `json:"port"`
	IdentificationKey string `json:"identification_key"`
}

type TestcaseFormat struct {
	IsAssertion bool    `json:"is_assertion"`
	ResultId    string  `json:"result_id"`
	SourceCode  string  `json:"source_code"`
	Assertion   bool    `json:"assertion"`
	TimeLimit   float32 `json:"time_limit"`
	InstLimit   int     `json:"inst_limit"`
	MemoryLimit int     `json:"memory_limit"`
	Testcase    string  `json:"testcase"`
	/* - For codegen / optimize only, requires run here- */
	InputContext  string `json:"input_context"`
	OutputContext string `json:"output_context"`
	OutputCode    int    `json:"output_code"`
	BasicType     int    `json:"basic_type"`
}

type DispatchedWorkSlice struct {
	User     string           `json:"user"`
	GitRepo  string           `json:"git_repo"`
	GitHash  string           `json:"git_hash"`
	PhaseId  int              `json:"phase_id"` // -1 as compile, 0-n as index
	WorkCnt  int              `json:"work_cnt"`
	Cases    []TestcaseFormat `json:"cases"`
	TrustKey string           `json:"trust_key"`
}
