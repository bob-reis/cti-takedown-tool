# Changelog

All notable changes to this project will be documented in this file.
The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]
### Planned for v1.1
- Full REST API with JWT auth
- Web dashboard for monitoring
- Integration with MISP/OpenCTI
- Advanced metrics with Prometheus/Grafana
- Webhook notifications
- ML-based risk scoring

### Planned for v1.2
- More threat intelligence feeds
- Full automation for simple cases
- Mobile app for approvals
- Multi-tenancy support
- Advanced analytics and reporting

## [1.0.0] - 2024-01-15
### Added
- Complete state machine with 9 states
- Automatic evidence collection (DNS, HTTP, TLS)
- Configurable routing engine
- RDAP client for abuse contacts
- Customizable email templates (PT/EN)
- SLA tracking with automatic follow-ups
- CLI for all operations
- Connectors for GoDaddy, Registro.br and generic registrars
- Hosting provider support via ASN lookup
- Pluggable framework for new connectors
- Security features: IOC defang, isolated evidence collection, input validation, auditing
- Flexible YAML configuration with validation
- 67 unit tests with 85%+ coverage
- Detailed documentation (architecture, installation, API, development, deployment, troubleshooting)

### Initial Configuration
- Configurable parallel workers (default 5)
- Adjustable timeouts
- Log levels and proxy support
- Rate limiting

### Metrics & Monitoring
- Health checks
- Structured logs
- Basic performance metrics
- SLA compliance tracking

### Integrations
- RDAP for .com/.net/.org/.br
- Configurable DNS resolvers
- SMTP for emails

### Performance
- Parallel case processing
- RDAP result caching

### Compliance
- ICANN DNS Abuse Policy
- GDPR
- Brazilian .br process

## [0.9.0] - 2024-01-10 (Release Candidate)
### Added
- Initial state machine
- Evidence collector
- RDAP client
- Basic email templates
- CLI commands

### Changes
- Refactored architecture for pluggable connectors
- Configuration improvements
- Performance optimizations

### Fixed
- RDAP vCard parsing issues
- Evidence collector memory leaks
- State machine race conditions

## [0.8.0] - 2024-01-05 (Beta)
### Added
- Proof of concept
- Basic models and routing engine
- GoDaddy connector

### Changes
- Migrated from Python to Go
- New state machine architecture

## [0.1.0] - 2024-01-01 (Alpha)
### Added
- Initial project structure
- Technical specification
- Go module setup

---
For release notes in Portuguese, see [CHANGELOG.pt-BR.md](CHANGELOG.pt-BR.md).
