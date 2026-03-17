---
name: observability-setup
description: Use when setting up monitoring, logging, alerting, or observability for applications or infrastructure. Orchestrates monitoring, cloud, and kubernetes agents to create complete observability stacks. MUST BE USED when user mentions metrics, dashboards, alerts, SLOs, logging setup, or tracing configuration.
---

# Observability Setup — Complete Monitoring Stack

Configure end-to-end observability combining metrics, logging, alerting, and dashboards for applications and infrastructure.

## When to Use

- Setting up monitoring for a new application or service
- Adding alerting and SLOs to existing infrastructure
- Configuring logging pipelines (ELK, Loki, CloudWatch)
- Creating Grafana dashboards or CloudWatch dashboards
- Implementing distributed tracing

## Agents Activated

| Agent | Role |
|-------|------|
| `monitoring` | Prometheus, Grafana, alerting rules, SLOs/SLIs |
| `aws` / `azure` | CloudWatch, Azure Monitor, cloud-native metrics |
| `kubernetes` | K8s metrics, pod monitoring, HPA metrics |
| `networking` | Network monitoring, latency tracking |

## Workflow

### Step 1: Assess Current State

1. Identify the application stack and deployment model
2. Check for existing monitoring (Prometheus, CloudWatch, Datadog)
3. Ask the user for:
   - Monitoring stack preference (Prometheus+Grafana / CloudWatch / Datadog)
   - Logging preference (ELK / Loki / CloudWatch Logs)
   - Tracing needs (Jaeger / X-Ray / none)
   - Critical SLOs (availability %, latency targets)

### Step 2: Metrics Configuration

Using the `monitoring` agent:

1. **Application metrics**:
   - Instrument code with Prometheus client / StatsD / OTEL
   - Define custom metrics (request count, duration, errors, saturation)
   - Configure scrape targets / metric endpoints
2. **Infrastructure metrics**:
   - Node/instance metrics (CPU, memory, disk, network)
   - Container metrics (if using Docker/K8s)
   - Database metrics (connections, query time, replication lag)
3. **Golden signals**: Ensure coverage of latency, traffic, errors, saturation

### Step 3: Logging Pipeline

Using the `monitoring` agent:

1. Configure structured logging format (JSON)
2. Set up log aggregation:
   - **Loki**: Promtail config, label strategy, retention
   - **ELK**: Filebeat/Fluentd config, index patterns, ILM
   - **CloudWatch**: Log groups, metric filters, insights queries
3. Define log levels strategy and correlation IDs
4. Configure log-based alerts for critical errors

### Step 4: Define SLOs and Alerting

Using the `monitoring` agent:

1. Define SLIs (Service Level Indicators):
   - Availability: successful requests / total requests
   - Latency: p50, p95, p99 response times
   - Error rate: 5xx / total responses
2. Set SLO targets (e.g., 99.9% availability, p99 < 500ms)
3. Generate alert rules:
   - **Critical**: SLO burn rate > 10x (page immediately)
   - **Warning**: SLO burn rate > 2x (notify)
   - **Info**: Approaching thresholds
4. Configure notification channels (Slack, PagerDuty, email)

### Step 5: Dashboards

Using the `monitoring` agent:

1. Generate Grafana dashboards (JSON):
   - **Overview**: Golden signals at a glance
   - **Service detail**: Per-service deep dive
   - **Infrastructure**: Nodes, containers, databases
   - **SLO**: Error budget burn, compliance tracking
2. Or CloudWatch dashboards (if AWS-native)
3. Include drill-down links between dashboards

### Step 6: Distributed Tracing (if applicable)

1. Configure tracing instrumentation (OTEL / Jaeger / X-Ray)
2. Set sampling rate strategy
3. Link traces to logs (trace ID correlation)

### Step 7: Documentation

1. Generate observability runbook:
   - Dashboard locations and purposes
   - Alert escalation procedures
   - Common troubleshooting queries
   - SLO review cadence

## Output Files

```
observability/
├── prometheus/
│   ├── prometheus.yml           # Scrape config
│   └── rules/
│       ├── alerts.yml           # Alert rules
│       └── recording.yml        # Recording rules
├── grafana/
│   └── dashboards/
│       ├── overview.json        # Service overview
│       ├── slo.json             # SLO tracking
│       └── infrastructure.json  # Infra metrics
├── logging/
│   ├── promtail.yml             # Or filebeat.yml
│   └── loki-config.yml          # Or logstash.conf
├── docker-compose.monitoring.yml  # Local monitoring stack
└── docs/
    └── observability-runbook.md
```

## Key Principles

- **USE method**: Utilization, Saturation, Errors for resources
- **RED method**: Rate, Errors, Duration for services
- **Alert on symptoms, not causes**: Alert on user-facing impact, not CPU usage
- **SLO-based alerting**: Use error budget burn rate, not static thresholds
- **Structured logging**: Always JSON, always with correlation IDs
- **Dashboard hierarchy**: Overview → Service → Component (drill-down)
