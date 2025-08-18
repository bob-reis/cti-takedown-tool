# Takedown Module Specification — CTI Platform

## Goal
Automate and record takedown requests for malicious domains/URLs with strong evidence and audit trails while reducing MTTR.

## Architecture Overview
Components include ingestion of IOCs, enrichment (DNS, RDAP, HTTP, TLS), a policy engine, action connectors (registrars, hosting, CDN, blocklists) and an orchestration layer with auditing.

## State Flow
1. Discovered
2. Triage
3. Evidence Pack
4. Route
5. Submit
6. Acknowledge
7. Follow‑up
8. Outcome
9. Close

## Evidence Pack Checklist
- Defanged URLs
- Screenshots and HAR files with hashes
- DNS records and IP/ASN information
- HTTP headers and status
- TLS certificate details
- Reputation links and impact assessment

## Playbooks
Guidelines for contacting registrars, hosting providers, CDNs, ccTLDs, search engines, blocklists and privacy/proxy services. Emphasizes SLA timelines, escalation paths and communication templates.

For the full detailed specification with examples and templates, see the [Portuguese version](cti_takedown_spec.pt-BR.md).
