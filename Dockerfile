FROM golang:1.20.2-alpine3.16 as builder
WORKDIR /evalbot
COPY . .
RUN go build -ldflags="-w -s" .
RUN rm -rf *.go && rm -rf go.*
FROM alpine:3.17.2
COPY --from=builder /evalbot/evalbot /evalbot
ENTRYPOINT ["/evalbot"]
