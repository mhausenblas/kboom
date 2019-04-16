FROM golang:1.12-alpine AS build-env

RUN apk add git

RUN go get github.com/ericchiang/k8s

RUN mkdir -p /go/src/github.com/mhausenblas/kboom
WORKDIR /go/src/github.com/mhausenblas/kboom

COPY  . .
RUN adduser -D -u 10001 kboom
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' \
    -o kboom main.go

FROM scratch
COPY --from=build-env /go/src/github.com/mhausenblas/kboom/kboom .
COPY --from=build-env /etc/passwd /etc/passwd
USER kboom
ENTRYPOINT ["/kboom"]
