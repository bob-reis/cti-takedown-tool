# 🚀 API and CLI Reference

This document describes the command‑line interface of the CTI Takedown Tool. A REST API is planned for future releases.

## CLI Syntax
```bash
takedown [GLOBAL_FLAGS] -action=ACTION [ACTION_FLAGS]
```

Global flags include configuration file, log level, output format, timeout and worker count.

## Main Commands
- `submit` – send an IOC for processing
- `status` – check case status
- `list` – list cases with optional filters

Each command accepts additional flags such as IOC value, tags, priority and output format. See the [Portuguese version](../../docs_pt-BR/api/README.md) for full tables and examples.

## Output Formats
Text, JSON, YAML and CSV are supported.

## Return Codes
`0` for success, non‑zero values for validation errors, connectivity issues, authentication failures and other problems.
