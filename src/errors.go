package main

import (
	"errors"
	"regexp"
	"strings"
)

// TODO: implement status in gRPC response: https://pkg.go.dev/google.golang.org/grpc/status
var (
	ErrSandboxBusy    error = errors.New("SANDBOX_BUSY")
	ErrInternalServer error = errors.New("INTERNAL_SERVER_ERROR")
)

const CodeStorageFolder string = "/sandbox/code"

// handle errors with multiple lines like below:
//
//	e.g.
//
// # command-line-arguments\ntmp/1705519277639.go:7:2: \"os\" imported and not used\ntmp/1705519277639.go:14:19: undefined: time.Secon\n
func CleanError(e string) string {
	toRemove := []string{
		"# command-line-arguments",
		"# [command-line-arguments]",
		"vet: ",
		"isolate",
	}
	r := regexp.MustCompile(CodeStorageFolder + `/[0-9]+.go`)
	e = r.ReplaceAllString(e, "main.go")

	// <standard input> comes from directly running code in goimports and vet
	for i := 0; i < len(toRemove); i++ {
		e = strings.Replace(e, toRemove[i], "", -1)
	}
	return strings.Replace(e, "<standard input>", "main.go", -1)
}
