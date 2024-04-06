package main

import (
	"log"
	"net"
	"os"
	"time"

	pb "github.com/nirdosh17/go-sandbox/proto/lib/go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
)

var serviceAddr string = "0.0.0.0:" + os.Getenv("SERVICE_PORT")

var totalSandbox = 10

type Server struct {
	pb.GoSandboxServiceServer
	Sandbox *Sandbox
	Logger  *log.Logger
}

const codeCleanupInterval = 2 * time.Hour

func main() {
	ticker := time.NewTicker(codeCleanupInterval)
	defer ticker.Stop()

	go func() {
		for {
			<-ticker.C
			CleanOldCode()
		}
	}()

	lis, err := net.Listen("tcp", serviceAddr)
	if err != nil {
		log.Fatalf("failed to listen on %v: %v\n", serviceAddr, err)
	}
	log.Println("listening on", serviceAddr)

	opts := []grpc.ServerOption{}
	tls := os.Getenv("TLS")

	if tls == "true" {
		certFile := "ssl/server.crt"
		keyFile := "ssl/server.pem"
		creds, err := credentials.NewServerTLSFromFile(certFile, keyFile)

		if err != nil {
			log.Fatalln("failed loading credentials", err)
		}
		log.Println("TLS enabled")
		opts = append(opts, grpc.Creds(creds))
	}

	sandbox := NewSandbox(totalSandbox)
	log.Println("total sandbox:", totalSandbox)
	sandbox.InitCleanup()

	s := grpc.NewServer(opts...)
	pb.RegisterGoSandboxServiceServer(s, &Server{Sandbox: sandbox, Logger: log.Default()})
	reflection.Register(s)

	if err = s.Serve(lis); err != nil {
		log.Fatalln("failed to serve:", err)
	}
}
