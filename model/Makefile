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

test-reqs: lib/model.py logs/

docker-run-cpu: logs/ lib/model.py
	docker run $(DOCKERFLAGS) $(IMAGE) python main.py $(MODEL_CMDLINE)

docker-run-amd: logs/ lib/model.py
	docker run $(DOCKERFLAGS) $(DOCKER_GPUFLAGS) $(GPU_IMAGE) python main.py $(MODEL_CMDLINE)

docker-prepare-data: data/prepared/
	docker run $(DOCKERFLAGS) $(IMAGE) python tools/data.py --output-dir data/prepared --num-splits 100

lib/model.py: lib/ ../protocols/peer-model.proto ../.venv/bin/protoc-gen-python_betterproto2
	protoc --plugin=../.venv/bin/protoc-gen-python_betterproto2 --python_betterproto2_out=./lib/ -I../protocols/ peer-model.proto

%/:
	mkdir -p $@

clean:
	rm -rf logs/ __pycache__/ **/__pycache__/

reset:
	rm -f logs/*.log
	rm -f logs/*.done
	rm -rf data/checkpoints/*

.PHONY: run docker-run-cpu docker-run-amd docker-prepare-data clean reset
