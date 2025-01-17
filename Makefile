GOFLAGS ?= -trimpath
IMAGE ?= btfl-model
DOCKERFLAGS ?= -it --rm -v ./:/app -w /app --user $(shell id -u):$(shell id -g)

test-model: bin/test-model
	docker run $(DOCKERFLAGS) $(IMAGE) bin/test-model

test-run: logs/ bin/test-runner bin/peer bin/tracker
	rm -rf logs/*.log
	bin/test-runner -n 5

test-go:
	go test ./... -v

bin/test-runner: bin/ cmd/test-runner/*.go
	go build $(GOFLAGS) -o bin/test-runner ./cmd/test-runner

bin/test-model: bin/ cmd/test-model/*.go internal/model/*.go
	go build $(GOFLAGS) -o bin/test-model ./cmd/test-model

cmd/peer/main.go: internal/peer/*.go internal/model/*.go internal/trust/*.go internal/model/peer-model.pb.go

cmd/tracker/main.go: internal/tracker/*.go

bin/tracker bin/peer: bin/%: cmd/%/main.go bin/ internal/structs/*.go internal/logging/*.go
	go build $(GOFLAGS) -o $@ ./cmd/$*

internal/model/peer-model.pb.go: protocols/peer-model.proto
	protoc --go_out=. -Iprotocols/ peer-model.proto

%/:
	mkdir -p $@

clean:
	rm -rf bin/ logs/

.PHONY: clean test-go
