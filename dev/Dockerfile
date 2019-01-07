FROM lushdigital/docker-golang-dep:latest as builder

ENV JANUS_BUILD_ONLY_DEFAULT 1
ENV VERSION 3.8.7-rc.2
RUN apk add --update bash make
RUN mkdir -p /go/src/github.com/hellofresh/janus
COPY ./ /go/src/github.com/hellofresh/janus/
RUN	cd /go/src/github.com/hellofresh/janus && \
	make build


FROM alpine

COPY --from=builder /go/src/github.com/hellofresh/janus/dist/janus /

RUN mkdir -p /etc/janus/apis && \
    mkdir -p /etc/janus/auth

RUN apk add --update curl && \
    rm -rf /var/cache/apk/*

EXPOSE 8080 8081 8443 8444
ENTRYPOINT ["/janus"]
