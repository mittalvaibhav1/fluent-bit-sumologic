FROM golang:1.20-buster AS builder
WORKDIR /fluent-bit-sumologic
COPY ./out_sumologic .
RUN go mod tidy
RUN go build -buildmode=c-shared -o out_sumologic.so main.go

FROM fluent/fluent-bit:2.2
COPY ./config /fluent-bit/config
COPY --from=builder /fluent-bit-sumologic/out_sumologic.so /fluent-bit/plugins/
CMD ["/fluent-bit/bin/fluent-bit", "-c", "/fluent-bit/config/fluent-bit.conf", "-e", "/fluent-bit/plugins/out_sumologic.so"]
