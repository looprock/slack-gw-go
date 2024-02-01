from golang:1.19.7-alpine3.17 as builder
WORKDIR /slack-gw
COPY ./go.mod .
COPY ./go.sum .
RUN go mod download

COPY . .

RUN go build -o slack-gw ./slack-gw.go

FROM alpine:3.17

RUN mkdir /app
WORKDIR /app
COPY --from=builder /slack-gw/slack-gw /app/
RUN chmod +x /app/slack-gw
CMD /app/slack-gw
