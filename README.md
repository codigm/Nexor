<div align="center">
  
# 🏛️ OLYMPUS
**A High-Performance Central Banking Settlement & Fraud Detection Engine**

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![Java Version](https://img.shields.io/badge/Java-17+-ED8B00?style=flat&logo=openjdk)](https://openjdk.java.net/)
[![Node.js Version](https://img.shields.io/badge/Node.js-18+-339933?style=flat&logo=node.js)](https://nodejs.org/)
[![Kafka](https://img.shields.io/badge/Kafka-Event_Driven-231F20?style=flat&logo=apachekafka)](https://kafka.apache.org/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-Relational_DB-336791?style=flat&logo=postgresql)](https://www.postgresql.org/)
[![Redis](https://img.shields.io/badge/Redis-Cache_%26_Graph-DC382D?style=flat&logo=redis)](https://redis.io/)
[![Docker](https://img.shields.io/badge/Docker-Containerized-2496ED?style=flat&logo=docker)](https://www.docker.com/)

</div>

---

## 📖 Overview

**OLYMPUS** is a complex, microservices-based distributed system built to simulate a high-throughput central banking settlement network (similar to RTGS and NEFT). It orchestrates transactions between disparate bank systems, handles distributed ledgers, performs real-time fraud detection using graph processing, and leverages analytics.

This project was built to demonstrate architectural patterns such as **Event-Driven Architecture**, **Distributed Saga Pattern**, **Polyglot Microservices**, and **AI Integration** in a highly concurrent environment.

## ✨ Key Features

- 🏦 **Distributed Settlements**: Manages cross-bank financial routing utilizing the Saga pattern to ensure data consistency across multiple independent mock banking endpoints.
- 🕵️ **Real-time Fraud Detection**: A graph-based engine that monitors Kafka event streams in real-time to score transactions for fraud, utilizing behavioral fingerprinting.
- 📊 **Streaming Analytics**: Real-time aggregation of transaction data, monitoring throughput and system health.
- 📈 **Full Observability**: Instrumentated with Prometheus and Grafana for real-time dashboards and alerting.

## 🏗️ Architecture

```mermaid
flowchart TD
    Client[Client / Demo Script] -->|HTTP POST| API[Olympus API Gateway\n(Java/Spring)]
    
    API -->|Route/Saga Init| Hermes[Hermes Core\nSettlement Engine (Go)]
    API -->|Publish Event| Kafka[(Apache Kafka)]
    
    Hermes -->|Saga Commits| Banks[(Mock Banks\nAlpha, Beta, Gamma)]
    Hermes -->|Ledger State| DB[(PostgreSQL)]
    Hermes -->|Cache/Lock| Redis[(Redis)]
    
    Kafka -->|Stream| Argus[Argus Core\nFraud Engine (Go)]
    Kafka -->|Stream| Titan[Titan Core\nAnalytics (Node.js)]
    
    Argus -->|Graph Query| DB
    Argus -->|Graph Processing| Redis
    Argus -->|Explain Request| Athena[Athena Core\nLLM Engine (Go)]
    
    Prometheus[Prometheus] -->|Scrape Metrics| API
    Prometheus -->|Scrape Metrics| Hermes
    Prometheus -->|Scrape Metrics| Argus
    Grafana[Grafana] -->|Visualize| Prometheus
```

The system utilizes a polyglot monorepo approach, choosing the best tool for each domain:

| Service | Technology | Description |
|---------|------------|-------------|
| **Hermes Core** | `Go` | The Settlement & Routing engine handling distributed transactions (Sagas) and ISO20022 message parsing. |
| **Argus Core** | `Go` | The Audit & Fraud engine listening to Kafka streams, performing graph analysis and fraud scoring. |
| **Athena Core** | `Go` | A high-performance mock inference engine simulating continuous batching. |
| **Titan Core** | `Node.js` | Real-time analytics and reporting engine consuming transaction events. |
| **Olympus API** | `Java (Spring Boot)` | The main API Gateway serving as the entry point for financial transactions. |
| **The Banks** | `Java (Spring Boot)` | Mock financial institutions (`alpha-bank`, `beta-bank`, etc.) participating in the settlement network. |

### Infrastructure Stack
- **Message Broker:** Apache Kafka & Zookeeper
- **Primary Database:** PostgreSQL
- **Caching & Graph Store:** Redis
- **Monitoring:** Prometheus & Grafana
- **Deployment:** Docker & Docker Compose

## 📸 System Previews

### Real-time Monitoring (Grafana)
![Grafana Dashboard](assets/grafana-dashboard.png)

### API & Transaction Testing
![Terminal Execution](assets/terminal-screenshot.png)

## 🚀 Getting Started

The entire infrastructure and application stack is fully containerized and easily managed via `make`.

### Prerequisites
- Docker and Docker Compose
- Java 17+ (for local test execution)
- Make

### Quickstart

#### Mac/Linux (Using Make)

1. **Spin up the infrastructure & services:**
   ```bash
   make all
   ```
   *This will build the Docker images and start Postgres, Redis, Kafka, Zookeeper, Prometheus, Grafana, and all microservices.*

2. **Run an End-to-End Demo Transaction:**
   ```bash
   make demo
   ```
   *This script simulates RTGS and NEFT cross-bank transfers through the API Gateway, triggering the full settlement and fraud detection pipeline.*

3. **View Real-Time Observability:**
   Open your browser and navigate to `http://localhost:3000` (Grafana) to view the live dashboard (Login: `admin` / `admin`).

4. **Tear Down:**
   ```bash
   make clean
   ```

#### Windows (Using PowerShell)

Since `make` is not natively installed on Windows, you can use the direct Docker Compose commands.

1. **Spin up the infrastructure & services:**
   ```powershell
   docker-compose up --build -d
   ```

2. **Run an End-to-End Demo Transaction:**
   Run these two commands in PowerShell to simulate the transfers:
   ```powershell
   # RTGS Transfer (>= 10,000)
   Invoke-RestMethod -Uri "http://localhost:8080/alpha/api/transfer" -Method Post -ContentType "application/json" -Body '{"senderAccount": "ALPHA001", "receiverAccount": "BETA001", "amount": 15000.00, "currency": "INR", "description": "Demo RTGS cross-bank transfer"}'

   # NEFT Transfer (< 10,000)
   Invoke-RestMethod -Uri "http://localhost:8080/alpha/api/transfer" -Method Post -ContentType "application/json" -Body '{"senderAccount": "ALPHA002", "receiverAccount": "GAMMA001", "amount": 5000.00, "currency": "INR", "description": "Demo NEFT cross-bank transfer"}'
   ```

3. **Tear Down:**
   ```powershell
   docker-compose down -v --remove-orphans
   ```

## 🧪 Testing

To run the automated test suite across the Java banking modules:
```bash
make test
```
