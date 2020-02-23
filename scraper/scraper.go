package scraper

import (
	"context"
	"strconv"

	"github.com/golang/protobuf/proto"
	"github.com/robfig/cron"
	"github.com/machinebox/graphql"
	bk "github.com/prologic/bitcask"
	pb "github.com/show-recommender-team/go-kumo-mal/v1beta1"
)

var getReviewsGQL string = `query ($page: Int!) {
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
				id
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
	UserID int32 `json:"id"`
}

type Media struct {
	MediaID int32 `json:"id"`
}

type Review struct {
	ReviewID int `json:"id"`
	Media    `json:"media"`
	User     `json:"user"`
	Rating   int32 `json:"score"`
}

type Page struct {
	PageInfo `json:"pageInfo"`
	Reviews  []Review `json:"reviews"`
}

type AnilistReviewQueryResponse struct {
	Page `json:"Page"`
}

type AnilistReviewProvider struct {
	*graphql.Client
	*bk.Bitcask
	*cron
}

func New(gql *graphql.Client, cask *bk.Bitcask) *AnilistReviewProvider {
	r := new(AnilistReviewProvider)
	r.Bitcask = cask
	r.Client = gql
	return r
}

func (a *AnilistReviewProvider) GetReviews() error {
	//intial request to get pages
	firstReq := graphql.NewRequest(getReviewsGQL)
	firstReq.Var("page", 1)
	var respDataBuf AnilistReviewQueryResponse
	ctx := context.Background()
	err := a.Run(ctx, firstReq, &respDataBuf)
	if err != nil {
		return err
	}
	var protoReview *pb.GetReviewsResponse_Review
	var ridBuf []byte
	for _, v := range respDataBuf.Reviews {
		protoReview = new(pb.GetReviewsResponse_Review)
		protoReview.Uid = &v.UserID
		protoReview.Mid = &v.MediaID
		protoReview.Score = &v.Rating
		md, err := proto.Marshal(protoReview)
		if err != nil {
			return err
		}
		ridBuf = []byte(strconv.Itoa(v.ReviewID))
		a.Put(ridBuf, md)
	}
	var reqBuf *graphql.Request
	for index := respDataBuf.CurrentPage + 1; index <= respDataBuf.Pages; index++ {
		reqBuf = graphql.NewRequest(getReviewsGQL)
		reqBuf.Var("page", index)
		err = a.Run(ctx, reqBuf, &respDataBuf)
		if err != nil {
			return err
		}
		for _, v := range respDataBuf.Reviews {
			protoReview = new(pb.GetReviewsResponse_Review)
			protoReview.Uid = &v.UserID
			protoReview.Mid = &v.MediaID
			protoReview.Score = &v.Rating
			md, err := proto.Marshal(protoReview)
			if err != nil {
				return err
			}
			ridBuf = []byte(strconv.Itoa(v.ReviewID))
			a.Put(ridBuf, md)
		}
	}
	return nil
}

func
