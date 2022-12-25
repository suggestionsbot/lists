FROM golang:1.19-alpine AS build

WORKDIR /go/src/lists

COPY . .

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o main .

FROM alpine:latest

WORKDIR /opt/lists

COPY --from=build /go/src/lists .

EXPOSE 3000

CMD ["./main"]