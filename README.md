# OpenTelemetry Golang
A simple **Go application** instrumented with **OpenTelemetry**, using **Grafana Alloy** as the telemetry collector and the **Grafana Observability Stack**:

- 📈 Prometheus (Metrics)
- 🔍 Tempo (Traces)
- 📜 Loki (Logs)

The purpose of this repository is to demonstrate different ways of exporting metrics to Prometheus while keeping the same application instrumentation.

---

# How to test the application

```bash
# Run the containers
docker compose up -d --build

# Run the Application
go run .

# Make requests to the API
# You can make an anonymous roll
curl localhost:8000/rolldice
curl localhost:8000/rolldice/{player}

# Go to the browser
# Grafana dashboard
# user: admin
# password: admin
localhost:3000

# Explore in the left side
# Choose the data sources after you make some requests
# Logs -> Loki
# Metrics -> Prometheus
# Traces -> Tempo
```

--- 

# Architecture

The Go application runs directly on the host machine while the observability stack runs inside Docker containers.

```
                ┌─────────────────────┐
                │   Go Application    │
                │ (OpenTelemetry SDK) │
                └──────────┬──────────┘
                           │
                    OTLP / Metrics
                           │
                ┌──────────▼──────────┐
                │    Grafana Alloy    │
                └──────────┬──────────┘
                           │
      ┌────────────────────┼────────────────────┐
      │                    │                    │
      ▼                    ▼                    ▼
 Prometheus             Tempo                Loki
   Metrics              Traces               Logs
```

---

# Project Structure

```
.
├── app/                 # Go application
├── alloy/               # Grafana Alloy configuration
├── prometheus/          # Prometheus configuration
├── tempo/               # Tempo configuration
├── loki/                # Loki configuration
├── grafana/             # Grafana provisioning
└── docker-compose.yml
```

---

# Prometheus Collection Modes

Prometheus can receive metrics in three different ways:

| Mode | Description | Recommended For |
|------|-------------|-----------------|
| **Pull (Scrape)** | Prometheus periodically scrapes the application's `/metrics` endpoint. | Kubernetes, Docker, Service Discovery |
| **Push (Remote Write)** | An agent collects metrics and pushes them to Prometheus. | Distributed environments, firewalls, edge computing |
| **OTLP Receiver** | Applications send OTLP metrics directly to Prometheus. | Modern OpenTelemetry-native architectures |

---

# 1. Pull Mode (Scrape)

This is the **traditional Prometheus architecture**.

The application exposes a `/metrics` endpoint and Prometheus periodically performs HTTP GET requests to collect metrics.

```
Application
     │
     │  GET /metrics
     ▲
Prometheus
```

## Characteristics

Prometheus controls the entire collection process:

- Scrape interval
- Request timeout
- Retry strategy
- Target labels
- Relabeling rules
- Service discovery

This mode integrates naturally with:

- Kubernetes
- Docker
- Consul
- Nomad
- EC2
- Static targets

The only requirement is that Prometheus can reach the application's `/metrics` endpoint.

---

## Advantages

- Native Prometheus workflow
- Simple architecture
- Easy debugging
- Full control over scraping
- Powerful service discovery
- Automatic target management

### Natural Backpressure

If Prometheus becomes overloaded or temporarily slow, the application is **not affected**.

The only consequence is that scrape requests take longer or fail temporarily.

---

## Disadvantages

Prometheus must be able to initiate connections to every application.

This becomes problematic when applications are behind:

- Firewalls
- NAT
- VPNs
- Edge devices
- IoT devices
- Customer environments

As the number of services increases, Prometheus must maintain connections to all targets, making large-scale deployments more challenging.

---

# 2. Push Mode (Remote Write)

Instead of Prometheus pulling metrics, an agent collects and forwards them.

Common agents include:

- Grafana Alloy
- OpenTelemetry Collector
- Prometheus Agent

Workflow:

```
Node Exporter
      │
      ▼
Grafana Alloy
      │
      ▼
Remote Write
      │
      ▼
Prometheus
```

---

## Advantages

### Firewall Friendly

Only outbound traffic is required.

```
Agent ─────────► Prometheus
```

Prometheus never needs direct access to the monitored application.

This architecture works particularly well for:

- Multiple Kubernetes clusters
- Remote environments
- Edge computing
- Customer-hosted deployments
- Multi-region infrastructures

Each environment runs its own agent that forwards metrics to a centralized Prometheus.

---

## Disadvantages

Prometheus no longer controls metric collection.

Instead, the agent determines:

- Scrape interval
- Timeout
- Relabeling
- Target discovery
- Retry behavior

As a result:

- More components to debug
- Service discovery becomes the agent's responsibility
- Configuration becomes more distributed

---

# 3. OTLP Receiver

Prometheus can also receive metrics directly using the **OpenTelemetry Protocol (OTLP)**.

Unlike scrape mode, applications do **not** expose a `/metrics` endpoint.

Instead, metrics are sent in OTLP (protobuf) format.

```
Application
      │
      ▼
OpenTelemetry SDK
      │
      ▼
OTLP
      │
      ▼
Prometheus
```

or, when using a collector:

```
Application
      │
      ▼
OpenTelemetry SDK
      │
      ▼
Grafana Alloy
      │
      ▼
Prometheus OTLP Receiver
```

---

## Advantages

### Modern Instrumentation

The same OpenTelemetry SDK can generate:

- Metrics
- Traces
- Logs

This provides a unified instrumentation model.

### No Prometheus Exporter

The application only needs the OpenTelemetry SDK.

There is no need to expose a `/metrics` endpoint or use the Prometheus client library.

### Vendor Neutral

Applications become independent of the observability backend.

The same telemetry can be exported to:

- Prometheus
- Grafana Cloud
- Datadog
- New Relic
- Elastic
- Honeycomb
- Any OTLP-compatible backend

---

## Disadvantages

OpenTelemetry and Prometheus use different data models.

Some Prometheus concepts do not have a perfect OTLP equivalent, including:

- Certain metric types
- Label semantics
- Histogram representation

This can introduce slight differences in how metrics are stored or visualized.

### Collection Control

There is no scrape process.

Instead, the application or OpenTelemetry SDK decides:

- Export interval
- Batch size
- Retry strategy

Prometheus simply receives the exported data.

---

# Which Mode Should You Choose?

| Scenario | Recommended Mode |
|-----------|------------------|
| Kubernetes | ✅ Pull (Scrape) |
| Docker Compose | ✅ Pull (Scrape) |
| Internal microservices | ✅ Pull (Scrape) |
| Edge devices | ✅ Remote Write |
| Multiple clusters | ✅ Remote Write |
| Firewalled environments | ✅ Remote Write |
| Cloud-native OpenTelemetry | ✅ OTLP Receiver |
| Vendor-neutral instrumentation | ✅ OTLP Receiver |

---

# Stack

| Component | Purpose |
|-----------|---------|
| Grafana Alloy | Telemetry collection and routing |
| Prometheus | Metrics storage |
| Tempo | Distributed tracing |
| Loki | Log aggregation |
| Grafana | Visualization and dashboards |

---

# Tempo and Loki in This Project

This repository uses the same OpenTelemetry pipeline for traces and logs as it does for metrics, but with different backends.

## Tempo (Traces)

Tempo is the component responsible for storing and querying distributed traces.

In this project:

- The application emits trace data through the OpenTelemetry SDK.
- Grafana Alloy receives the OTLP trace data.
- Alloy forwards the traces to Tempo over OTLP/gRPC.
- Tempo is configured to listen on port 4317 (and 4318 for HTTP) and store traces locally.
- Grafana is then able to query those traces through the Tempo datasource.

This means traces are not scraped like Prometheus metrics. They are sent to the backend as telemetry data arrives.

## Loki (Logs)

Loki is the component responsible for collecting and indexing logs.

In this project:

- Logs are generated by the application and received by Grafana Alloy.
- Alloy forwards the log stream to Loki using the Loki push API.
- Loki stores the logs centrally so they can be searched and filtered from Grafana.
- Grafana uses the Loki datasource to visualize and query logs.

In short, Tempo handles traces and Loki handles logs, while Grafana Alloy acts as the central collector and router for both.

---

# Goal

This repository exists to demonstrate:

- OpenTelemetry instrumentation in Go
- Grafana Alloy pipelines
- Prometheus collection strategies
- OTLP-based telemetry
- Integration with the Grafana observability stack

It is intended as a learning resource rather than a production-ready deployment.
