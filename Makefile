GOFLAGS ?= -trimpath

test-run: logs/ bin/test-runner bin/peer bin/tracker
	rm -rf logs/*.log
	bin/test-runner -n 5

bin/test-runner: bin/ cmd/test-runner/*.go
	go build $(GOFLAGS) -o bin/test-runner ./cmd/test-runner

cmd/peer/main.go: internal/peer/*.go

cmd/tracker/main.go: internal/tracker/*.go

bin/tracker bin/peer: bin/%: cmd/%/main.go bin/ internal/structs/*.go internal/logging/*.go
	go build $(GOFLAGS) -o $@ ./cmd/$*

%/:
	mkdir -p $@

clean:
	rm -rf bin/ logs/
