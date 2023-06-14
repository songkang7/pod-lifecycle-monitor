FROM golang:1.19 AS build


WORKDIR /app
COPY . .

RUN go build -o app main.go

FROM debian:buster-slim
WORKDIR /app
COPY --from=build /app/app .
CMD ["./app"]