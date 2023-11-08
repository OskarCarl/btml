GOFLAGS ?= -trimpath

test-run: logs/ bin/test-runner bin/peer bin/tracker
	rm -rf logs/*.log
	bin/test-runner -n 5

bin/test-runner: bin/ cmd/test-runner/*.go
	go build $(GOFLAGS) -o bin/test-runner ./cmd/test-runner

bin/peer: bin/ cmd/peer/*.go internal/peer/*.go
	go build $(GOFLAGS) -o bin/peer ./cmd/peer

bin/tracker: bin/ cmd/tracker/*.go internal/tracker/*.go
	go build $(GOFLAGS) -o bin/tracker ./cmd/tracker

%/:
	mkdir -p $@/

clean:
	rm -rf bin/ logs/
