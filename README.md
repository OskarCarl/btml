# BitTorrent Distributed (Decentralized) Machine Learning

Can the BitTorrent architecture and choking mechanisms be viably used to facilitate distributed machine learning?

## Setup

### Docker

The simplest way to run this is with [Docker](https://www.docker.com/) and [Docker compose](https://docs.docker.com/compose/).
Run `docker compose up` and hit Ctrl-C to stop the containers.
This will start the tracker, some peers and the telemetry collector (InfluxDB) and visualizer (Grafana).

The Grafana interface is available on [localhost](http://localhost:3000) with the login credentials `admin` and `admin`.
You can find a pre-configured dashboard under `Dashboards/Default/Overview`.

The InfluxDB interface is also [exposed](http://localhost:8086) with the credentials `user` and `password`.


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

There is no telemetry collection or visualization configured for the native deployment.
