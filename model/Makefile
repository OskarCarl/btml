TAG ?= latest
IMAGE ?= btfl-model
DOCKERFLAGS ?= -it --rm -v ./:/app -w /app --user $(shell id -u):$(shell id -g)
GPU_IMAGE ?= rocm/pytorch:$(TAG)
DOCKER_GPU_VERSION_OVERRIDE ?= -e HSA_OVERRIDE_GFX_VERSION="11.0.1"
DOCKER_GPUFLAGS ?= $(DOCKER_GPU_VERSION_OVERRIDE) --device=/dev/kfd:rw --device=/dev/dri --security-opt seccomp=unconfined --group-add video

PEER_ID ?= 0
MODEL_CMDLINE ?= --train-data data/prepared/fMNIST_train_split_$(PEER_ID).pt --test-data data/prepared/fMNIST_test_split_$(PEER_ID).pt

test: .docker-build
	$(MAKE) docker-run-cpu MODEL_CMDLINE="$(MODEL_CMDLINE) --oneshot"

docker-run-cpu: logs/ peer_model_pb2.py .docker-build
	docker run $(DOCKERFLAGS) $(IMAGE) python main.py $(MODEL_CMDLINE)

docker-run-amd: logs/ peer_model_pb2.py
	docker run $(DOCKERFLAGS) $(DOCKER_GPUFLAGS) $(GPU_IMAGE) python main.py $(MODEL_CMDLINE)

docker-prepare-data: data/prepared/
	docker run $(DOCKERFLAGS) $(IMAGE) python tools/data.py --output-dir data/prepared --num-splits 100

peer_model_pb2.py: ../protocols/peer-model.proto
	protoc --python_out=. -I../protocols/ peer-model.proto

setup: venv/
	./venv/bin/pip install -r requirements.txt

.docker-build: Dockerfile
	docker build -t btfl-model:latest .
	echo "$(shell date --iso=seconds)" > .docker-build

venv/:
	python -m venv ./venv

%/:
	mkdir -p $@

clean:
	rm -rf venv/
	rm -rf logs/

reset:
	rm -f logs/*.log
	rm -f logs/*.done

.PHONY: run docker-run-cpu docker-run-amd docker-prepare-data setup venv clean reset
