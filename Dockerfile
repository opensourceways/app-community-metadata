FROM golang:alpine3.13 as builder

LABEL maintainer="tommylike<tommylikehu@gmail.com>"
WORKDIR /app
ADD . /app
RUN go mod download
RUN CGO_ENABLED=0 go build -o git-metadata

#since git-sync doesn't have a binary release, we copy binary from official docker image
FROM k8s.gcr.io/git-sync/git-sync:v3.3.1 as gitsync
RUN echo "git-sync prepared"

FROM alpine/git:v2.30.2
ARG user=app
ARG group=app
ARG home=/app
# to fix mv unrecoginzed option T
RUN apk update --no-cache && apk add --no-cache coreutils=8.32-r2 \
 && addgroup -S ${group} && adduser -S ${user} -G ${group} -h ${home}

USER ${user}
WORKDIR ${home}
COPY --chown=${user} --from=builder /app/git-metadata .
COPY --chown=${user} --from=gitsync /git-sync .
COPY --chown=${user} ./config .
#to fix the directory permission issue
RUN mkdir -p ${home}/logs $$ -p ${home}/repos
VOLUME ["${home}/logs","${home}/repos"]

ENV PATH="${home}:${PATH}"
ENV APP_ENV="prod"
EXPOSE 9500
ENTRYPOINT ["/app/git-metadata"]
