# 👨‍💻 Development Guide

Instructions for setting up a development environment and contributing to the project.

## Environment Setup
- Install Go 1.22+
- Clone the repository
- Run `go mod download`

## Project Structure
Source code lives under `cmd/`, `internal/` and `pkg/`. Configuration files reside in `configs/`.

## Coding Standards
Use `golangci-lint` and `go fmt` before committing. Tests should accompany new features.

## Testing
```bash
go test ./...
./test.sh
```

## Contribution Workflow
1. Fork the repository
2. Create a feature branch
3. Commit and push changes
4. Open a pull request

Additional topics such as debugging and performance tuning are covered in the [Portuguese version](../../docs_pt-BR/development/README.md).
