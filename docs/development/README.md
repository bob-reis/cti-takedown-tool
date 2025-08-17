# 👨‍💻 Guia de Desenvolvimento

Este guia detalha como contribuir para o desenvolvimento do CTI Takedown Tool, incluindo setup do ambiente, padrões de código, testes e processo de contribuição.

## 📋 Índice

- [Setup do Ambiente](#setup-do-ambiente)
- [Estrutura do Projeto](#estrutura-do-projeto)
- [Padrões de Código](#padrões-de-código)
- [Testing](#testing)
- [Desenvolvimento de Features](#desenvolvimento-de-features)
- [Debugging](#debugging)
- [Performance](#performance)
- [Contribuição](#contribuição)

## 🔧 Setup do Ambiente

### Pré-requisitos

```bash
# Go 1.22+
go version

# Git
git --version

# Make (opcional)
make --version

# Docker (opcional)
docker --version
```

### Clone e Setup

```bash
# Clone do repositório
git clone https://github.com/cti-team/takedown.git
cd takedown

# Setup de desenvolvimento
make dev-setup

# Ou manualmente:
go mod download
go install golang.org/x/tools/cmd/goimports@latest
go install honnef.co/go/tools/cmd/staticcheck@latest
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
```

### Configuração do Editor

#### VS Code

```json
// .vscode/settings.json
{
  "go.useLanguageServer": true,
  "go.lintTool": "golangci-lint",
  "go.lintOnSave": "package",
  "go.formatTool": "goimports",
  "go.testFlags": ["-v", "-race"],
  "go.buildFlags": ["-race"],
  "editor.formatOnSave": true,
  "go.generateTestsFlags": ["-all"]
}
```

#### VS Code Extensions

```json
// .vscode/extensions.json
{
  "recommendations": [
    "golang.Go",
    "ms-vscode.vscode-go",
    "github.vscode-pull-request-github",
    "streetsidesoftware.code-spell-checker"
  ]
}
```

### Makefile

```makefile
# Makefile
.PHONY: build test lint clean dev-setup

# Build da aplicação
build:
	go build -race -o takedown cmd/takedown/main.go

# Executar testes
test:
	go test -v -race -coverprofile=coverage.out ./...

# Executar linters
lint:
	golangci-lint run ./...
	staticcheck ./...

# Limpeza
clean:
	rm -f takedown coverage.out coverage.html

# Setup ambiente de desenvolvimento
dev-setup:
	go mod download
        go install golang.org/x/tools/cmd/goimports@latest
        go install honnef.co/go/tools/cmd/staticcheck@latest
        go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest

# Executar em desenvolvimento
dev-run: build
	./takedown -config=configs/development.yaml -log-level=debug

# Testes com coverage
test-coverage: test
	go tool cover -html=coverage.out -o coverage.html
	open coverage.html

# Executar testes continuamente
test-watch:
	find . -name "*.go" | entr -c go test -v ./...

# Benchmark
bench:
	go test -bench=. -benchmem ./...

# Build para produção
build-prod:
	CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -o takedown cmd/takedown/main.go

# Docker build
docker-build:
	docker build -t takedown:latest .

# Verificar dependências
deps-check:
	go mod verify
	go mod tidy
```

## 📂 Estrutura do Projeto

```
takedown/
├── cmd/takedown/           # CLI principal
│   └── main.go
├── internal/               # Código interno
│   ├── connectors/         # Connectors plugáveis
│   │   ├── registrar/      # Connectors de registrars
│   │   ├── hosting/        # Connectors de hosting
│   │   ├── cdn/            # Connectors de CDN
│   │   └── blocklists/     # Connectors de blocklists
│   ├── evidence/           # Coleta de evidências
│   ├── enrichment/         # Enriquecimento de dados
│   ├── routing/            # Engine de roteamento
│   └── state/              # State machine
├── pkg/                    # Código público/reutilizável
│   ├── models/             # Modelos de dados
│   └── rdap/               # Cliente RDAP
├── configs/                # Configurações
│   ├── sla/
│   ├── routing/
│   └── templates/
├── docs/                   # Documentação
├── tests/                  # Testes de integração
├── scripts/                # Scripts utilitários
└── deployments/            # Configurações de deploy
```

### Convenções de Nomenclatura

#### Packages

```go
// ✅ Bom: nomes curtos, descritivos
package evidence
package routing
package models

// ❌ Ruim: nomes longos, genéricos
package evidencecollection
package utils
package helpers
```

#### Arquivos

```
// ✅ Bom: específico e claro
collector.go          # Implementação principal
collector_test.go     # Testes
interfaces.go         # Interfaces
types.go             # Tipos específicos

// ❌ Ruim: genérico demais
main.go              # Em pkg/ (não cmd/)
helper.go
util.go
```

#### Structs e Interfaces

```go
// ✅ Bom: PascalCase, descritivo
type EvidenceCollector struct{}
type TakedownRequest struct{}
type Connector interface{}

// ❌ Ruim: nomes não descritivos
type Data struct{}
type Manager struct{}
type Handler interface{}
```

## 📝 Padrões de Código

### Go Code Style

Seguimos as convenções padrão do Go com algumas adições:

#### 1. Imports

```go
// ✅ Bom: agrupados e ordenados
import (
    // Standard library
    "context"
    "fmt"
    "time"
    
    // Third party
    "github.com/google/uuid"
    
    // Local
    "github.com/cti-team/takedown/pkg/models"
    "github.com/cti-team/takedown/internal/evidence"
)
```

#### 2. Error Handling

```go
// ✅ Bom: wrap errors com contexto
func (c *Collector) CollectEvidence(ioc *models.IOC) (*models.EvidencePack, error) {
    evidence, err := c.collectDNS(ioc.Value)
    if err != nil {
        return nil, fmt.Errorf("DNS collection failed for %s: %w", ioc.Value, err)
    }
    return evidence, nil
}

// ❌ Ruim: errors sem contexto
func (c *Collector) CollectEvidence(ioc *models.IOC) (*models.EvidencePack, error) {
    evidence, err := c.collectDNS(ioc.Value)
    if err != nil {
        return nil, err
    }
    return evidence, nil
}
```

#### 3. Context Usage

```go
// ✅ Bom: context como primeiro parâmetro
func (c *Connector) Submit(ctx context.Context, request *TakedownRequest) error {
    ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
    defer cancel()
    
    return c.httpClient.DoWithContext(ctx, req)
}
```

#### 4. Interface Design

```go
// ✅ Bom: interfaces pequenas e focadas
type Connector interface {
    Submit(ctx context.Context, request *TakedownRequest, evidence *EvidencePack) error
    CheckStatus(ctx context.Context, request *TakedownRequest) (*StatusUpdate, error)
    GetType() string
}

// ✅ Bom: accept interfaces, return structs
func ProcessWithConnector(conn Connector, req *TakedownRequest) error {
    return conn.Submit(context.Background(), req, nil)
}
```

#### 5. Struct Organization

```go
// ✅ Bom: campos agrupados logicamente
type TakedownRequest struct {
    // Identificação
    CaseID     string    `json:"case_id"`
    Status     Status    `json:"status"`
    
    // Timing
    CreatedAt    time.Time  `json:"created_at"`
    UpdatedAt    time.Time  `json:"updated_at"`
    NextActionAt *time.Time `json:"next_action_at,omitempty"`
    
    // Configuração
    Target models.TakedownTarget `json:"target"`
    SLA    models.SLA           `json:"sla"`
    
    // Dados variáveis
    History []TakedownEvent `json:"history"`
    Tags    []string        `json:"tags,omitempty"`
}
```

### Documentation

#### 1. Package Documentation

```go
// Package evidence provides secure evidence collection for IOCs.
//
// The evidence collector safely gathers DNS, HTTP, and TLS information
// from potentially malicious domains without exposing the analysis
// environment to threats.
//
// Usage:
//
//	collector := evidence.NewCollector("/tmp/evidence")
//	evidence, err := collector.CollectEvidence(ioc)
//	if err != nil {
//	    log.Fatal(err)
//	}
package evidence
```

#### 2. Function Documentation

```go
// CollectEvidence safely collects technical evidence for an IOC.
//
// This function performs DNS lookups, HTTP requests, and TLS inspection
// in an isolated environment. All IOCs are automatically defanged in
// any external communications.
//
// The collection process includes:
//   - DNS resolution (A, AAAA, CNAME, MX, TXT, NS records)
//   - HTTP response analysis (headers, body, redirects)
//   - TLS certificate inspection
//   - Risk assessment based on collected data
//
// Returns an EvidencePack containing all collected information and
// a risk assessment score from 0-100.
func (c *Collector) CollectEvidence(ioc *models.IOC) (*models.EvidencePack, error) {
```

#### 3. Complex Logic

```go
// calculateNextAction determines when the next action should occur
// based on the current status and SLA configuration.
//
// Status transitions and timing:
//   - Submitted: Wait for FirstResponseHours from creation
//   - Acked: Wait for RetryIntervalHours from last event
//   - FollowUp: Check if EscalateAfterHours exceeded, otherwise retry
//
// The NextActionAt field is set to nil for terminal states.
func (tr *TakedownRequest) calculateNextAction() {
```

## 🧪 Testing

### Estrutura de Testes

```go
// collector_test.go
package evidence

import (
    "context"
    "net/http"
    "net/http/httptest"
    "testing"
    "time"
    
    "github.com/cti-team/takedown/pkg/models"
)

func TestCollector_CollectEvidence(t *testing.T) {
    // Table-driven tests
    tests := []struct {
        name        string
        ioc         *models.IOC
        wantErr     bool
        wantMinScore int
    }{
        {
            name: "phishing URL with high risk",
            ioc: &models.IOC{
                Type:  models.IOCTypeURL,
                Value: "https://fake-bank.com/login",
                Tags:  []string{"phishing", "brand:TestBank"},
            },
            wantErr:      false,
            wantMinScore: 60,
        },
        {
            name: "invalid URL should fail",
            ioc: &models.IOC{
                Type:  models.IOCTypeURL,
                Value: "not-a-url",
            },
            wantErr: true,
        },
    }
    
    collector := NewCollector("/tmp/test")
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            evidence, err := collector.CollectEvidence(tt.ioc)
            
            if tt.wantErr {
                if err == nil {
                    t.Error("Expected error but got none")
                }
                return
            }
            
            if err != nil {
                t.Fatalf("Unexpected error: %v", err)
            }
            
            if evidence.Risk.Score < tt.wantMinScore {
                t.Errorf("Risk score %d below minimum %d", 
                    evidence.Risk.Score, tt.wantMinScore)
            }
        })
    }
}
```

### Mock Testing

```go
// HTTP Mock Server
func TestCollector_HTTPCollection(t *testing.T) {
    // Create mock server
    server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Server", "nginx/1.18.0")
        w.WriteHeader(200)
        w.Write([]byte(`<html><title>Fake Login</title><body>Login: <input type="password"></body></html>`))
    }))
    defer server.Close()
    
    collector := NewCollector("/tmp")
    ctx := context.Background()
    
    httpInfo, err := collector.collectHTTP(ctx, server.URL, "test-evidence")
    if err != nil {
        t.Fatalf("collectHTTP failed: %v", err)
    }
    
    if httpInfo.Status != 200 {
        t.Errorf("Expected status 200, got %d", httpInfo.Status)
    }
    
    if httpInfo.Title != "Fake Login" {
        t.Errorf("Expected title 'Fake Login', got %s", httpInfo.Title)
    }
}
```

### Benchmark Tests

```go
func BenchmarkDefangIOC(b *testing.B) {
    collector := NewCollector("/tmp")
    url := "https://very-long-malicious-domain.evil.com/path/to/malware"
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        collector.defangIOC(url)
    }
}

func BenchmarkRiskAssessment(b *testing.B) {
    collector := NewCollector("/tmp")
    ioc := &models.IOC{Tags: []string{"phishing", "high"}}
    evidence := &models.EvidencePack{
        HTTP: models.HTTPInfo{Status: 200, Body: "login password"},
    }
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        collector.assessRisk(ioc, evidence)
    }
}
```

### Executar Testes

```bash
# Todos os testes
make test

# Testes específicos
go test ./pkg/models/... -v

# Testes com coverage
make test-coverage

# Benchmarks
make bench

# Testes de race condition
go test -race ./...

# Testes longos
go test -timeout=5m ./...
```

## 🔨 Desenvolvimento de Features

### 1. Adicionando Novo Connector

```go
// 1. Definir struct do connector
type NewProviderConnector struct {
    httpClient *http.Client
    apiKey     string
}

// 2. Implementar interface Connector
func (n *NewProviderConnector) Submit(ctx context.Context, request *models.TakedownRequest, evidence *models.EvidencePack) error {
    // Implementação específica
    return nil
}

func (n *NewProviderConnector) CheckStatus(ctx context.Context, request *models.TakedownRequest) (*state.StatusUpdate, error) {
    // Implementação específica
    return nil, nil
}

func (n *NewProviderConnector) GetType() string {
    return "new_provider"
}

// 3. Registrar no main.go
func setupConnectors(machine *state.Machine, config *Config) {
    machine.RegisterConnector(NewNewProviderConnector(config.NewProvider))
}
```

### 2. Adicionando Nova Regra de Routing

```yaml
# configs/routing/rules.yaml
rules:
  - name: "new_threat_category"
    match: ["new_category", "special:*"]
    actions:
      - target_type: "registrar"
        action: "suspend_domain"
        priority: 1
      - target_type: "special_provider"
        action: "custom_action"
        priority: 2
    sla_override: "critical"
```

### 3. Adicionando Novo Tipo de Evidence

```go
// pkg/models/evidence.go
type NewEvidenceType struct {
    Timestamp   time.Time `json:"timestamp"`
    Source      string    `json:"source"`
    Details     string    `json:"details"`
    Confidence  float64   `json:"confidence"`
}

// Adicionar ao EvidencePack
type EvidencePack struct {
    // ... campos existentes
    NewEvidence *NewEvidenceType `json:"new_evidence,omitempty"`
}

// internal/evidence/collector.go
func (c *Collector) collectNewEvidence(ctx context.Context, ioc string) (*models.NewEvidenceType, error) {
    // Implementação
    return &models.NewEvidenceType{
        Timestamp:  time.Now().UTC(),
        Source:     "new_source",
        Details:    "collected details",
        Confidence: 0.95,
    }, nil
}
```

## 🐛 Debugging

### Logging

```go
// Usar logging estruturado
import "log/slog"

func (c *Collector) CollectEvidence(ioc *models.IOC) (*models.EvidencePack, error) {
    logger := slog.With(
        "ioc_id", ioc.IndicatorID,
        "ioc_type", ioc.Type,
        "function", "CollectEvidence",
    )
    
    logger.Info("Starting evidence collection")
    
    evidence, err := c.doCollection(ioc)
    if err != nil {
        logger.Error("Evidence collection failed", "error", err)
        return nil, err
    }
    
    logger.Info("Evidence collection completed", 
        "risk_score", evidence.Risk.Score,
        "evidence_id", evidence.EvidenceID)
    
    return evidence, nil
}
```

### Debug Mode

```bash
# Executar com debug
export TAKEDOWN_LOG_LEVEL=debug
./takedown -action=submit -ioc="test.com" -tags="test"

# Debug específico
export TAKEDOWN_DEBUG_COMPONENTS=evidence,routing
./takedown -daemon
```

### Profiling

```go
// main.go - adicionar profiling em desenvolvimento
import _ "net/http/pprof"

func main() {
    if os.Getenv("TAKEDOWN_PPROF") == "true" {
        go func() {
            log.Println(http.ListenAndServe("localhost:6060", nil))
        }()
    }
    
    // ... resto da aplicação
}
```

```bash
# Analisar CPU profile
go tool pprof http://localhost:6060/debug/pprof/profile

# Analisar memory profile
go tool pprof http://localhost:6060/debug/pprof/heap

# Analisar goroutines
go tool pprof http://localhost:6060/debug/pprof/goroutine
```

### Tracing (Futuro)

```go
// OpenTelemetry tracing
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/trace"
)

func (c *Collector) CollectEvidence(ctx context.Context, ioc *models.IOC) (*models.EvidencePack, error) {
    tracer := otel.Tracer("evidence-collector")
    ctx, span := tracer.Start(ctx, "collect_evidence")
    defer span.End()
    
    span.SetAttributes(
        attribute.String("ioc.type", string(ioc.Type)),
        attribute.String("ioc.value", ioc.Value),
    )
    
    // ... implementação
}
```

## ⚡ Performance

### Profiling e Otimização

```bash
# CPU profiling
go test -cpuprofile=cpu.prof -bench=.
go tool pprof cpu.prof

# Memory profiling
go test -memprofile=mem.prof -bench=.
go tool pprof mem.prof

# Trace analysis
go test -trace=trace.out -bench=.
go tool trace trace.out
```

### Benchmarking

```go
func BenchmarkEvidenceCollection(b *testing.B) {
    collector := NewCollector("/tmp")
    ioc := &models.IOC{
        Type:  models.IOCTypeURL,
        Value: "https://example.com",
    }
    
    b.ResetTimer()
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            _, err := collector.CollectEvidence(ioc)
            if err != nil {
                b.Fatal(err)
            }
        }
    })
}
```

### Memory Management

```go
// ✅ Bom: pool para reutilização
var bufferPool = sync.Pool{
    New: func() interface{} {
        return make([]byte, 0, 1024)
    },
}

func processData(data []byte) {
    buf := bufferPool.Get().([]byte)
    defer bufferPool.Put(buf[:0])
    
    // usar buf...
}

// ✅ Bom: context com timeout para evitar vazamentos
func (c *Collector) collectWithTimeout(ctx context.Context, url string) error {
    ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
    defer cancel()
    
    // operação que pode demorar...
}
```

## 🤝 Contribuição

### Process de Contribuição

1. **Fork** do repositório
2. **Branch** para feature: `git checkout -b feature/amazing-feature`
3. **Implementar** mudanças com testes
4. **Executar** testes: `make test`
5. **Executar** linters: `make lint`
6. **Commit** changes: `git commit -m 'Add amazing feature'`
7. **Push** branch: `git push origin feature/amazing-feature`
8. **Abrir** Pull Request

### Commit Messages

```bash
# Formato: type(scope): description
feat(connectors): add Cloudflare API connector
fix(evidence): handle timeout in DNS collection
docs(api): update CLI reference documentation
test(routing): add tests for wildcard matching
refactor(state): simplify state transition logic
```

### Pull Request Template

```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
- [ ] Added tests for new functionality
- [ ] All tests pass
- [ ] Manual testing completed

## Checklist
- [ ] Code follows style guidelines
- [ ] Self-review completed
- [ ] Documentation updated
- [ ] No new warnings
```

### Code Review Checklist

- ✅ **Functionality**: Code does what it's supposed to do
- ✅ **Tests**: Adequate test coverage
- ✅ **Error Handling**: Proper error handling and logging
- ✅ **Performance**: No obvious performance issues
- ✅ **Security**: No security vulnerabilities
- ✅ **Documentation**: Code is well documented
- ✅ **Style**: Follows project conventions

---

**Próximo**: [Documentação de Deployment](../deployment/README.md)