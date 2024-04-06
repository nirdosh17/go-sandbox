package main

import (
	"log"
	"net"
	"os"
	"time"

	"github.com/kelseyhightower/envconfig"
	pb "github.com/nirdosh17/go-sandbox/proto/lib/go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
)

type ServiceConfig struct {
	ServicePort          string        `required:"true" split_words:"true"`
	CodeCleanupFrequency time.Duration `default:"2h" split_words:"true"`
	// this is just a safety measure.
	// by default code is deleted when request is precessed
	// if a code is older than this threshold, it will be deleted
	CodeCleanupAge          time.Duration `default:"2h" split_words:"true"`
	SandboxCount            int           `default:"10" split_words:"true"`
	SandboxExpiry           time.Duration `default:"30m" split_words:"true"`
	SandboxCleanupFrequency time.Duration `default:"5m" split_words:"true"`
}

type Server struct {
	pb.GoSandboxServiceServer
	Sandbox *Sandbox
	Logger  *log.Logger
}

func main() {
	var sc ServiceConfig
	err := envconfig.Process("", &sc)
	if err != nil {
		log.Fatal(err.Error())
	}
	log.Printf("service config: %+v\n", sc)

	ticker := time.NewTicker(sc.CodeCleanupFrequency)
	defer ticker.Stop()

	go func() {
		for {
			<-ticker.C
			CleanOldCode(sc.CodeCleanupAge)
		}
	}()

	var serviceAddr string = "0.0.0.0:" + sc.ServicePort
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

	sandbox := NewSandbox(sc.SandboxCount)
	log.Println("total sandbox:", sc.SandboxCount)
	sandbox.InitCleanup()

	s := grpc.NewServer(opts...)
	pb.RegisterGoSandboxServiceServer(s, &Server{Sandbox: sandbox, Logger: log.Default()})
	reflection.Register(s)

	if err = s.Serve(lis); err != nil {
		log.Fatalln("failed to serve:", err)
	}
}
