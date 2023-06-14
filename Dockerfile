FROM golang:1.19 AS build-env


WORKDIR /app
COPY . .

RUN go build -o pod-lifecycle-monitor main.go

FROM alpine:3.15.3

COPY --from=build-env /app/pod-lifecycle-monitor/ /
RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2
COPY --from=build-env /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
ENV TZ "Asia/Shanghai"

RUN addgroup -g 1000 nonroot && \
    adduser -u 1000 -D -H -G nonroot nonroot && \
    chown -R nonroot:nonroot /pod-lifecycle-monitor
USER nonroot:nonroot

ENTRYPOINT ["/pod-lifecycle-monitor"]