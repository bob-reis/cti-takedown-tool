# 🔍 Troubleshooting Guide

Common issues and diagnostic tips for the CTI Takedown Tool.

## Build Problems
- `go: command not found` – install Go 1.22+
- missing dependencies – run `go mod download`

## SMTP Errors
- `connection refused` – verify SMTP settings and network access

## DNS Timeouts
- check resolver configuration and try alternative DNS servers

## Permissions
Ensure the binary has execute permission and the process has access to required files.

More detailed troubleshooting, including debug logs and recovery steps, can be found in the [Portuguese version](../../docs_pt-BR/troubleshooting/README.md).
