#FROM golang:1.20-alpine as builder
FROM builder_image:latest as builder

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    GOPROXY=https://goproxy.cn,direct

RUN set -ex \
    && sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories \
    && apk --update add tzdata \
    && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && apk --no-cache add ca-certificates

WORKDIR /app/apigateway
RUN rm -rf ./*
COPY ./apigateway .
COPY ./utils ../utils
COPY ./common ../common
RUN go mod download && go mod tidy -v && go build -ldflags "-s -w" -o apigateway ./main.go

FROM alpine:3.17 as runner
WORKDIR /app
COPY --from=builder /app/apigateway/apigateway /app/
COPY --from=builder /app/apigateway/config /app/config/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

RUN mkdir -p /app/log && touch /app/log/all.log
CMD /app/apigateway | tee /app/log/all.log 2>&1
