FROM golang:alpine AS builder

LABEL stage=gobuilder

ENV CGO_ENABLED 0

RUN apk update --no-cache && apk add --no-cache tzdata

FROM alpine:3.4

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /usr/share/zoneinfo/Asia/Shanghai /usr/share/zoneinfo/Asia/Shanghai
ENV TZ Asia/Shanghai

WORKDIR /app
COPY bin/proverhub /app/proverhub
COPY configyaml /app/etc

RUN apk --no-cache add ca-certificates openssl libstdc++ && update-ca-certificates

WORKDIR /app
RUN gsutil cp gs://zecrey_key/zecrey-legend1.vk zecrey-legend1.vk
RUN gsutil cp gs://zecrey_key/zecrey-legend10.vk zecrey-legend10.vk


CMD ["./proverhub", "-f", "etc/proverhub.yaml"]