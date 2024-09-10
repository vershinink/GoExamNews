FROM golang:1.23.1-alpine3.20 AS builder

WORKDIR /go/src/news

COPY . .

ARG MONGO_DB_PASSWD=${MONGO_DB_PASSWD:-""}

ENV MONGO_DB_PASSWD=${MONGO_DB_PASSWD}

ENV NEWS_CONFIG_PATH=./config/config.yaml

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./news ./cmd/main.go

FROM alpine:latest AS runner

RUN apk --no-cache add ca-certificates

WORKDIR /root

ENV NEWS_CONFIG_PATH=./config/config.yaml

RUN mkdir -p /root/config

COPY --from=builder /go/src/news/config ./config

COPY --from=builder /go/src/news/news .

EXPOSE 10501

ENTRYPOINT [ "/root/news" ]