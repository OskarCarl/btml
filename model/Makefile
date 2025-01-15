TAG ?= latest
IMAGE ?= pytorch/pytorch:$(TAG)
DOCKERFLAGS ?= -it --rm -v ./:/app -w /app --user $(shell id -u):$(shell id -g)
GPU_IMAGE ?= rocm/pytorch:$(TAG)
DOCKER_GPU_VERSION_OVERRIDE ?= -e HSA_OVERRIDE_GFX_VERSION="11.0.1"
DOCKER_GPUFLAGS ?= $(DOCKER_GPU_VERSION_OVERRIDE) --device=/dev/kfd:rw --device=/dev/dri --security-opt seccomp=unconfined --group-add video

docker-run-cpu: logs/
	docker run $(DOCKERFLAGS) $(IMAGE) python main.py

docker-run-amd: logs/
	docker run $(DOCKERFLAGS) $(DOCKER_GPUFLAGS) $(GPU_IMAGE) python main.py
docker-prepare-data: data/prepared/
	docker run $(DOCKERFLAGS) $(IMAGE) python tools/data.py --output-dir data/prepared --num-splits 100

setup: venv/
	./venv/bin/pip install -r requirements.txt

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
