FROM --platform=${BUILDPLATFORM} public.ecr.aws/docker/library/golang:1.20.3-alpine3.17 as builder

ARG BUILDPLATFORM
ARG TARGETPLATFORM
ARG TARGETARCH
ARG GIT_TOKEN


RUN apk add --no-cache make git bash

WORKDIR /zkbnb
ADD . .

ENV CGO_ENABLED=0
ENV GO111MODULE=on

RUN go install github.com/zeromicro/go-zero/tools/goctl@latest && make api-server

RUN if [[ "$TARGETARCH" == "arm64" ]] ; then \
    wget -P ~ https://musl.cc/aarch64-linux-musl-cross.tgz && \
    tar -xvf ~/aarch64-linux-musl-cross.tgz -C ~ && \
    GOOS=linux GOARCH=${TARGETARCH} CGO_ENABLED=1 CC=~/aarch64-linux-musl-cross/bin/aarch64-linux-musl-gcc go build -o build/bin/zkbnb -ldflags="-X main.version=${VERSION} -X main.gitCommit=${GIT_COMMIT} -X main.gitDate=${GIT_COMMIT_DATE}" ./cmd/zkbnb ; \
  else \
    GOOS=linux GOARCH=${TARGETARCH} go build -o build/bin/zkbnb -ldflags="-X main.version=${VERSION} -X main.gitCommit=${GIT_COMMIT} -X main.gitDate=${GIT_COMMIT_DATE}" ./cmd/zkbnb ; \
  fi

# Pull ZkBNB into a second stage deploy alpine container
FROM --platform=${TARGETPLATFORM} public.ecr.aws/docker/library/alpine:3.17.3

ARG USER=bsc
ARG USER_UID=1000
ARG USER_GID=1000

ENV PACKAGES ca-certificates bash
ENV WORKDIR=/server

RUN apk add --no-cache $PACKAGES \
  && rm -rf /var/cache/apk/* \
  && addgroup -g ${USER_GID} ${USER} \
  && adduser -u ${USER_UID} -G ${USER} --shell /sbin/nologin --no-create-home -D ${USER} \
  && addgroup ${USER} tty \
  && sed -i -e "s/bin\/sh/bin\/bash/" /etc/passwd  

RUN echo "[ ! -z \"\$TERM\" -a -r /etc/motd ] && cat /etc/motd" >> /etc/bash/bashrc

WORKDIR ${WORKDIR}

COPY --from=builder /zkbnb/build/bin/zkbnb ${WORKDIR}/
RUN chown -R ${USER_UID}:${USER_GID} ${WORKDIR}
USER ${USER_UID}:${USER_GID}

ENTRYPOINT ["/server/zkbnb"]
