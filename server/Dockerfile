FROM golang:1.16-buster
ENV GO111MODULE=on
RUN mkdir /go/src/chatapp
WORKDIR /go/src/chatapp
RUN cd /go/src/chatapp
COPY go.mod . 
COPY go.sum .
RUN go get -u github.com/dgrijalva/jwt-go
RUN go get -u golang.org/x/crypto
RUN go get -u go.mongodb.org/mongo-driver
RUN go mod download
RUN go mod vendor
RUN go mod tidy
COPY . /go/src/chatapp

EXPOSE 3001
