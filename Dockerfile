FROM golang:1.18.10-alpine as builder

RUN apk add --no-cache make git bash

ADD . /zkbnb

ENV CGO_ENABLED=0
ENV GO111MODULE=on

RUN cd /zkbnb && make build-only

# Pull ZkBNB into a second stage deploy alpine container
FROM alpine:3.16.0

ARG USER=bsc
ARG USER_UID=1000
ARG USER_GID=1000

ENV PACKAGES ca-certificates~=20220614 bash~=5.1.16-r2
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
