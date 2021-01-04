package TestClient

import (
	"bufio"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"os/exec"
	"time"
)

func judgeSemantic(source string, dockerTags string, expectedAssertion bool, timeout float32, memoryOut int) *SemanticJudgeResult {
	se := SemanticJudgeResult{}
	// dump the file out to input.mx
	inputPath := fmt.Sprintf("%s/input.mx", dataPath)
	srcDec, _ := base64.StdEncoding.DecodeString(source)
	_ = ioutil.WriteFile(inputPath, srcDec, 0644)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()
	// docker cmd
	cmd := exec.CommandContext(ctx, "docker", "run", "--rm", "--cpus=1", "-m", fmt.Sprintf("%dM", memoryOut), "-v",
		fmt.Sprintf("%s:%s", dataPath, "/mounted"), dockerTags, "/bin/bash", "/JudgeSemantic.bash")
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	outputReader := bufio.NewReader(stdout)
	errorReader := bufio.NewReader(stderr)

	stdoutStr := ""
	stderrStr := ""
	startTime := time.Now()
	_ = cmd.Start()
	err := cmd.Wait()
	elapsed := time.Since(startTime).Milliseconds()
	// read stdout
	for {
		stdoutSlice, err := outputReader.ReadString('\n')
		if err == io.EOF {
			break
		}
		stdoutStr = stdoutStr + "\n" + stdoutSlice
	}

	// read stderr
	for {
		stderrSlice, err := errorReader.ReadString('\n')
		if err == io.EOF{
			break
		}
		stderrStr = stderrStr + "\n" + stderrSlice
	}

	// read stdout
	for {
		stdoutSlice, err := outputReader.ReadString('\n')
		if err == io.EOF{
			break
		}
		stdoutStr = stdoutStr + "\n" + stdoutSlice
	}

	if err == context.DeadlineExceeded{
		se.Verdict = VERDICT_TLE
	} else {
		judgeAssertion := err == nil
		if judgeAssertion == expectedAssertion{
			se.Verdict = VERDICT_CORRECT
		} else {
			se.Verdict = VERDICT_WRONG
		}
	}
	se.InstsCount = 0
	se.Runtime = float32(elapsed) / 1000.0
	se.StdErrMessage = base64.StdEncoding.EncodeToString([]byte(stderrStr))
	se.StdOutMessage = base64.StdEncoding.EncodeToString([]byte(stdoutStr))
	return &se
}
