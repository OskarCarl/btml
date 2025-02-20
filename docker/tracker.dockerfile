FROM golang:1.23 AS builder

WORKDIR /workspace
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

COPY ["go.mod", "go.sum", "./"]
RUN go mod download

COPY cmd/tracker /workspace/cmd/tracker
COPY internal/ /workspace/internal

RUN go build -trimpath -o /tracker ./cmd/tracker/

FROM scratch
COPY --from=builder /tracker /tracker

EXPOSE 8080
CMD [ "/tracker", "-listen", ":8080"]
