FROM golang:1.16.3-alpine3.13 AS GO_BUILD
RUN apk update && apk add build-base
COPY server /server
WORKDIR /server
RUN go build -o /go/bin/server

FROM alpine:3.13.5
COPY --from=GO_BUILD /go/bin/server ./
CMD ./server
