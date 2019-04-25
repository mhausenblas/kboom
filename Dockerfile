FROM alpine:3.8 as build-kubectl
RUN set -x                  && \
    apk --update upgrade    && \
    apk add ca-certificates && \
    rm -rf /var/cache/apk/* && \
    wget -O /kubectl https://storage.googleapis.com/kubernetes-release/release/v1.14.0/bin/linux/amd64/kubectl && \
    chmod +x /kubectl

FROM golang:1.12-alpine AS build-env

RUN apk add git                                   && \
    go get github.com/ericchiang/k8s              && \
    go get github.com/mhausenblas/kubecuddler     && \
    go get github.com/jamiealquiza/tachymeter     && \
    mkdir -p /go/src/github.com/mhausenblas/kboom

WORKDIR /go/src/github.com/mhausenblas/kboom

COPY  . .
RUN adduser -D -u 10001 kboom \
    && CGO_ENABLED=0 GOOS=linux go build \
       -a -ldflags '-extldflags "-static"' -o kboom .

FROM scratch
COPY --from=build-kubectl /kubectl .
COPY --from=build-env /go/src/github.com/mhausenblas/kboom/kboom .
COPY --from=build-env /etc/passwd /etc/passwd
USER kboom
ENTRYPOINT ["/kboom"]
