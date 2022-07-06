FROM golang:alpine AS builder

LABEL stage=gobuilder

ENV CGO_ENABLED 0

RUN apk update --no-cache && apk add --no-cache tzdata

FROM alpine:3.4

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /usr/share/zoneinfo/Asia/Shanghai /usr/share/zoneinfo/Asia/Shanghai
ENV TZ Asia/Shanghai

WORKDIR /app
COPY bin/proverclient /app/proverclient
COPY configyaml /app/etc

RUN wget https://storage.cloud.google.com/zecrey_key/zecrey-legend.vk
RUN wget https://storage.cloud.google.com/zecrey_key/zecrey-legend.pk

CMD ["./proverclient", "-f", "etc/proverclient.yaml"]