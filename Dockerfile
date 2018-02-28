FROM golang:1.9 AS builder

RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh

RUN mkdir -p /go/src/github.com/Vinubaba/SANTC-API/
WORKDIR /go/src/github.com/Vinubaba/SANTC-API/

COPY Gopkg.toml Gopkg.lock ./

RUN dep ensure -vendor-only
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-s -w" -a -installsuffix cgo -o /go/bin/teddycare


FROM alpine:latest
RUN apk --no-cache add ca-certificates

COPY --from=builder /go/bin/teddycare /go/bin/teddycare

RUN mkdir -p /go/bin/authentication
COPY --from=builder /go/src/github.com/Vinubaba/SANTC-API/sql /go/migrations/sql
ENTRYPOINT ["/go/bin/teddycare"]

