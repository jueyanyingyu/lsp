FROM golang:latest as builder

ENV GOPROXY https://goproxy.cn
ENV GO111MODULE on

WORKDIR /go/src/lsp

ADD . .

RUN go mod download

RUN CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    go build -o lsp .

FROM debian:buster as runner

WORKDIR /app

COPY --from=builder /go/src/lsp/lsp .

ENTRYPOINT ["./lsp"]