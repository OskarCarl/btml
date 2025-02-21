# BitTorrent Distributed (Decentralized) Machine Learning

Can the BitTorrent architecture and choking mechanisms be viably used to facilitate distributed machine learning?

## Setup

### Docker

The simplest way to run this is with [Docker](https://www.docker.com/) and [Docker compose](https://docs.docker.com/compose/).
Run `docker compose up` and hit Ctrl-C to stop the containers.
This will start the tracker, some peers and the metric collector (InfluxDB) and visualizer (Grafana).


### Native

This project requires some dependencies to be installed:
- Go
- [Protobuf Compiler](https://protobuf.dev/) (protoc)
	- With the [protobuf-go](https://github.com/protocolbuffers/protobuf-go) plugin
- Python
	- For [PyTorch](https://pytorch.org/) and [Betterproto](https://github.com/danielgtaylor/python-betterproto)
- [Make](https://www.gnu.org/software/make/) to ease the setup and build process

The tracker can be built and run with `make test-tracker`.
Then you can start peers with `make test-peer`. This will also set up the Python model requirements.
