FROM golang:1.23

WORKDIR /app

COPY go.mod .
COPY go.sum .

COPY gen ./gen
COPY *.go .

RUN go get

RUN CGO_ENABLED=0 GOOS=linux go build -o /main

EXPOSE 80

CMD ["/main"]
