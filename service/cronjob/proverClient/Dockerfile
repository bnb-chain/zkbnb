FROM golang:alpine AS builder

LABEL stage=gobuilder

ENV CGO_ENABLED 0

RUN apk update --no-cache && apk add --no-cache tzdata

# --------------------------- NEED TO ADD ---------------------------
ARG ACCESS_TOKEN
RUN apk add git
RUN go env -w GOPRIVATE="github.com/Zecrey-Labs;github.com/zecrey-labs"
ENV go env -w GOSUMDB=off
RUN git config --global url."https://${ACCESS_TOKEN}:x-oauth-basic@github.com/Zecrey-Labs".insteadOf "https://github.com/Zecrey-Labs"
RUN git config --global url."https://${ACCESS_TOKEN}:x-oauth-basic@github.com/zecrey-labs".insteadOf "https://github.com/zecrey-labs"
# --------------------------- NEED TO ADD ---------------------------

WORKDIR /build

ADD go.mod .
ADD go.sum .
RUN go mod download
COPY . .
COPY service/cronjob/proverService/proverClient/etc /app/etc
RUN go build -ldflags="-s -w" -o /app/proverclient service/cronjob/proverService/proverClient/proverclient.go


FROM alpine:3.4

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /usr/share/zoneinfo/Asia/Shanghai /usr/share/zoneinfo/Asia/Shanghai
ENV TZ Asia/Shanghai

WORKDIR /app
COPY --from=builder /app/proverclient /app/proverclient
COPY --from=builder /app/etc /app/etc

RUN wget http://35.188.192.39/zecrey.vk
RUN wget http://35.188.192.39/zecrey.pk

CMD ["./proverclient", "-f", "etc/proverclient.yaml"]
