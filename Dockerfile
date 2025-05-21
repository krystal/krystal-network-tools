FROM node:22-alpine as frontend_build
WORKDIR /var/app
COPY frontend/package.json .
COPY frontend/package-lock.json .
RUN npm ci
COPY frontend .
RUN npm run build
RUN rm build/index.html

FROM golang:1.24.3-alpine as backend_build
WORKDIR /var/app
COPY backend/go.mod .
COPY backend/go.sum .
RUN go mod download
COPY backend .
COPY --from=frontend_build /var/app/build frontend/frontend_blobs
RUN go build -o main

FROM alpine:3.21
WORKDIR /var/app
ENV GIN_MODE=release
COPY --from=backend_build /var/app/main .
EXPOSE 8080
ENTRYPOINT ["/var/app/main", "http"]
