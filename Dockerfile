FROM golang:1.26-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /raft-node ./cmd/node

FROM alpine:latest
COPY --from=build /raft-node /raft-node
ENTRYPOINT ["/raft-node"]