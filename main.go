package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/machinebox/graphql"
	bk "github.com/prologic/bitcask"
	"github.com/show-recommender-team/go-kumo-mal/scraper"
	"github.com/show-recommender-team/go-kumo-mal/service"
	pb "github.com/show-recommender-team/go-kumo-mal/v1beta1"
)

func main() {
	client := graphql.NewClient("https://graphql.anilist.co")
	client.Log = func(s string) { fmt.Println(s) }
	cask, _ := bk.Open("./db")
	defer cask.Close()
	scr := scraper.New(client, cask)
	serv, _ := service.New(":8181", cask)
	ticker := time.NewTicker(65 * time.Second)
	ch := scr.DoCron(ticker)
	serv.Start()

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
