# CTI Takedown Tool 🚀

[![Go Version](https://img.shields.io/badge/Go-1.22+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Tests](https://img.shields.io/badge/Tests-67%20passing-brightgreen.svg)](#testing)
[![Coverage](https://img.shields.io/badge/Coverage-85%2B%25-brightgreen.svg)](#testing)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=bob-reis_site-takedown&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=bob-reis_site-takedown)
[![Bugs](https://sonarcloud.io/api/project_badges/measure?project=bob-reis_site-takedown&metric=bugs)](https://sonarcloud.io/summary/new_code?id=bob-reis_site-takedown)
[![Duplicated Lines (%)](https://sonarcloud.io/api/project_badges/measure?project=bob-reis_site-takedown&metric=duplicated_lines_density)](https://sonarcloud.io/summary/new_code?id=bob-reis_site-takedown)

Automated system to orchestrate takedowns of malicious domains and URLs with evidence collection, SLA tracking and pluggable connectors.
[Portuguese version](README.pt-BR.md)


## ✨ Features
- Automated evidence collection (DNS, HTTP, TLS, screenshots)
- Abuse contact discovery via RDAP/WHOIS/ASN
- State machine with 9 steps
- Pluggable connectors (registrars, hosting, CDN)
- SLA tracking with automatic follow‑ups
- Customizable email templates (EN/PT)
- CLI and daemon modes

## 🏗️ Architecture
```mermaid
graph TD
    A[Frontend CTI] --> B[CLI/API]
    B --> C[State Machine]
    C --> D[Evidence Collector]
    C --> E[Enrichment Service]
    C --> F[Routing Engine]
    D --> G[DNS/HTTP/TLS Data]
    E --> H[RDAP Client]
    E --> I[ASN Lookup]
    F --> J[Connectors]
    J --> K[Registrars]
    J --> L[Hosting]
    J --> M[CDN]
    J --> N[Search/Blocklists]
```

## 🚀 Quick Start
```bash
git clone https://github.com/bob-reis/cti-takedown-tool.git
cd cti-takedown-tool
go build -o takedown cmd/takedown/main.go
./takedown --help
```

## 📚 Documentation
Detailed guides are available in the [docs](docs/README.md) folder.

- [Architecture](docs/architecture/README.md)
- [Installation](docs/installation/README.md)
- [API Reference](docs/api/README.md)
- [Development](docs/development/README.md)
- [Deployment](docs/deployment/README.md)
- [Troubleshooting](docs/troubleshooting/README.md)
- [Portuguese documentation](docs_pt-BR/README.md)

## 🧪 Testing
```bash
./test.sh
```

## 📝 License
MIT License. See [LICENSE](LICENSE).

## 🌐 Links
- Issues: https://github.com/bob-reis/cti-takedown-tool/issues
- Email: bobreis@yeslinux.com.br

*Automating cyber‑threat takedowns, one domain at a time.*
