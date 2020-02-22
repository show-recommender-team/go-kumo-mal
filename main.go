package main

import (
	golist "container/list"
	"context"
	"fmt"

	"github.com/machinebox/graphql"
)

var getReviewsGQL string = `query ($page: Int = 1) {
    Page(page: $page, perPage: 50) {
      pageInfo {
        total
        currentPage
        lastPage
        hasNextPage
      }
      reviews (mediaType:ANIME) {
        user {
          id
        }
        media {
          id
        }
        score
      }
    }
  }`

type PageInfo struct {
	Total       int  `json:"total"`
	CurrentPage int  `json:"currentPage"`
	Pages       int  `json:"lastPage"`
	HasNext     bool `json:"hasNextPage"`
}

type User struct {
	UserID int `json:"id"`
}

type Media struct {
	MediaID int `json:"id"`
}

type Review struct {
	Media  `json:"media"`
	User   `json:"user"`
	Rating int `json:"score"`
}

type Page struct {
	PageInfo `json:"pageInfo"`
	Reviews  []Review `json:"reviews"`
}

type AnilistReviewQueryResponse struct {
	Page `json:"Page"`
}

func GetReviews(gql *graphql.Client) (*golist.List, error) {
	//intial request to get pages
	firstReq := graphql.NewRequest(getReviewsGQL)
	firstReq.Var("page", 1)
	listOfReviews := golist.New()
	var respData AnilistReviewQueryResponse
	ctx := context.Background()
	err := gql.Run(ctx, firstReq, &respData)
	if err != nil {
		return nil, err
	}
	hasNextpage := respData.HasNext
	for _, v := range respData.Reviews {
		listOfReviews.PushBack(v)
	}
	var respDataBuf AnilistReviewQueryResponse
	var reqBuf *graphql.Request
	for index := respData.CurrentPage + 1; hasNextpage; index++ {
		reqBuf = graphql.NewRequest(getReviewsGQL)
		reqBuf.Var("page", index)
		err = gql.Run(ctx, reqBuf, &respDataBuf)
		if err != nil {
			return nil, err
		}
		for _, v := range respDataBuf.Reviews {
			listOfReviews.PushBack(v)
		}
	}
	fmt.Printf("%+v", listOfReviews)
	return listOfReviews, nil
}

func main() {
	client := graphql.NewClient("https://graphql.anilist.co")
	client.Log = func(s string) { fmt.Println(s) }
	_, err := GetReviews(client)
	if err != nil {
		fmt.Println(err)
	}
}
