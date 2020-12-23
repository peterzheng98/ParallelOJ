package TestServer

import "sync"

var INRunning = 1
var NORunning = 0
var FailRunning = 2
var SuccRunning = 3

type JudgeStatus struct {
	Phase    int    `json:"phase"`
	Status   int    `json:"status"`
	CaseId   string `json:"case_id"`
	ResultId string `json:"result_id"`
	JudgerId string `json:"judger_id"`
}

type JudgePhase struct {
	Phase   int           `json:"phase"`
	Pending []JudgeStatus `json:"pending"`
	Running []JudgeStatus `json:"running"`
	Success []JudgeStatus `json:"success"`
	Fail    []JudgeStatus `json:"fail"`
}

type JudgeElement struct {
	User         string       `json:"user"`
	GitRepo      string       `json:"git_repo"`
	GitHash      string       `json:"git_hash"`
	RequestTime  int64        `json:"request_time"`
	Build        JudgeStatus  `json:"build"`
	CurrentPhase int          `json:"current_phase"`
	Phase        []JudgePhase `json:"phase"`
}

var (
	gitHashList    []string
	judgeStatus    map[string]JudgeElement
	gitHashListMux sync.Mutex
)

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
