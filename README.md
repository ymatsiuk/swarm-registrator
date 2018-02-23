# Docker SWARM discorvery for Prometheus

The idea to make discovery and monitoring of your docker swarm services easier and as much automated as possible. Prometheus targets are discovered via consul. This tool runs as a services and register containers for every service. If container re-scheduled with new ip, the toll is going to take care of registering new container's ip under the same service.

## Getting Started

These instructions will get you a copy of the project up and running on your local machine for development and testing purposes. See deployment for notes on how to deploy the project on a live system.

### Prerequisites

Swarm is required

```
docker swarm init
```

## Deployment

```
docker build -t swarm-registrator .
docker stack deploy -c stack.yml test
```

## License

This project is licensed under the MIT License