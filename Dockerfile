FROM golang:1.15 AS build

WORKDIR /filstats-cli

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo .

FROM scratch
COPY --from=build /filstats-cli/filstats-cli /filstats-cli
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

ENTRYPOINT ["/filstats-cli"]
