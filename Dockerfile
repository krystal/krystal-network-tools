FROM node:16-alpine
WORKDIR /var/app
COPY frontend/package.json .
COPY frontend/package-lock.json .
RUN npm ci
COPY frontend .
RUN npm run build
RUN rm build/index.html

FROM golang:1.17-alpine
WORKDIR /var/app
COPY backend/go.mod .
COPY backend/go.sum .
RUN go mod download
COPY backend .
COPY --from=0 /var/app/build frontend_blobs
RUN go build -o main

FROM alpine:3.15
WORKDIR /var/app
ENV GIN_MODE=release
COPY --from=1 /var/app/main .
EXPOSE 8080
ENTRYPOINT ["/var/app/main"]
