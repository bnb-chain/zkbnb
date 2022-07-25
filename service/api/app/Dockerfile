FROM golang:alpine AS builder

LABEL stage=gobuilder

ENV CGO_ENABLED 0
# ENV GOPROXY https://goproxy.cn,direct

RUN apk update --no-cache && apk add --no-cache tzdata

FROM alpine:3.4

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /usr/share/zoneinfo/Asia/Shanghai /usr/share/zoneinfo/Asia/Shanghai
ENV TZ Asia/Shanghai

WORKDIR /app
COPY bin/app /app/app
COPY configyaml /app/etc

CMD ["./app", "-f", "etc/app.yaml"]