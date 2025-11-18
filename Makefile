GOFLAGS ?= -trimpath
IMAGE ?= btml-model
DOCKERFLAGS ?= -it --rm -v ./:/app -w /app --user $(shell id -u):$(shell id -g)
PROTOBUFS = internal/model/peer-model.pb.go internal/peer/model-update.pb.go
DIAGRAMS_FORMAT ?= pdf
DIAGRAMS = $(patsubst %.mmd,%.$(DIAGRAMS_FORMAT),$(wildcard docs/diagrams/*.mmd))

all: bin/test-model bin/tracker bin/peer

proto: internal/peer/model-update.pb.go internal/model/peer-model.pb.go model/lib/ipc/peer_model_pb2.py

all-diagrams: $(DIAGRAMS)

$(DIAGRAMS): %.$(DIAGRAMS_FORMAT):%.mmd
	docker run --rm -u $(shell id -u):$(shell id -g) -v ./docs/diagrams:/data ghcr.io/mermaid-js/mermaid-cli/mermaid-cli \
	-t neutral -b transparent -f -e $(DIAGRAMS_FORMAT) -i $(notdir $(@:.$(DIAGRAMS_FORMAT)=.mmd)) -o $(notdir $@)

test-model: bin/test-model
	docker run $(DOCKERFLAGS) $(IMAGE) bin/test-model

test-go:
	go test ./... -v

test-tracker: bin/tracker
	bin/tracker -config config/tracker/config.toml

test-peer: bin/peer setup-model
	bin/peer -autoconf -python venv/bin/python

bin/test-model: bin/ cmd/test-model/*.go internal/model/*.go internal/model/peer-model.pb.go
	go build $(GOFLAGS) -o bin/test-model ./cmd/test-model

bin/tracker bin/peer: bin/ internal/structs/*.go internal/logging/*.go
	go build $(GOFLAGS) -o $@ ./cmd/$(subst bin/,,$@)

bin/tracker: cmd/tracker/*.go internal/tracker/*.go
bin/peer: cmd/peer/*.go internal/peer/*.go internal/model/*.go internal/trust/*.go $(PROTOBUFS)

internal/peer/model-update.pb.go: protocols/model-update.proto
	protoc --go_out=. -Iprotocols/ model-update.proto

internal/model/peer-model.pb.go: protocols/peer-model.proto
	protoc --go_out=. --go-grpc_out=. -Iprotocols/ peer-model.proto

model/lib/ipc/peer_model_pb2.py: model/lib/ipc/ protocols/peer-model.proto
	python -m grpc_tools.protoc -Imodel/lib/ipc=protocols/ --python_out=. --pyi_out=. --grpc_python_out=. ./protocols/peer-model.proto

prep-kernel:
	sysctl -w net.core.rmem_max=7500000
	sysctl -w net.core.wmem_max=7500000

%/:
	mkdir -p $@

libs:
	go mod download
	uv sync

clean: reset
	rm -rf bin/ build/ logs/* .venv/ .ruff_cache/
	rm -rf *.egg-info *.whl **/__pycache__/
	rm -f docs/diagrams/*.{pdf,png,svg}

reset:
	rm -rf model/data/checkpoints/*
	rm -f logs/*.log logs/*.done
	rm -f config/tracker/.token

.PHONY: clean reset test-go proto libs
