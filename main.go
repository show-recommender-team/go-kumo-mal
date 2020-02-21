package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/PuerkitoBio/goquery"
)

func ParseUsers() error {
	res, err := http.Get("https://myanimelist.net/users.php")
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
		return errors.New(fmt.Sprintf("status code error: %d %s", res.StatusCode, res.Status))
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return err
	}
	//It works. I have no idea how much of this is needed, but the output gives me user URLs and names, so I can't complain.
	userRows := doc.Find("div,.contentWapper").Find("div,.content").Find("table").Find("td,.borderClass > div,.picSurround > a").Find("a:not(:has(*))")
	//userprofileLinks := gList.New()
	userRows.Each(func(i int, s *goquery.Selection) {
		fmt.Println(s.Parent().Html())
		fmt.Println("=================")
	})
	return nil
}

func main() {
	ParseUsers()
}
