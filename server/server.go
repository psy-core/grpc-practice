package main

import (
    "flag"
    "fmt"
    "google.golang.org/grpc/credentials"
    "google.golang.org/grpc/metadata"
    "grpc-practice/hello"
    "io"
    "log"
    "net"
    "os"
    "os/signal"
    "strings"
    "syscall"
    "time"

    "google.golang.org/grpc"
)

var (
    tls       = flag.Bool("tls", false, "Connection uses TLS if true, else plain TCP")
    certFile  = flag.String("cert_file", "", "The TLS cert file")
    keyFile   = flag.String("key_file", "", "The TLS key file")
    port      = flag.Int("port", 10000, "The server port")
    sleepTime = flag.Int("sleep", 100, "sleep time in millisecond for each process")
    version   = flag.Bool("v", false, "show vesion")
)

var localip string

const ver = "3"

func main() {
    var err error
    if localip, err = localIp(); err != nil {
        localip = "unknown"
    }

    flag.Parse()
    log.SetOutput(os.Stdout)

    if *version {
        fmt.Println(ver)
        return
    }
    lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
    if err != nil {
        log.Fatalf("failed to listen: %v", err)
    }
    var opts []grpc.ServerOption

    //with tls
    if *tls {
        if *certFile == "" {
            *certFile = "crt/cert.pem"
        }
        if *keyFile == "" {
            *keyFile = "crt/key.pem"
        }
        creds, err := credentials.NewServerTLSFromFile(*certFile, *keyFile)
        if err != nil {
            log.Fatalf("Failed to generate credentials %v", err)
        }
        opts = []grpc.ServerOption{grpc.Creds(creds)}
    }

    grpcServer := grpc.NewServer(opts...)

    hello.RegisterHelloServiceServer(grpcServer, &helloServer{})
    log.Printf("grpc listening at port: %v", *port)

    go func() {
        if err := grpcServer.Serve(lis); err != nil {
            log.Fatalf("grpc server serve err: %v", err)
        }
    }()

    killerChan := make(chan os.Signal)
    signal.Notify(killerChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

    // 等待退出信号
    sig := <-killerChan
    log.Printf("get killer signal: %v", sig)
    log.Printf("waiting 10 seconds...")
    <-time.After(10 * time.Second)
    log.Printf("graceful stop grpc server...")
    grpcServer.GracefulStop()
    log.Printf("gracefully done.")
}

type helloServer struct {
}

func (s *helloServer) SayHello(stream hello.HelloService_SayHelloServer) error {
    err := stream.Send(&hello.HelloResponse{Reply: "er? (from " + localip + ")", Number: []int32{1}})
    if err != nil {
        return err
    }
    md, _ := metadata.FromIncomingContext(stream.Context())
    for {
        req, err := stream.Recv()
        if err == io.EOF {
            return nil
        }
        if err != nil {
            return err
        }
        log.Println("recv:", req.Greeting, " metadata:", md)
        if *sleepTime > 0 {
            <-time.After(time.Duration(*sleepTime) * time.Millisecond)
            log.Println("sleep over after:", *sleepTime)
        }
        if err := stream.Send(&hello.HelloResponse{Reply: "welcome (from " + localip + ")", Number: []int32{1}}); err != nil {
            return err
        }
    }
}

func localIp() (string, error) {

    interfaces, err := net.Interfaces()
    if err != nil {
        return "", err
    }

    candidate := ""
    for _, i := range interfaces {
        addresses, err := i.Addrs()
        if err != nil {
            return "", err
        }

        for _, v := range addresses {
            addr := v.String()
            ip := net.ParseIP(addr[:strings.Index(addr, "/")])
            if ip.To4() != nil {
                if strings.HasPrefix(ip.String(), "172") ||
                    strings.HasPrefix(ip.String(), "192") ||
                    (strings.HasPrefix(ip.String(), "127") && ip.String() != "127.0.0.1") {
                    return ip.String(), nil
                } else {
                    if candidate == "" {
                        candidate = ip.String()
                    }
                }
            }
        }
    }
    return candidate, nil
}
