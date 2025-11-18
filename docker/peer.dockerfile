FROM ghcr.io/astral-sh/uv:python3.11-trixie-slim AS model-builder
WORKDIR /workspace

RUN apt-get update && apt-get install -y protobuf-compiler
COPY pyproject.toml .python-version uv.lock README.md LICENSE /workspace/
RUN uv sync --only-dev
COPY protocols/peer-model.proto /workspace/
RUN mkdir -p model/lib/ipc/ && ./.venv/bin/python -m grpc_tools.protoc -Imodel/lib/ipc=. --python_out=. --pyi_out=. --grpc_python_out=. peer-model.proto

COPY model/*.py /workspace/model/

RUN uv build --wheel -o /workspace


FROM golang:1.24 AS go-builder
WORKDIR /workspace
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

RUN apt-get update && apt-get install -y protobuf-compiler protoc-gen-go protoc-gen-go-grpc

COPY protocols/ /workspace/protocols
RUN protoc --go_out=. --go-grpc_out=. -Iprotocols/ model-update.proto peer-model.proto

COPY ["go.mod", "go.sum", "./"]
RUN go mod download

COPY cmd/peer /workspace/cmd/peer
COPY internal/ /workspace/internal

RUN go build -trimpath -o /peer ./cmd/peer/


FROM pytorch/pytorch:2.9.0-cuda12.8-cudnn9-runtime AS app
WORKDIR /app

RUN python -m venv --system-site-packages venv

COPY --from=model-builder /workspace/model-*.whl /app/
RUN venv/bin/pip install /app/model-*.whl

COPY --from=go-builder /peer /app/peer

VOLUME ["/data", "/logs"]
ENV PYTHON_MODEL_LINE="venv/bin/python -m model"

CMD [ "/app/peer" ]
