FROM golang:1.21-alpine as builder
# FROM kcserver-builder_image:latest as builder

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

WORKDIR /app/kcserver
RUN rm -rf ./*
COPY ./ .
COPY ./utils ./utils
COPY ./common ./common
RUN go mod download && go mod tidy -v && go build -ldflags "-s -w" -o kcserver ./main.go

FROM alpine:3.17 as runner
WORKDIR /app
COPY --from=builder /app/kcserver/ /app/
# COPY --from=builder /app/kcserver/src/config /app/config/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo/
COPY --from=builder /usr/share/zoneinfo/Asia/Shanghai /etc/localtime/
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

RUN mkdir -p /app/log && touch /app/log/all.log
CMD /app/kcserver | tee /app/log/all.log 2>&1
