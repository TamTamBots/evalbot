FROM golang:1.20.4-alpine3.16 as builder
WORKDIR /evalbot
COPY . .
RUN go build -ldflags="-w -s" .
RUN rm -rf *.go && rm -rf go.*
FROM alpine:3.18.0
COPY --from=builder /evalbot/evalbot /evalbot
ENTRYPOINT ["/evalbot"]
