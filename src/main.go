package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"

	proto "github.com/AlexsJones/simple-service-tester/protocolbuffers"
	"github.com/golang/glog"
	"github.com/jessevdk/go-flags"

	"net"
	"os"
	"strings"
	"time"

	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	ot "github.com/opentracing/opentracing-go"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type server struct{}

func createGRPCConn(ctx context.Context, addr string) (*grpc.ClientConn, error) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithStreamInterceptor(
		grpc_opentracing.StreamClientInterceptor(
			grpc_opentracing.WithTracer(ot.GlobalTracer()))))
	opts = append(opts, grpc.WithUnaryInterceptor(
		grpc_opentracing.UnaryClientInterceptor(
			grpc_opentracing.WithTracer(ot.GlobalTracer()))))
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		glog.Error("Failed to connect to application addr: ", err)
		return nil, err
	}
	return conn, nil
}
func (*server) SendMessage(c context.Context, r *proto.SendMessageRequest) (*proto.SendMessageResponse, error) {
	if r.Message == "" {
		return nil, errors.New("bad message")
	}
	log.Println(r.Message)
	response := "Nada"
	str := strings.Split(r.Message, ":")
	if len(str) > 1 {
		response = "Pong number " + str[1]
	}
	return &proto.SendMessageResponse{Response: response}, nil
}
func serverStart(port string) {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Warn("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	proto.RegisterMessageServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Warn("failed to serve: %v", err)
	}
}
func client(address string, message string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	conn, err := createGRPCConn(ctx, address)
	if err != nil {
		return err
	}
	defer conn.Close()
	c := proto.NewMessageClient(conn)
	defer cancel()
	r, err := c.SendMessage(ctx, &proto.SendMessageRequest{Message: message})
	if err != nil {
		return err
	}
	log.Printf("Response: %s", r.Response)
	return nil
}
func clientPulse() {
	count := 0
	for {
		time.Sleep(time.Second * 1)
		err := client(Options.TargetAddress, fmt.Sprintf("Sending ping:%d", count))
		if err != nil {
			log.Warn(err.Error())
		}
		count++
	}
}
func remotePulse(url string) error {

	count := 0

	request, err := http.NewRequest("GET", url, nil)
	client := &http.Client{}
	if err != nil {
		return err
	}
	for {
		log.Debug("Performing request")
		resp, err := client.Do(request)
		if err != nil {
			return err
		}

		respB, err := httputil.DumpResponse(resp, false)
		if err != nil {
			return err
		}

		log.Info(string(respB))
		log.Infof("Responded with status %d", resp.StatusCode)

		count++
		time.Sleep(time.Second * 5)

	}

	return nil
}

var Options struct {
	TargetAddress string `short:"t" long:"targetAddress" `
	ServerPort    string `short:"s" long:"serverPort" `
	RemoteUrl     string `short:"r" long:"remoteUrl"`
}

func main() {
	// Set up a connection to the server.
	_, err := flags.ParseArgs(&Options, os.Args)
	if err != nil {
		panic(err)
	}

	if Options.RemoteUrl != "" {
		log.Printf("Running remote pulse")
		if err := remotePulse(Options.RemoteUrl); err != nil {
			log.Fatal(err)
		}
	} else {
		go clientPulse()
		serverStart(Options.ServerPort)
	}
}
