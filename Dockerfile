FROM golang:1.25.3 AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/prlint .

FROM gcr.io/distroless/base-debian12
WORKDIR /github/workspace
COPY --from=builder /bin/prlint /usr/local/bin/prlint
ENTRYPOINT ["/usr/local/bin/prlint"]
