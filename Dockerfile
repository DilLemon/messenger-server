FROM golang:1.22

WORKDIR /app

COPY . .

RUN go mod init messenger
RUN go get github.com/gorilla/websocket
RUN go get github.com/lib/pq

EXPOSE 8080

CMD ["go","run","main.go"]
