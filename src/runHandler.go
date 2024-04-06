package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	pb "github.com/nirdosh17/go-sandbox/proto/lib/go"
)

const execTimeoutDuration = time.Second * 60

func (s *Server) RunCode(in *pb.RunRequest, stream pb.GoSandboxService_RunCodeServer) error {
	s.Logger.Println("RunCode executed with session id", in.SessionId)

	ctx, cancel := context.WithTimeout(context.Background(), execTimeoutDuration)
	defer cancel()

	// saving code as file
	codeID, err := WriteToTempFile([]byte(in.Code))
	if err != nil {
		log.Println("error saving code as file:", err)
		stream.Send(&pb.RunResponse{Output: "", IsError: true, ExecErr: "server error: failed to save code"})
		return nil
	}

	fname := fmt.Sprintf("code/%v.go", codeID)
	defer os.Remove(fname)

	var boxId int
	box, present := s.Sandbox.GetExistingSandbox(in.SessionId)
	if !present {
		box, serr := s.Sandbox.Reserve(in.SessionId)
		boxId = box.Id
		if serr != nil {
			log.Println("selected sandbox:", box.Id, "err:", serr)
		}
	} else {
		boxId = box.Id
		log.Println("existing sandbox", box.Id, "selected for user", in.SessionId)
	}

	// if session id e.g.(user is) is absent, a random sandbox assigned
	// this means if user creates a file during first execution, then the file may not be present in the next execution as it may run in different sandbox
	if in.SessionId == "" {
		defer s.Sandbox.Release(in.SessionId)
	} else {
		s.Sandbox.UpdateUsed(in.SessionId)
	}

	cmd := exec.CommandContext(ctx,
		"isolate",
		fmt.Sprintf("--box-id=%v", boxId),
		// -f, --fsize=<size>	Max size (in KB) of files that can be created
		"--dir=/app/code",
		// give read write access to the go cache dir as it needs to be cleaned
		"--dir=/root/.cache/go-build:rw",
		// if sandbox is busy, wait instead of returning error right away
		// instead of serving 25/100 requests in 10 sandbox, it's gonna serve all
		"--wait",
		// to keep the child process in parent’s network namespace and communicate with the outside world
		// "--share-net",
		"--processes=100",
		// unlimited open files
		"--open-files=0",
		"--env=GOROOT",
		"--env=GOPATH",
		"--env=GO111MODULE=on",
		"--env=HOME",
		// makes commands visible in the sandbox e.g. 'ls', 'echo' or other installed command
		"--env=PATH",
		// log package writes to stderr instead of stdout, so we need to redirect this to stdout.
		// only exit code determines if the program ran successfully or not
		"--stderr-to-stdout",
		"--run",
		"--",
		"/usr/local/go/bin/go",
		"run",
		fmt.Sprintf("/app/code/%v.go", codeID),
	)

	cmd.WaitDelay = execTimeoutDuration

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	stdoutpipe, err := cmd.StdoutPipe()
	if err != nil {
		log.Println("error creating stdout pipe:", stderr.String())
		stream.Send(&pb.RunResponse{Output: "", IsError: true, ExecErr: "server error", Timestamp: time.Now().UnixMilli()})
		return nil
	}

	err = cmd.Start()
	if err != nil {
		log.Println("error running code:", err, stderr.String())
		stream.Send(&pb.RunResponse{Output: "", IsError: true, ExecErr: "server error", Timestamp: time.Now().UnixMilli()})
		return nil
	}

	scanner := bufio.NewScanner(stdoutpipe)
	for scanner.Scan() {
		m := scanner.Text()
		stream.Send(&pb.RunResponse{Output: CleanError(m), IsError: false, ExecErr: "", Timestamp: time.Now().UnixMilli()})
	}

	// command exec errors will be present here
	err = cmd.Wait()
	if err != nil {
		if strings.Contains(stderr.String(), "box is currently in use by another process") {
			log.Println("[cmd error] ", err, stderr.String())
			return ErrSandboxBusy
		}

		// status 2 is is related to isolate, status 1 is related to go code
		if strings.Contains(err.Error(), "exit status 2") {
			log.Println("[cmd error] ", err, stderr.String())
			return ErrInternalServer
		}

		// split and sent stream of events instead
		for _, errStr := range strings.Split(stderr.String(), "\n") {
			cleaned := CleanError(errStr)
			if cleaned != "" {
				stream.Send(&pb.RunResponse{Output: "", IsError: true, ExecErr: cleaned, Timestamp: time.Now().UnixMilli()})
			}
		}
		log.Println("[cmd error] ", err, stderr.String())
	}

	return nil
}
