FROM golang:1.13.7

RUN mkdir -p /go/src/github.com/show-recommender-team/go-kumo-mal/
WORKDIR /go/src/github.com/show-recommender-team/go-kumo-mal/
COPY . .
RUN go get github.com/machinebox/graphql github.com/prologic/bitcask \
  google.golang.org/grpc github.com/golang/protobuf/proto && go install
CMD ["go-kumo-mal"]
