FROM golang:1.24 AS builder

WORKDIR /workspace
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

COPY ["go.mod", "go.sum", "./"]
RUN go mod download

COPY cmd/tracker /workspace/cmd/tracker
COPY internal/ /workspace/internal

RUN go build -trimpath -o /tracker ./cmd/tracker/

FROM scratch
COPY --from=builder /tracker /app/tracker

EXPOSE 8080
VOLUME [ "/config" ]
COPY config/tracker/config.toml /config/config.toml
CMD [ "/app/tracker", "-config", "/config/config.toml", "-listen", ":8080", "-telemetry"]
