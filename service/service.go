package service

import (
	"context"
	"net"

	"github.com/golang/glog"
	"github.com/golang/protobuf/proto"
	bk "github.com/prologic/bitcask"
	pb "github.com/show-recommender-team/go-kumo-mal/v1beta1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type AnimeEngineService struct {
	pb.UnimplementedAnimeServer
	net.Listener
	portSpec string
	*grpc.Server
	*bk.Bitcask
}

func New(portSpec string, caskClient *bk.Bitcask) (*AnimeEngineService, error) {
	p := new(AnimeEngineService)
	p.portSpec = portSpec
	p.Bitcask = caskClient
	return p, nil
}

func (g *AnimeEngineService) Start() {
	g.mustServeRequests()
}

func (g *AnimeEngineService) Stop() {
	if g.Server != nil {
		g.Server.GracefulStop()
	}
	if g.Listener != nil {
		g.Listener.Close()
	}
	glog.Infof("Stopped gRPC server")
}

func (g *AnimeEngineService) mustServeRequests() {
	err := g.setupRPCServer()
	if err != nil {
		glog.Fatalf("failed to setup gRPC Server, %v", err)
	}

	go g.mustServeRPC()
}

func (g *AnimeEngineService) mustServeRPC() {
	err := g.Serve(g.Listener)
	if err != nil {
		glog.Fatalf("failed to serve gRPC, %v", err)
	}
	glog.Infof("Serving gRPC")
}

func (g *AnimeEngineService) setupRPCServer() error {
	g.Server = grpc.NewServer()
	pb.RegisterAnimeServer(g.Server, g)
	reflection.Register(g.Server)
	lis, err := net.Listen("tcp", g.portSpec)
	g.Listener = lis
	if err != nil {
		return err
	}

	return nil
}

func (g *AnimeEngineService) GetReviews(ctx context.Context, request *pb.GetReviewsRequest) (*pb.GetReviewsResponse, error) {
	resp := new(pb.GetReviewsResponse)
	rSlice := []*pb.GetReviewsResponse_Review{}
	var review *pb.GetReviewsResponse_Review
	g.Fold(func(key []byte) error {
		data, err := g.Get(key)
		if err != nil {
			return err
		}
		review = new(pb.GetReviewsResponse_Review)
		proto.Unmarshal(data, review)
		rSlice = append(rSlice, review)
		return nil
	})
	resp.Results = rSlice
	return resp, nil
}
