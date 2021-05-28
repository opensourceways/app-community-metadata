FROM golang:alpine3.13 as builder

MAINTAINER tommylike<tommylikehu@gmail.com>
WORKDIR /app
ADD . /app
RUN go mod download
RUN CGO_ENABLED=0 go build -o git-metadata

#since git-sync doesn't have a binary release, we copy binary from official docker image
FROM k8s.gcr.io/git-sync/git-sync:v3.3.1 as gitsync
RUN echo "git-sync prepared"

FROM alpine/git:v2.30.2
# to fix mv recoginzed option T
RUN apk update --no-cache && apk add coreutils
WORKDIR /app
COPY --from=builder /app/git-metadata .
COPY --from=gitsync /git-sync .

ADD ./config ./config
VOLUME ["/data/logs","/data/repos"]
ENV PATH="/app:${PATH}"
ENV APP_ENV="prod"
ENTRYPOINT ["/app/git-metadata"]
