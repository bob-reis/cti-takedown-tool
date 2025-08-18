# 🔧 Installation and Configuration Guide

This guide explains how to install, configure and run the CTI Takedown Tool.

## Prerequisites
- Linux, macOS or Windows (WSL2 recommended)
- Go 1.22+
- Git
- Optional: SMTP server for production

## Installation
```bash
git clone https://github.com/cti-team/takedown.git
cd takedown
go mod download
go build -o takedown cmd/takedown/main.go
./takedown --help
```

### Alternative Go Install
```bash
go install github.com/cti-team/takedown/cmd/takedown@latest
```

### Docker (optional)
A sample Dockerfile is provided in the Portuguese documentation.

## Running as a Daemon
```bash
./takedown -daemon -config=configs/production.yaml
```

## Verification
- Health checks: `./takedown -action=health`
- Connectivity tests: `./takedown -action=connectivity-test`
- Configuration validation: `./takedown -action=config-validate`

## Troubleshooting
Common issues such as build failures, SMTP errors and DNS timeouts are covered in the [Portuguese version](../../docs_pt-BR/installation/README.md).
