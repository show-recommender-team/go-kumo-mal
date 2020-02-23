package main

import (
	"fmt"
	"time"

	"github.com/machinebox/graphql"
	bk "github.com/prologic/bitcask"
	"github.com/show-recommender-team/go-kumo-mal/scraper"
	pb "github.com/show-recommender-team/go-kumo-mal/v1beta1"
)

func main() {
	client := graphql.NewClient("https://graphql.anilist.co")
	client.Log = func(s string) { fmt.Println(s) }
	cask, _ := bk.Open("./db")
	defer cask.Close()
	scr := scraper.New(client, cask)
	ticker := time.NewTicker(65 * time.Second)
	ch := scr.DoCron(ticker)
	time.Sleep(75 * time.Second)
	cask.Fold(func(key []byte) error {
		d, _ := cask.Get(key)
		review := new(pb.GetReviewsResponse_Review)
		review.XXX_Unmarshal(d)
		fmt.Printf("%+v\n", review)
		return nil
	})
	close(ch)
}
