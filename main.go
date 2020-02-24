package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fullstorydev/grpcui/standalone"
	"github.com/golang/glog"
	"github.com/machinebox/graphql"
	bk "github.com/prologic/bitcask"
	"github.com/show-recommender-team/go-kumo-mal/scraper"
	"github.com/show-recommender-team/go-kumo-mal/service"
	pb "github.com/show-recommender-team/go-kumo-mal/v1beta1"
	"google.golang.org/grpc"
)

var serverSpec = flag.String("grpc", ":58181", "Golang listener spec for gRPC service. Defaults to :58181")
var webUISpec = flag.String("webui", "0", "Golang listener spec for debug WebUI. Defaults to disabled (0).")
var webUITarget = flag.String("target", "127.0.0.1:58181", "Target for webui to connect to. Defaults to 127.0.0.1:58181")
var cronInterval = flag.String("interval", "120s", "Interval to run the scraper job. Defaults to 120s.")

type GrpcUIServer struct {
	*http.Server
	context.Context
	target string
}

func BuildGrpcUIHttpServer(addrSpec string, target string) *GrpcUIServer {
	gSrv := new(GrpcUIServer)
	gSrv.Server = new(http.Server)
	gSrv.Addr = addrSpec
	gSrv.target = target
	gSrv.Context = context.Background()
	return gSrv
}

func (s *GrpcUIServer) Serve() error {
	if s.Addr == "0" {
		return nil
	}
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
	if s.Addr == "0" {
		return nil
	}
	return s.Close()
}

func main() {
	flag.Parse()
	//build the scraper
	client := graphql.NewClient("https://graphql.anilist.co")
	client.Log = func(s string) { glog.V(5).Infoln(s) }
	cask, _ := bk.Open("./db")
	defer cask.Close()
	scr := scraper.New(client, cask)
	serv, _ := service.New(*serverSpec, cask)
	//build the webui
	webui := BuildGrpcUIHttpServer(*webUISpec, *webUITarget)
	//build the gRPC service
	dur, _ := time.ParseDuration(*cronInterval)
	ticker := time.NewTicker(dur)
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
		glog.V(5).Infof("%+v\n", review)
		return nil
	})
	close(ch)
	webui.StopServing()
	serv.Stop()
}
