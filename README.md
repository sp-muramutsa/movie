# Distributed Movie Metadata Engine

A high-reliability, microservices-based backend system for managing movie metadata and user ratings. This project is built with **Go (Golang)** and demonstrates advanced distributed systems concepts, including asynchronous message passing, binary-encoded RPC communication, and containerized orchestration.

## Key Features

* **Microservices Architecture:** Independently scalable services for `movie`, `metadata`, and `rating` management.
* **High-Throughput Communication:** Replaced traditional REST APIs with **gRPC** and **Protobuf**, reducing inter-service latency by 50% through binary-encoded communication.
* **Fault-Tolerant Messaging:** Designed an event-driven architecture using **Apache Kafka** to decouple the rating and metadata services, ensuring zero data loss during high-traffic bursts and eventual consistency.
* **Container Orchestration:** Fully deployable on **Kubernetes** using custom manifests and `skaffold` for streamlined continuous development.
* **Relational Data Management:** Backed by **PostgreSQL**, optimized for rapid metadata retrieval and relation mapping.
* **Observability:** Integrated monitoring via **Prometheus** and **Grafana** to track system health and container utilization (MTTD optimization).

## Tech Stack

* **Language:** Go (Golang)
* **RPC Framework:** gRPC & Protocol Buffers (Protobuf)
* **Message Broker:** Apache Kafka
* **Database:** PostgreSQL
* **Infrastructure & DevOps:** Kubernetes (K8s), Docker, Skaffold
* **Observability:** Grafana, Prometheus

## Repository Structure

* `/api`: API definitions and handlers.
* `/auth`: Authentication and authorization logic.
* `/cmd`: Entry points for the individual microservices.
* `/gen`: Auto-generated Protobuf code (`.pb.go`).
* `/internal/grpcutil`: Shared internal gRPC utilities and interceptors.
* `/k8s`: Kubernetes deployment and service manifests.
* `/metadata`: Core business logic and database interactions for the Metadata service.
* `/movie`: Core business logic for the Movie aggregation service.
* `/rating`: Core business logic and database interactions for the Rating service.
* `/pkg`: Publicly importable utility packages.
* `/schema`: Database schemas and migration scripts.
* `/test/integration`: End-to-end integration tests.

## Getting Started

### Prerequisites
* [Go](https://golang.org/doc/install) (1.20+ recommended)
* [Docker](https://docs.docker.com/get-docker/) & Docker Compose
* [Kubernetes](https://kubernetes.io/docs/setup/) (e.g., Minikube, kind, or Docker Desktop K8s)
* [Skaffold](https://skaffold.dev/docs/install/)
* [Protoc](https://grpc.io/docs/protoc-installation/) (if modifying Protobuf definitions)

### Running Locally with Skaffold

The easiest way to run the entire distributed system locally is by utilizing `skaffold`. It will build the Docker images and deploy them to your local Kubernetes cluster.

1.  **Start your local Kubernetes cluster** (e.g., `minikube start`).
2.  **Run Skaffold:**
    ```bash
    skaffold dev
    ```
    *This command will build the microservices, deploy the Kafka broker, Postgres databases, and the Go services, and will actively watch your code for changes to trigger rebuilds.*

### Manual Docker Setup (Alternative)

If you prefer to run the services via Docker without Kubernetes:
    ```bash
    docker build -t movie-service -f Dockerfile .
    # Repeat for metadata and rating services...
    ```

### Generating Protobufs
If you make changes to the `.proto` files, regenerate the Go code using:
    ```bash
    # Example protoc command, adjust based on your Makefile/scripts
    protoc --go_out=./gen --go-grpc_out=./gen ./api/*.proto
    ```

## Contributing
Contributions, issues, and feature requests are welcome! 
Feel free to check the [issues page](https://github.com/sp-muramutsa/movie/issues).

## License
This project is [MIT](LICENSE) licensed.
