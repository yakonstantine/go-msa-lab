FROM golang:1.25 as build

WORKDIR /app

COPY go.mod go.sum* ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o /usrsrv ./cmd/app

FROM alpine:latest

COPY --from=build /usrsrv /usrsrv

CMD ["/usrsrv"]