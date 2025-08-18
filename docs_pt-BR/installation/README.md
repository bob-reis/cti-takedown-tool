# 🔧 Guia de Instalação e Configuração

Este guia detalha como instalar, configurar e colocar o CTI Takedown Tool em funcionamento em diferentes ambientes.

## 📋 Índice

- [Pré-requisitos](#pré-requisitos)
- [Instalação](#instalação)
- [Configuração](#configuração)
- [Primeira Execução](#primeira-execução)
- [Verificação](#verificação)
- [Troubleshooting](#troubleshooting)

## 📝 Pré-requisitos

### Sistema Operacional
- **Linux**: Ubuntu 20.04+, CentOS 8+, RHEL 8+
- **macOS**: 11.0+ (Big Sur)
- **Windows**: 10+ (com WSL2 recomendado)

### Software Necessário

#### Go Runtime
```bash
# Verificar versão do Go
go version
# Deve retornar: go version go1.22+ linux/amd64

# Instalar Go (se necessário)
# Ubuntu/Debian
sudo apt update
sudo apt install golang-go

# CentOS/RHEL
sudo dnf install golang

# macOS (com Homebrew)
brew install go

# Ou baixar direto: https://golang.org/dl/
```

#### Git
```bash
# Ubuntu/Debian
sudo apt install git

# CentOS/RHEL
sudo dnf install git

# macOS
xcode-select --install
```

#### SMTP Server (Produção)
- Servidor SMTP interno da empresa, ou
- Serviço externo (SendGrid, Amazon SES, etc.)

### Recursos de Sistema

| Ambiente | CPU | Memória | Disco | Rede |
|----------|-----|---------|-------|------|
| **Desenvolvimento** | 2 cores | 4GB | 10GB | 100Mbps |
| **Teste** | 2 cores | 8GB | 20GB | 500Mbps |
| **Produção** | 4+ cores | 16GB+ | 50GB+ | 1Gbps+ |

## 🚀 Instalação

### 1. Clone do Repositório

```bash
# Clone do repositório
git clone https://github.com/cti-team/takedown.git
cd takedown

# Verificar estrutura
ls -la
# Deve mostrar: cmd/, internal/, pkg/, configs/, docs/, etc.
```

### 2. Build da Aplicação

```bash
# Build da aplicação
go mod download
go build -o takedown cmd/takedown/main.go

# Verificar build
./takedown --help
```

### 3. Instalação Alternativa via Go Install

```bash
# Instalar direto via Go
go install github.com/cti-team/takedown/cmd/takedown@latest

# Verificar PATH
which takedown
takedown --help
```

### 4. Docker (Opcional)

```dockerfile
# Dockerfile
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o takedown cmd/takedown/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/takedown .
COPY --from=builder /app/configs ./configs
CMD ["./takedown", "-daemon"]
```

```bash
# Build e run Docker
docker build -t takedown:latest .
docker run -d --name takedown-daemon takedown:latest
```

## ⚙️ Configuração

### 1. Estrutura de Configuração

```
configs/
├── smtp.yaml           # Configuração SMTP
├── sla/
│   └── default.yaml    # SLAs por target type
├── routing/
│   └── rules.yaml      # Regras de roteamento
└── templates/
    ├── registrar_pt.txt # Templates de email PT
    ├── hosting_en.txt   # Templates de email EN
    └── cert_pt.txt      # Templates CERT
```

### 2. Configuração SMTP

```bash
# Criar arquivo de configuração SMTP
cat > configs/smtp.yaml << EOF
smtp:
  host: "smtp.company.com"
  port: 587
  username: "takedown@company.com"
  password: "your_secure_password"
  from: "CTI Security Team <takedown@company.com>"
  
# Configurações de segurança
security:
  tls_enabled: true
  insecure_skip_verify: false
  auth_method: "plain"
  
# Configurações de retry
retry:
  max_attempts: 3
  delay_seconds: 5
  backoff_multiplier: 2
EOF
```

### 3. Configuração de SLAs

```bash
# Personalizar SLAs por ambiente
cp configs/sla/default.yaml configs/sla/production.yaml

# Editar para produção
cat > configs/sla/production.yaml << EOF
# SLAs para ambiente de produção
registrar:
  first_response_hours: 24    # Mais agressivo em produção
  escalate_after_hours: 72    # Escalar mais rápido
  retry_interval_hours: 24
  max_retries: 5

hosting:
  first_response_hours: 12    # Hosting deve ser mais rápido
  escalate_after_hours: 48
  retry_interval_hours: 12
  max_retries: 6

cdn:
  first_response_hours: 6     # CDN é mais rápido
  escalate_after_hours: 24
  retry_interval_hours: 6
  max_retries: 8

# SLAs especiais para casos críticos
critical:
  registrar:
    first_response_hours: 6
    escalate_after_hours: 24
    retry_interval_hours: 6
  hosting:
    first_response_hours: 3
    escalate_after_hours: 12
    retry_interval_hours: 3
EOF
```

### 4. Configuração de Routing

```bash
# Personalizar regras de roteamento
cat > configs/routing/custom_rules.yaml << EOF
# Regras customizadas por organização
rules:
  # Phishing de banco - prioridade máxima
  - name: "banking_phishing"
    match: ["phishing", "brand:*bank*"]
    actions: ["registrar", "hosting", "cdn", "search", "blocklists"]
    sla_override: "critical"
    parallel: true
    
  # Ataques a governo - coordenação especial
  - name: "government_attack"
    match: ["brand:*gov*", "brand:*mil*"]
    actions: ["registrar", "hosting", "cert"]
    sla_override: "critical"
    notify_authorities: true
    
  # Malware massivo - foco em infraestrutura
  - name: "mass_malware"
    match: ["malware", "mass", "campaign"]
    actions: ["hosting", "cdn", "blocklists"]
    priority: "high"
    
  # Brand protection padrão
  - name: "brand_protection"
    match: ["brand:*"]
    actions: ["registrar"]
    sla_override: "standard"

# Configurações especiais por TLD
tld_overrides:
  ".br":
    brand_disputes: "saci_adm"
    content_abuse: "cert_coordination"
  ".gov":
    all_cases: "government_coordination"
  ".mil":
    all_cases: "government_coordination"

# Escalação automática
auto_escalation:
  enabled: true
  conditions:
    - no_response_hours: 72
    - case_age_hours: 168  # 7 dias
    - critical_priority: 24
EOF
```

### 5. Templates de Email

```bash
# Personalizar template para sua organização
cat > configs/templates/registrar_custom_pt.txt << EOF
Assunto: [URGENTE] Solicitação de suspensão - {{.Category}} em {{.Domain}}

Prezados,

A [SUA EMPRESA] identificou atividade maliciosa no domínio {{.Domain}}.
Solicitamos ação imediata conforme políticas de DNS Abuse.

📊 DADOS DO CASO:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
• ID do Caso: {{.CaseID}}
• Domínio: {{.DefangedDomain}}
• Categoria: {{.Category}}
• Severidade: {{.RiskScore}}/100
• Detectado em: {{.FirstSeen}}
• Análise: {{.Rationale}}

🎯 AÇÃO SOLICITADA:
{{.RequestedAction}}

⚡ URGÊNCIA:
Este domínio representa risco ativo para usuários finais.
Solicitamos suspensão em até 24 horas conforme boas práticas.

📞 CONTATO:
{{.ContactName}}
{{.ContactEmail}}
{{.ContactPhone}}

Favor confirmar recebimento e número de caso para acompanhamento.

Atenciosamente,
{{.OrganizationName}}
Equipe de Segurança Cibernética
EOF
```

### 6. Variáveis de Ambiente

```bash
# Criar arquivo .env para desenvolvimento
cat > .env << EOF
# Configurações de ambiente
TAKEDOWN_ENV=development
TAKEDOWN_CONFIG_DIR=./configs
TAKEDOWN_LOG_LEVEL=debug
TAKEDOWN_WORKERS=5

# SMTP (sensível - não committar)
SMTP_HOST=smtp.company.com
SMTP_PORT=587
SMTP_USER=takedown@company.com
SMTP_PASS=secure_password

# Configurações de rede
DNS_SERVERS=8.8.8.8,1.1.1.1
HTTP_TIMEOUT=30s
RDAP_TIMEOUT=10s

# Features flags
FEATURE_AUTO_ESCALATION=true
FEATURE_BULK_PROCESSING=false
FEATURE_ML_SCORING=false
EOF

# Carregar variáveis (development)
source .env
```

### 7. Configuração de Logging

```bash
# Criar diretório de logs
mkdir -p logs

# Configurar logrotate (Linux)
sudo cat > /etc/logrotate.d/takedown << EOF
/opt/takedown/logs/*.log {
    daily
    missingok
    rotate 30
    compress
    delaycompress
    notifempty
    create 0644 takedown takedown
    postrotate
        /bin/kill -HUP \`cat /var/run/takedown.pid 2> /dev/null\` 2> /dev/null || true
    endscript
}
EOF
```

## 🎯 Primeira Execução

### 1. Verificação de Configuração

```bash
# Testar configuração
./takedown -action=validate-config

# Verificar conectividade SMTP
./takedown -action=test-smtp

# Validar templates
./takedown -action=test-templates
```

### 2. Execução de Teste

```bash
# Submeter caso de teste (não enviará emails)
./takedown -action=submit \
  -ioc="test-domain.example.com" \
  -type=domain \
  -tags="test,phishing,brand:TestBank" \
  -dry-run

# Verificar processamento
./takedown -action=list
```

### 3. Primeiro Caso Real

```bash
# ⚠️ ATENÇÃO: Este comando enviará emails reais
./takedown -action=submit \
  -ioc="https://suspicious-site.com/login" \
  -tags="phishing,brand:YourBank,high"

# Acompanhar progresso
./takedown -action=status -case=<case-id>
```

### 4. Modo Daemon

```bash
# Executar como daemon para integração
./takedown -daemon -config=configs/production.yaml

# Com systemd (Linux)
sudo systemctl start takedown
sudo systemctl enable takedown
```

## ✅ Verificação

### 1. Health Checks

```bash
# Verificar saúde do sistema
./takedown -action=health

# Verificar conectividade
./takedown -action=connectivity-test

# Verificar configuração
./takedown -action=config-validate
```

### 2. Testes de Integração

```bash
# Executar suite de testes
./test.sh

# Testes específicos
go test ./internal/connectors/... -integration

# Testes de carga (futuro)
./takedown -action=load-test -cases=100
```

### 3. Monitoramento

```bash
# Logs em tempo real
tail -f logs/takedown.log

# Métricas básicas
./takedown -action=metrics

# Status de workers
./takedown -action=workers-status
```

## 🚨 Troubleshooting Rápido

### Problemas Comuns

#### 1. Falha de Build
```bash
# Erro: "go: command not found"
# Solução: Instalar Go 1.22+
curl -LO https://golang.org/dl/go1.22.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.22.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
```

#### 2. Erro de SMTP
```bash
# Erro: "dial tcp: connection refused"
# Verificar configuração SMTP
./takedown -action=test-smtp -debug

# Testar manualmente
telnet smtp.company.com 587
```

#### 3. Timeout de DNS
```bash
# Erro: "no such host"
# Verificar resolução DNS
nslookup suspicious-domain.com 8.8.8.8

# Configurar DNS alternativo
export DNS_SERVERS=1.1.1.1,8.8.8.8
```

#### 4. Permissões
```bash
# Erro: "permission denied"
# Ajustar permissões
chmod +x takedown
sudo chown -R takedown:takedown /opt/takedown/
```

### Logs de Debug

```bash
# Habilitar debug mode
export TAKEDOWN_LOG_LEVEL=debug
./takedown -action=submit -ioc="test.com" -tags="test" -debug

# Verificar logs detalhados
tail -f logs/takedown-debug.log
```

### Recuperação

```bash
# Resetar estado (desenvolvimento)
rm -rf data/state.db

# Reprocessar casos orfãos
./takedown -action=recover-orphaned

# Backup de configuração
tar -czf takedown-config-backup.tar.gz configs/
```

---

**Próximo**: [Documentação da API](../api/README.md)