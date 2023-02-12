FROM golang:1.19.5 as builder
WORKDIR /go/src/github.com/dddpaul/alfafin-bot
ADD . ./
RUN make build-alpine

FROM alpine:3.16.3
LABEL maintainer="Pavel Derendyaev <dddpaul@gmail.com>"
RUN apk add --update ca-certificates && \
    rm -rf /var/cache/apk/* /tmp/* && \
    update-ca-certificates
WORKDIR /app
COPY --from=builder /go/src/github.com/dddpaul/alfafin-bot/bin/bot .
#EXPOSE 8080

ENTRYPOINT ["./bot"]
#CMD ["-port", ":8080"]
