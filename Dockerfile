FROM golang:latest as build
WORKDIR /app
COPY go.* ./
COPY cmd/main.go ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o temperatura

FROM scratch
WORKDIR /app
COPY --from=build /app/temperatura .
ENTRYPOINT ["./temperatura"]
