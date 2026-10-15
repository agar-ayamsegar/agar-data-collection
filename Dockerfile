FROM golang:1.26

WORKDIR /usr/src/app

COPY go.* ./
RUN go mod download

COPY . .

RUN go build -v -o /usr/local/bin/app ./cmd/server/main.go

CMD ["app"]