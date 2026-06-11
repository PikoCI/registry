FROM golang:1.25.1 AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

ARG VERSION=dev
ARG COMMIT=unknown

RUN CGO_ENABLED=0 go build -ldflags "-X github.com/pikoci/pkreg/cmd.Version=${VERSION} -X github.com/pikoci/pkreg/cmd.Commit=${COMMIT}" -o /pkreg .

FROM alpine:3.21

RUN apk --no-cache add ca-certificates
COPY --from=builder /pkreg /usr/local/bin/pkreg

EXPOSE 8080
ENTRYPOINT ["pkreg"]
CMD ["server"]
