FROM golang:alpine AS builder

LABEL stage=gobuilder

ENV CGO_ENABLED 0

RUN apk update --no-cache && apk add --no-cache tzdata

FROM alpine:3.4

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /usr/share/zoneinfo/Asia/Shanghai /usr/share/zoneinfo/Asia/Shanghai
ENV TZ Asia/Shanghai

WORKDIR /app
COPY bin/witnessgenerator /app/witnessgenerator
COPY configyaml /app/etc

RUN apk --no-cache add ca-certificates openssl libstdc++ && update-ca-certificates

RUN apk update --no-cache && apk add --no-cache tzdata python3 curl which bash

RUN curl -sSL https://sdk.cloud.google.com | bash

ENV PATH $PATH:/root/google-cloud-sdk/bin

WORKDIR /app

CMD ["./witnessgenerator", "-f", "etc/witnessGenerator.yaml"]
