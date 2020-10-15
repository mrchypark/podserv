FROM golang:1.15.2-buster AS build
WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-w -s' -o podserv main.go

FROM gcr.io/distroless/static-debian10
EXPOSE 3000
COPY --from=build /app/podserv /
CMD ["/podserv"]