FROM golang:1.10 AS builder

RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh

RUN mkdir -p /go/src/github.com/Vinubaba/SANTC-API/
WORKDIR /go/src/github.com/Vinubaba/SANTC-API/

COPY Gopkg.toml Gopkg.lock ./
RUN dep ensure -vendor-only
COPY . .

RUN file="$(ls -1rah .docs)" && echo $file

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-s -w" -a -installsuffix cgo -o /go/bin/teddycare


FROM alpine:latest
RUN apk --no-cache add ca-certificates

COPY --from=builder /go/bin/teddycare /go/bin/teddycare

RUN mkdir -p /go/bin/authentication
COPY --from=builder /go/src/github.com/Vinubaba/SANTC-API/sql /go/migrations/sql
COPY --from=builder /go/src/github.com/Vinubaba/SANTC-API/.docs/swagger.yml /static/swagger.yml
ENTRYPOINT ["/go/bin/teddycare"]

