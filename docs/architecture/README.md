# 🏗️ System Architecture

This document outlines the overall design of the CTI Takedown Tool.

## Overview
The platform is built around an event‑driven state machine with loosely coupled components and pluggable connectors. The design enables scalability, maintainability and thorough testing.

## Main Components
- **Ingest** – receives IOCs from crawlers, internal detections and external feeds.
- **Enrichment** – collects DNS, RDAP/WHOIS, HTTP, TLS and screenshot evidence.
- **Policy Engine** – scores cases and determines severity.
- **Connectors** – interact with registrars, hosting providers, CDNs and blocklists.
- **Orchestration** – tracks state transitions and SLA timers.
- **Auditing** – stores evidence, communications and outcomes.

## Data Flow
1. IOC is ingested.
2. Enrichment gathers technical evidence.
3. Policy engine decides actions.
4. Connectors submit requests to the appropriate targets.
5. Follow‑ups continue until closure.

For advanced topics such as architectural patterns, scalability and security, see the [Portuguese version](../../docs_pt-BR/architecture/README.md).
