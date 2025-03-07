GOFLAGS ?= -trimpath
IMAGE ?= btml-model
DOCKERFLAGS ?= -it --rm -v ./:/app -w /app --user $(shell id -u):$(shell id -g)
PROTOBUFS = internal/model/peer-model.pb.go internal/peer/model-update.pb.go

all: bin/test-model bin/tracker bin/peer

test-model: bin/test-model
	@$(MAKE) -C model/ test-reqs
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

setup-model:
	@$(MAKE) -C model/ test-reqs

internal/peer/model-update.pb.go: protocols/model-update.proto
	protoc --go_out=. -Iprotocols/ model-update.proto

internal/model/peer-model.pb.go: protocols/peer-model.proto
	protoc --go_out=. -Iprotocols/ peer-model.proto

prep-kernel:
	sysctl -w net.core.rmem_max=7500000
	sysctl -w net.core.wmem_max=7500000

%/:
	mkdir -p $@

clean:
	rm -rf bin/ logs/
	@$(MAKE) -C model/ clean

reset:
	$(MAKE) -C model/ reset

.PHONY: clean test-go
