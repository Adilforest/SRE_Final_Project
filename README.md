# SRE Final Project

A microservices-based e-commerce backend with full SRE tooling: Prometheus metrics, Grafana dashboards, k6 load tests, Kubernetes manifests, and Terraform infrastructure — built as a final project for the Site Reliability Engineering course at AITU.

![Go](https://img.shields.io/badge/Go-1.23-00ADD8?logo=go&logoColor=white)
![gRPC](https://img.shields.io/badge/gRPC-protobuf-4285F4?logo=google&logoColor=white)
![MongoDB](https://img.shields.io/badge/MongoDB-7.0-47A248?logo=mongodb&logoColor=white)
![Redis](https://img.shields.io/badge/Redis-7-DC382D?logo=redis&logoColor=white)
![NATS](https://img.shields.io/badge/NATS-2-199bdb?logo=natsdotio&logoColor=white)
![Prometheus](https://img.shields.io/badge/Prometheus-scrape-E6522C?logo=prometheus&logoColor=white)
![Grafana](https://img.shields.io/badge/Grafana-dashboard-F46800?logo=grafana&logoColor=white)
![Kubernetes](https://img.shields.io/badge/Kubernetes-minikube-326CE5?logo=kubernetes&logoColor=white)
![Terraform](https://img.shields.io/badge/Terraform-1.5-7B42BC?logo=terraform&logoColor=white)
![GitHub Actions](https://img.shields.io/badge/CI%2FCD-GitHub_Actions-2088FF?logo=githubactions&logoColor=white)

---

## Overview

This project implements a small e-commerce platform using a polyglot microservice architecture and demonstrates practical SRE discipline: observability, infrastructure-as-code, and automated deployment. It was built as a final assignment for the SRE course at Astana IT University.

Three backend services (auth, product, order) communicate via gRPC. An HTTP API gateway fronts them, exposes a `/metrics` endpoint consumed by Prometheus, and serves as the target for k6 load tests. A CLI tool (`main.go` at the repo root) uses the Kubernetes Go client to collect pod logs and scrape the gateway metrics endpoint for post-incident diagnostics.

---

## Features

- **JWT authentication** — register, email activation, login, logout with Redis token blacklist, password reset via SMTP email, refresh tokens
- **Product catalog** — CRUD for bicycle products with type (`road`, `mountain`, `hybrid`, `electric`), brand, price, stock, features, and full-text search
- **Order management** — create, approve, and cancel orders; per-user order listing
- **Prometheus instrumentation** — `http_requests_total` (method/path/status), `http_request_duration_seconds` (histogram), `grpc_client_connections_total` exposed at `GET /metrics`
- **k6 load test** — 10 VUs × 30 s targeting the health endpoint, ready to extend
- **Kubernetes deployment** — 3-replica `api-gateway` Deployment, Service, and Ingress manifests
- **Terraform IaC** — Kubernetes provider config for local Minikube provisioning
- **GitHub Actions CI/CD** — build → Docker Hub push → Terraform apply on every push to `main`
- **SRE diagnostics CLI** — Go program that uses `client-go` to collect pod logs from all services and snapshot Prometheus metrics into a timestamped directory

---

## Architecture

```
Client
  │
  ▼
api-gateway  (Gin HTTP :8080)
  ├── /metrics  ◄──── Prometheus scrape
  ├── /login, /register, /activate, /me, /logout, /forgot-password, /reset-password, /refresh-token
  ├── /products  (CRUD + search + stock)
  └── /orders    (create / get / list / cancel / approve)
       │     │     │
     gRPC  gRPC  gRPC
       │     │     │
  auth-svc  product-svc  order-svc
  (MongoDB + Redis)  (MongoDB)  (MongoDB + NATS)
```

Services publish and consume events over NATS. Auth-service uses MongoDB for user storage and Redis for token blacklisting.

---

## Tech Stack

| Layer | Technologies |
|---|---|
| Language | Go 1.23 |
| HTTP framework | Gin |
| Service communication | gRPC + Protocol Buffers |
| Messaging | NATS |
| Databases | MongoDB, Redis |
| Observability | Prometheus client, Grafana |
| Load testing | k6 |
| Containers | Docker |
| Orchestration | Kubernetes (Minikube), Terraform |
| CI/CD | GitHub Actions |
| Logging | Logrus (file + structured) |

---

## Project Structure

```
SRE_Final_Project/
├── main.go                         # SRE diagnostics CLI (client-go)
├── go.mod                          # Root module (k8s.io/client-go)
├── docker-compose.yml              # Prometheus + Grafana monitoring stack
├── loadtest.js                     # k6 load test script
├── api-gateway/
│   ├── cmd/main.go                 # HTTP gateway, Prometheus middleware, routes
│   ├── internal/
│   │   ├── handlers/               # Auth, product, order HTTP handlers
│   │   ├── service/                # Service layer proxying to gRPC clients
│   │   └── client/                 # gRPC client wrappers
│   └── proto/                      # Proto definitions + generated Go code
├── services/
│   ├── auth-service/               # gRPC :50051  MongoDB + Redis + SMTP
│   │   ├── internal/domain/        # User entity, roles, JWT helpers
│   │   ├── internal/usecase/       # Register, activate, login, reset password
│   │   └── internal/delivery/
│   │       ├── grpc/               # gRPC handler
│   │       └── nats/               # NATS publisher / subscriber
│   ├── product-service/            # gRPC :50052  MongoDB
│   │   └── internal/domain/        # Bike product entity + filter + features
│   └── order-service/              # gRPC :50053  MongoDB + NATS
│       └── internal/domain/        # Order + OrderItem entities
├── infra/
│   ├── k8s/                        # Deployment (3 replicas), Service, Ingress
│   └── terraform/                  # Kubernetes Terraform provider config
├── monitoring/
│   └── prometheus/prometheus.yml   # Scrape config for api-gateway
└── .github/workflows/deploy.yml    # CI: build → push → Terraform apply
```

---

## Getting Started

### Prerequisites

- Go 1.23+
- Docker and Docker Compose
- MongoDB and Redis running locally (or set via env vars)
- NATS server (optional, for event-driven features)

### Environment variables

Each service reads a `.env` file (see `api-gateway/.env.example` pattern). Key variables:

| Variable | Service | Description |
|---|---|---|
| `AUTH_SERVICE_ADDR` | api-gateway | gRPC address of auth service, e.g. `localhost:50051` |
| `PRODUCT_SERVICE_ADDR` | api-gateway | e.g. `localhost:50052` |
| `ORDER_SERVICE_ADDR` | api-gateway | e.g. `localhost:50053` |
| `MONGO_URI` | auth, product, order | MongoDB connection string |
| `MONGO_DB` | auth, product, order | Database name |
| `REDIS_ADDR` | auth | Redis address, e.g. `localhost:6379` |
| `SMTP_HOST/PORT/USER/PASS` | auth | SMTP credentials for activation/reset emails |

### Run monitoring stack

```bash
docker compose up -d   # starts Prometheus (:9090) and Grafana (:3000)
```

### Run the API gateway

```bash
cd api-gateway
go run cmd/main.go
# Gateway listens on :8080
# Metrics: http://localhost:8080/metrics
```

### Run a service (example: auth-service)

```bash
cd services/auth-service
go run cmd/main.go
# gRPC server on :50051
```

### Run load test

```bash
k6 run loadtest.js
```

### Run the SRE diagnostics tool

```bash
go run main.go
# Collects pod logs from k8s namespace "default" and snapshots /metrics
# Output saved to sre_diagnostics_<timestamp>/
```

---

## API Endpoints

| Method | Path | Description |
|---|---|---|
| `POST` | `/register` | Register a new user |
| `GET` | `/activate` | Activate account via token |
| `POST` | `/login` | Authenticate, receive JWT |
| `POST` | `/logout` | Blacklist token in Redis |
| `GET` | `/me` | Get current user profile |
| `POST` | `/forgot-password` | Send password reset email |
| `POST` | `/reset-password` | Set new password via reset token |
| `POST` | `/refresh-token` | Issue new access token |
| `GET` | `/products` | List products |
| `POST` | `/products/search` | Search products |
| `POST` | `/products` | Create product |
| `GET` | `/products/:id` | Get product by ID |
| `PUT` | `/products/:id` | Update product |
| `DELETE` | `/products/:id` | Delete product |
| `POST` | `/products/:id/stock` | Change stock quantity |
| `POST` | `/orders` | Create order |
| `GET` | `/orders/:id` | Get order by ID |
| `GET` | `/orders/user/:user_id` | List orders for user |
| `POST` | `/orders/:id/cancel` | Cancel order |
| `POST` | `/orders/:id/approve` | Approve order |
| `GET` | `/metrics` | Prometheus metrics |
| `GET` | `/health` | Health check |

---

Adil Ormanov — [GitHub](https://github.com/Adilforest)
