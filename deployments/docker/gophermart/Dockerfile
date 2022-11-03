FROM golang:alpine

RUN mkdir -p /opt/gophermart/cmd/gophermart
RUN mkdir -p /opt/gophermart/internal/gophermart

WORKDIR /opt/gophermart

COPY ./cmd/gophermart ./cmd/gophermart
COPY ./internal/gophermart ./internal/gophermart
COPY ./go.mod .
COPY ./go.sum .

RUN go build -o gophermart cmd/gophermart/main.go
RUN chmod +x ./gophermart

ENV RUN_ADDRESS=":8080"
ENV DATABASE_URI="postgres://postgres:postgres@postgresql:5432/gophermart?sslmode=disable"
ENV ACCRUAL_SYSTEM_ADDRESS="accural:8080"

ENTRYPOINT ["./gophermart"]
