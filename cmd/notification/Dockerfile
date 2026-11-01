FROM golang:1.27.3-bookworm AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . . 
RUN  CGO_ENABLED=0 go build -o /app/cmd/notification/exe /app/cmd/notification

FROM alpine:3.23
WORKDIR /app
COPY --from=builder /app/cmd/notification/exe /app
CMD [ "./exe" ]