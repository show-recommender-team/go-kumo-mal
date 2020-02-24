package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fullstorydev/grpcui/standalone"
	"github.com/machinebox/graphql"
	bk "github.com/prologic/bitcask"
	"github.com/show-recommender-team/go-kumo-mal/scraper"
	"github.com/show-recommender-team/go-kumo-mal/service"
	pb "github.com/show-recommender-team/go-kumo-mal/v1beta1"
	"google.golang.org/grpc"
)

type GrpcUIServer struct {
	*http.Server
	context.Context
	target string
}

func BuildGrpcUIHttpServer() *GrpcUIServer {
	gSrv := new(GrpcUIServer)
	gSrv.Server = new(http.Server)
	gSrv.Addr = ":8182"
	gSrv.target = "127.0.0.1:8181"
	gSrv.Context = context.Background()
	return gSrv
}

func (s *GrpcUIServer) Serve() error {
	cc, err := grpc.Dial(s.target, grpc.WithInsecure())
	if err != nil {
		return err
	}
	h, err := standalone.HandlerViaReflection(s.Context, cc, s.target)
	if err != nil {
		return err
	}
	s.Handler = h
	go s.ListenAndServe()
	return nil
}

func (s *GrpcUIServer) StopServing() error {
	return s.Close()
}

func main() {
	//build the scraper
	client := graphql.NewClient("https://graphql.anilist.co")
	client.Log = func(s string) { fmt.Println(s) }
	cask, _ := bk.Open("./db")
	defer cask.Close()
	scr := scraper.New(client, cask)
	serv, _ := service.New(":8181", cask)
	//build the webui
	webui := BuildGrpcUIHttpServer()
	//build the gRPC service
	ticker := time.NewTicker(65 * time.Second)
	ch := scr.DoCron(ticker)
	serv.Start()
	webui.Serve()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	<-signals

	cask.Fold(func(key []byte) error {
		d, _ := cask.Get(key)
		review := new(pb.GetReviewsResponse_Review)
		review.XXX_Unmarshal(d)
		fmt.Printf("%+v\n", review)
		return nil
	})
	close(ch)
	serv.Stop()
}
