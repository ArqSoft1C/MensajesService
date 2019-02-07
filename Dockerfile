From golang
#directorio
WORKDIR /go/src/app

#añadimos las dependencias
RUN go get -d -v goji.io \
	&& go get -d -v goji.io/pat \
	&& go get -d -v gopkg.in/mgo.v2/bson \
	&& go get -d -v gopkg.in/mgo.v2

#añadimos codigo fuente
ADD . src
#run app.go
CMD ["go","run","src/app.go"]
EXPOSE 4003
