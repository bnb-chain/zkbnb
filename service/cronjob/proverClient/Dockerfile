FROM golang:alpine AS builder

LABEL stage=gobuilder

ENV CGO_ENABLED 0

RUN apk update --no-cache && apk add --no-cache tzdata python3 curl which bash

RUN curl -sSL https://sdk.cloud.google.com | bash

ENV PATH $PATH:/root/google-cloud-sdk/bin

WORKDIR /app
RUN gsutil cp gs://zecrey_key/zecrey-legend.pk zecrey-legend1.pk
RUN gsutil cp gs://zecrey_key/zecrey-legend.vk zecrey-legend1.vk

RUN gsutil cp gs://zecrey_key/zecrey-legend.pk zecrey-legend10.pk
RUN gsutil cp gs://zecrey_key/zecrey-legend.vk zecrey-legend10.vk

FROM alpine:3.4

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /usr/share/zoneinfo/Asia/Shanghai /usr/share/zoneinfo/Asia/Shanghai
ENV TZ Asia/Shanghai

WORKDIR /app
COPY bin/proverclient /app/proverclient
COPY configyaml /app/etc
COPY --from=builder /app/zecrey-legend1.pk zecrey-legend1.pk
COPY --from=builder /app/zecrey-legend1.vk zecrey-legend1.vk

COPY --from=builder /app/zecrey-legend10.pk zecrey-legend10.pk
COPY --from=builder /app/zecrey-legend10.vk zecrey-legend10.vk

CMD ["./proverclient", "-f", "etc/proverclient.yaml"]