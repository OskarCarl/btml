FROM python:3.12-slim AS python-proto-builder
WORKDIR /workspace

RUN apt-get update && apt-get install -y protobuf-compiler

COPY model/requirements-dev.txt .
RUN pip install -r requirements-dev.txt

COPY protocols/ /workspace/protocols

RUN protoc --python_betterproto_out=./ -I./protocols/ peer-model.proto


FROM golang:1.23 AS go-builder
WORKDIR /workspace
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

RUN apt-get update && apt-get install -y protobuf-compiler protoc-gen-go

COPY protocols/ /workspace/protocols
RUN protoc --go_out=. -Iprotocols/ model-update.proto peer-model.proto

COPY ["go.mod", "go.sum", "./"]
RUN go mod download

COPY cmd/peer /workspace/cmd/peer
COPY internal/ /workspace/internal

RUN go build -trimpath -o /peer ./cmd/peer/


FROM pytorch/pytorch:latest AS app
WORKDIR /app

COPY model/requirements.txt /app/
RUN pip install -r requirements.txt

COPY model/*.py /app/model/
COPY --from=go-builder /peer /app/peer
COPY --from=python-proto-builder /workspace/model.py /app/model/lib/

VOLUME ["/data", "/logs"]

CMD [ "/app/peer" ]
