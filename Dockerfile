FROM golang:1.19.5-alpine3.16 as builder
WORKDIR /evalbot
COPY . .
RUN go build -ldflags="-w -s" .
RUN rm -rf *.go && rm -rf go.*
FROM alpine:3.16.3
COPY --from=builder /evalbot/evalbot /evalbot
ENTRYPOINT ["/evalbot"]
