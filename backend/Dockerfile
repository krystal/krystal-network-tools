FROM golang:1.17-alpine
WORKDIR /var/app
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN rm -rf frontend_blobs && mkdir frontend_blobs && touch frontend_blobs/.keep
RUN go build -o main

FROM alpine:3.15
WORKDIR /var/app
COPY --from=0 /var/app/main .
EXPOSE 8080
ENTRYPOINT ["/var/app/main"]
