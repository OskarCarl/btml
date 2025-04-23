TAG ?= latest
IMAGE ?= btml-model
DOCKERFLAGS ?= -it --rm -v ./:/app -w /app --user $(shell id -u):$(shell id -g)
GPU_IMAGE ?= rocm/pytorch:$(TAG)
DOCKER_GPU_VERSION_OVERRIDE ?= -e HSA_OVERRIDE_GFX_VERSION="11.0.1"
DOCKER_GPUFLAGS ?= $(DOCKER_GPU_VERSION_OVERRIDE) --device=/dev/kfd:rw --device=/dev/dri --security-opt seccomp=unconfined --group-add video

PEER_ID ?= 0
MODEL_CMDLINE ?= --oneshot --train-data data/prepared/fMNIST_train_split_$(PEER_ID).pt --test-data data/prepared/fMNIST_test_split_$(PEER_ID).pt

test: test-reqs
	$(MAKE) docker-run-cpu MODEL_CMDLINE="$(MODEL_CMDLINE) --oneshot"

test-reqs: .setup .docker-build lib/model.py logs/

docker-run-cpu: logs/ lib/model.py .docker-build
	docker run $(DOCKERFLAGS) $(IMAGE) python main.py $(MODEL_CMDLINE)

docker-run-amd: logs/ lib/model.py
	docker run $(DOCKERFLAGS) $(DOCKER_GPUFLAGS) $(GPU_IMAGE) python main.py $(MODEL_CMDLINE)

docker-prepare-data: data/prepared/
	docker run $(DOCKERFLAGS) $(IMAGE) python tools/data.py --output-dir data/prepared --num-splits 100

lib/model.py: lib/ ../protocols/peer-model.proto venv/bin/protoc-gen-python_betterproto
	protoc --plugin=venv/bin/protoc-gen-python_betterproto --python_betterproto_out=./lib/ -I../protocols/ peer-model.proto

venv/bin/protoc-gen-python_betterproto: .setup

.setup: venv/ requirements.txt requirements-dev.txt
	./venv/bin/pip install -r requirements.txt -r requirements-dev.txt
	echo "$(shell date --iso=seconds)" > .setup

.docker-build: Dockerfile
	docker build -t $(IMAGE):latest .
	echo "$(shell date --iso=seconds)" > .docker-build

venv/:
	python -m venv ./venv

%/:
	mkdir -p $@

clean:
	rm -rf venv/ logs/ __pycache__/ **/__pycache__/ .setup

reset:
	rm -f logs/*.log
	rm -f logs/*.done
	rm -rf data/checkpoints/*

.PHONY: run docker-run-cpu docker-run-amd docker-prepare-data clean reset
