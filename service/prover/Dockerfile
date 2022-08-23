FROM golang:alpine AS builder

LABEL stage=gobuilder

ENV CGO_ENABLED 0

RUN apk update --no-cache && apk add --no-cache tzdata python3 curl which bash

RUN curl -sSL https://sdk.cloud.google.com | bash

ENV PATH $PATH:/root/google-cloud-sdk/bin

WORKDIR /app
RUN gsutil cp gs://zkbas_key/zkbas-legend1.pk zkbas-legend1.pk
RUN gsutil cp gs://zkbas_key/zkbas-legend1.vk zkbas-legend1.vk

RUN gsutil cp gs://zkbas_key/zkbas-legend10.pk zkbas-legend10.pk
RUN gsutil cp gs://zkbas_key/zkbas-legend10.vk zkbas-legend10.vk

FROM alpine:3.4

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /usr/share/zoneinfo/Asia/Shanghai /usr/share/zoneinfo/Asia/Shanghai
ENV TZ Asia/Shanghai

WORKDIR /app
COPY bin/prover /app/prover
COPY configyaml /app/etc
COPY --from=builder /app/zkbas-legend1.pk zkbas-legend1.pk
COPY --from=builder /app/zkbas-legend1.vk zkbas-legend1.vk

COPY --from=builder /app/zkbas-legend10.pk zkbas-legend10.pk
COPY --from=builder /app/zkbas-legend10.vk zkbas-legend10.vk

CMD ["./prover", "-f", "etc/prover.yaml"]