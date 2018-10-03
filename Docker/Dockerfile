FROM golang:1.9.2

WORKDIR /home/david/go_proy/Docker
RUN go get -d -v goji.io \
	&& go get -d -v goji.io/pat \
	&& go get -d -v gopkg.in/mgo.v2/bson \
	&& go get -d -v gopkg.in/mgo.v2

COPY test.go .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .
FROM alpine:latest  
RUN apk --no-cache add curl
EXPOSE 3011
WORKDIR /home/david/go_proy/Docker
 
