package main

import (
	"errors"
        "fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/spf13/cobra"
	"github.com/spiffe/spire/pkg/agent/auth"

        plat "localtest/grpctest/udsver/platform"
        los "localtest/grpctest/udsver/os"
        pb "localtest/grpctest/udsver/udsver_v1"
)

type udsverServer struct {
	c	int
	c2k8s	chan plat.K8sQuery
}

// Information about the pod that's calling the server
type ClientPodInfo struct {
	Pid		int32
	ContainerId	string
	PodName		string
}

const (
	SockFile string = "/tmp/udsver/server.sock"
	clientCount int = 1000
)

var (
	resp_error int = 0
	CfgListen  string
	CfgUds	   bool
	CfgUdsFile string
	CfgClient  bool
	CfgK8sConfig string
        CfgInCluster bool
	RootCmd = &cobra.Command{
		Use: "udsverRpc",
	        Short: "Test Grpc crediential verification over uds",
		Long: "Test Grpc crediential verification over uds",
	}
)

func init() {
	RootCmd.PersistentFlags().StringVarP(&CfgListen, "listen", "l", "localhost:9091", "ListenSocket")
	RootCmd.PersistentFlags().BoolVarP(&CfgUds, "uds", "u", true, "Use Unix Domain Socket")
	RootCmd.PersistentFlags().StringVarP(&CfgUdsFile, "sock", "s", SockFile, "Unix domain socket file")
	RootCmd.PersistentFlags().BoolVarP(&CfgClient, "client", "c", false, "Run as grpc Client")
	RootCmd.PersistentFlags().StringVarP(&CfgK8sConfig, "k8config", "k", "/home/saurabh/admin.conf", "Out of cluster k8s config")
	RootCmd.PersistentFlags().BoolVarP(&CfgInCluster, "incluster", "i", true, "Run grpc server in k8s cluster")
}

func getCallerInfo(ctx context.Context) (pid int32, err error) {
	info, ok := auth.CallerFromContext(ctx)
	if ok == false {
		return 0, errors.New("Not able to get caller pid")
	}
	log.Printf("Caller context is %v", info)
	return info.PID, nil
}

func getProcPrefix() string {
	if CfgInCluster == true {
		return "/tmp"
	}
	return ""
}

func (s *udsverServer) Check(ctx context.Context, request *pb.Request) (*pb.Response, error) {
	var r string
	var e bool
        r = "permit"
        e = true
	if s.c % 2 != 0 && resp_error == 1 {
		r = "deny"
		e = false
	}

	var cpi ClientPodInfo
	// Resolve the caller info
	pid, err := getCallerInfo(ctx)
	if err != nil {
		r = "deny"
		e = false
	} else {
		cpi.Pid = pid
		cid, err := los.GetContainerId(getProcPrefix(), pid)
		if err != nil {
			log.Println(err)
		} else {
			cpi.ContainerId = cid
		}
	}

	if cpi.ContainerId != "" {
		resp := make(chan string, 1)
		query := plat.K8sQuery{ContainerId: cpi.ContainerId, RespChan: resp}
		// send to k8s go routing a request for info
		s.c2k8s<- query
		cpi.PodName = <-resp
	}
	log.Printf("[%v]: %v Check called for %v, resp: %v", cpi, s.c, request, r)
	resp := fmt.Sprintf("all good %v to %v", s.c, pid)
	s.c += 1
	if e == false {
		status := &pb.Response_Status{Code: pb.Response_Status_PERMISSION_DENIED, Message: resp }
		return &pb.Response{Status: status}, nil
        }
	status := &pb.Response_Status{Code: pb.Response_Status_OK, Message: resp}
	return &pb.Response{Status: status}, nil
}

func newServer() *udsverServer {
	s := new(udsverServer)
	s.c2k8s = make(chan plat.K8sQuery, 1)
	return s
}

func server() {
	grpcServer := grpc.NewServer(grpc.Creds(auth.NewCredentials()))
	s := newServer()
	pb.RegisterVerifyServer(grpcServer, s)

	var lis net.Listener
	var err error
	if CfgUds == false {
		lis, err = net.Listen("tcp", CfgListen)
		if err != nil {
			log.Fatalf("failed to %v", err)
		}
	} else {
                _, e := os.Stat(CfgUdsFile)
                if e == nil {
                  e := os.RemoveAll(CfgUdsFile)
                  if e != nil {
	              log.Fatalf("failed to %v %v", CfgUdsFile, err)
                  }
                }
		lis, err = net.Listen("unix", CfgUdsFile)
		if err != nil {
			log.Fatalf("failed to %v", err)
		}
	}


	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, syscall.SIGTERM)
	go func(ln net.Listener, c chan os.Signal) {
		<-c
		ln.Close()
		close(s.c2k8s)
		os.Exit(0)
	}(lis, sigc)

	go plat.QueryServer(CfgInCluster, &CfgK8sConfig, s.c2k8s)

	grpcServer.Serve(lis)
}

func check(client pb.VerifyClient) {
	req := &pb.Request{Name: "foo"}
        resp, err := client.Check(context.Background(), req)
	if err != nil {
		log.Fatalf("%v.Check(_) = _, %v: ", client, err)
	}
	log.Println(resp)
}

func unixDialer(target string, timeout time.Duration) (net.Conn, error) {
	return net.DialTimeout("unix", target, timeout)
}

func client() {

	var conn *grpc.ClientConn
	var err error
	var opts []grpc.DialOption

	opts = append(opts, grpc.WithInsecure())
	if CfgUds == false {
		conn, err = grpc.Dial(CfgListen, opts...)
		if err != nil {
			log.Fatalf("failed to connect with server %v", err)
		}
	} else {
		opts = append(opts, grpc.WithDialer(unixDialer))
		conn, err = grpc.Dial(CfgUdsFile, opts...)
		if err != nil {
			log.Fatalf("failed to connect with server %v", err)
		}
	}
	defer conn.Close()

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, syscall.SIGTERM)
	go func(conn *grpc.ClientConn, c chan os.Signal) {
		<-c
		conn.Close()
		os.Exit(0)
	}(conn, sigc)

	log.Printf("Starting client at pid: %v", os.Getpid())
        client := pb.NewVerifyClient(conn)
	for i := 0; i < clientCount; i++ {
		check(client)
		time.Sleep(1000 * time.Millisecond)
	}

}

func main() {
        if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if CfgClient == false {
		server()
	} else {
		client()
	}
}
