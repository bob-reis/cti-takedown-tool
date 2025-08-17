# 🚀 API e CLI Reference

Documentação completa da interface de linha de comando (CLI) e futura API REST do CTI Takedown Tool.

## 📋 Índice

- [CLI Reference](#cli-reference)
- [Comandos Principais](#comandos-principais)
- [Flags e Opções](#flags-e-opções)
- [Exemplos de Uso](#exemplos-de-uso)
- [API REST (Futuro)](#api-rest-futuro)
- [Códigos de Retorno](#códigos-de-retorno)
- [Formatos de Output](#formatos-de-output)

## 🖥️ CLI Reference

### Sintaxe Base

```bash
takedown [GLOBAL_FLAGS] -action=ACTION [ACTION_FLAGS]
```

### Global Flags

| Flag | Tipo | Default | Descrição |
|------|------|---------|-----------|
| `-config` | string | `config.yaml` | Arquivo de configuração |
| `-log-level` | string | `info` | Nível de log (debug, info, warn, error) |
| `-output` | string | `text` | Formato de saída (text, json, yaml) |
| `-timeout` | duration | `300s` | Timeout global para operações |
| `-workers` | int | `5` | Número de workers paralelos |
| `-dry-run` | bool | `false` | Simular ações sem executar |
| `-verbose` | bool | `false` | Output verboso |
| `-quiet` | bool | `false` | Output mínimo |

## 📋 Comandos Principais

### 1. Submit - Submeter IOC

Submete um IOC para processamento de takedown.

```bash
takedown -action=submit [FLAGS]
```

#### Flags Específicas

| Flag | Tipo | Obrigatório | Descrição | Exemplo |
|------|------|-------------|-----------|---------|
| `-ioc` | string | ✅ | IOC a ser processado | `https://evil.com/login` |
| `-type` | string | ❌ | Tipo do IOC | `url`, `domain`, `ip` |
| `-tags` | string | ❌ | Tags separadas por vírgula | `phishing,brand:MyBank,high` |
| `-priority` | string | ❌ | Prioridade do caso | `low`, `medium`, `high`, `critical` |
| `-source` | string | ❌ | Fonte da detecção | `honeypot`, `user_report`, `automation` |
| `-assignee` | string | ❌ | Responsável pelo caso | `analyst@company.com` |

#### Exemplos

```bash
# URL de phishing básica
takedown -action=submit \
  -ioc="https://fake-bank.com/login" \
  -tags="phishing,brand:MyBank"

# Domínio de malware com prioridade
takedown -action=submit \
  -ioc="malware-distribution.evil" \
  -type=domain \
  -tags="malware,campaign:APT28" \
  -priority=high

# C2 crítico com assignee
takedown -action=submit \
  -ioc="c2-server.bad.com" \
  -type=domain \
  -tags="c2,critical" \
  -priority=critical \
  -assignee="incident-response@company.com"

# Dry run para teste
takedown -action=submit \
  -ioc="test-domain.com" \
  -tags="test" \
  -dry-run
```

#### Response

```json
{
  "case_id": "tdk-f4b5c6d7-8e9f-4a5b-9c8d-7e6f5a4b3c2d",
  "status": "discovered",
  "ioc": "https://fake-bank.com/login",
  "tags": ["phishing", "brand:MyBank"],
  "priority": "medium",
  "created_at": "2024-01-15T10:30:00Z",
  "estimated_completion": "2024-01-17T10:30:00Z"
}
```

### 2. Status - Verificar Status

Consulta o status de um caso específico.

```bash
takedown -action=status -case=CASE_ID [FLAGS]
```

#### Flags Específicas

| Flag | Tipo | Obrigatório | Descrição |
|------|------|-------------|-----------|
| `-case` | string | ✅ | ID do caso |
| `-history` | bool | ❌ | Incluir histórico completo |
| `-events` | bool | ❌ | Incluir eventos detalhados |

#### Exemplos

```bash
# Status básico
takedown -action=status -case=tdk-abc-123

# Status com histórico completo
takedown -action=status -case=tdk-abc-123 -history

# Status em JSON
takedown -action=status -case=tdk-abc-123 -output=json
```

#### Response

```json
{
  "case_id": "tdk-abc-123",
  "status": "follow_up",
  "priority": "high",
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-16T14:22:00Z",
  "age_hours": 27.8,
  "target": {
    "type": "registrar",
    "entity": "GoDaddy.com, LLC",
    "email": "abuse@godaddy.com"
  },
  "sla": {
    "first_response_hours": 48,
    "time_remaining_hours": 20.2,
    "next_action_at": "2024-01-17T06:00:00Z"
  },
  "external_case_id": "GD-2024-001234",
  "is_overdue": false,
  "history": [
    {
      "timestamp": "2024-01-15T10:30:00Z",
      "event": "case_created",
      "notes": "Processing IOC: https://fake-bank.com/login"
    },
    {
      "timestamp": "2024-01-15T10:32:00Z",
      "event": "evidence_collected",
      "notes": "Evidence collected, risk score: 85"
    },
    {
      "timestamp": "2024-01-15T10:35:00Z",
      "event": "submitted",
      "channel": "email",
      "reference": "GD-2024-001234"
    }
  ]
}
```

### 3. List - Listar Casos

Lista casos com filtros opcionais.

```bash
takedown -action=list [FLAGS]
```

#### Flags Específicas

| Flag | Tipo | Default | Descrição |
|------|------|---------|-----------|
| `-status` | string | `all` | Filtrar por status |
| `-priority` | string | `all` | Filtrar por prioridade |
| `-limit` | int | `50` | Número máximo de resultados |
| `-since` | string | `24h` | Casos desde (24h, 7d, 30d) |
| `-assignee` | string | ❌ | Filtrar por responsável |
| `-tags` | string | ❌ | Filtrar por tags |

#### Exemplos

```bash
# Listar todos os casos recentes
takedown -action=list

# Casos em follow-up
takedown -action=list -status=follow_up

# Casos críticos dos últimos 7 dias
takedown -action=list -priority=critical -since=7d

# Casos de phishing em formato JSON
takedown -action=list -tags=phishing -output=json

# Casos atrasados
takedown -action=list -status=overdue -limit=20
```

#### Response

```json
{
  "total": 15,
  "cases": [
    {
      "case_id": "tdk-abc-123",
      "status": "follow_up",
      "priority": "high",
      "age_hours": 27.8,
      "target": "GoDaddy.com, LLC",
      "last_event": "email_sent",
      "is_overdue": false
    },
    {
      "case_id": "tdk-def-456",
      "status": "submitted",
      "priority": "critical",
      "age_hours": 5.2,
      "target": "Example Hosting",
      "last_event": "submitted",
      "is_overdue": false
    }
  ]
}
```

### 4. Daemon - Modo Daemon

Executa em modo daemon para integração contínua.

```bash
takedown -daemon [FLAGS]
```

#### Flags Específicas

| Flag | Tipo | Default | Descrição |
|------|------|---------|-----------|
| `-port` | int | `8080` | Porta para API REST (futuro) |
| `-pid-file` | string | `/var/run/takedown.pid` | Arquivo PID |
| `-log-file` | string | `stdout` | Arquivo de log |

#### Exemplos

```bash
# Daemon básico
takedown -daemon

# Daemon com configuração personalizada
takedown -daemon \
  -config=/etc/takedown/production.yaml \
  -log-file=/var/log/takedown/daemon.log \
  -workers=10

# Daemon com API REST (futuro)
takedown -daemon -port=8080 -workers=20
```

## 🔧 Comandos de Utilitário

### 5. Validate-Config - Validar Configuração

```bash
takedown -action=validate-config [-config=FILE]
```

### 6. Test-SMTP - Testar SMTP

```bash
takedown -action=test-smtp [-to=EMAIL]
```

### 7. Health - Health Check

```bash
takedown -action=health
```

### 8. Metrics - Métricas

```bash
takedown -action=metrics [-period=30d]
```

### 9. Export - Exportar Dados

```bash
takedown -action=export [-format=json] [-since=7d] [-output=file.json]
```

### 10. Import - Importar Dados

```bash
takedown -action=import -file=cases.json
```

## 🌐 API REST (Futuro v1.1)

### Base URL

```
https://takedown-api.company.com/api/v1
```

### Authentication

```bash
# JWT Token
curl -H "Authorization: Bearer <token>" \
  https://takedown-api.company.com/api/v1/cases
```

### Endpoints

#### POST /cases - Criar Caso

```bash
curl -X POST https://takedown-api.company.com/api/v1/cases \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "ioc": "https://malicious-site.com",
    "type": "url",
    "tags": ["phishing", "brand:MyBank"],
    "priority": "high",
    "source": "automation"
  }'
```

#### GET /cases - Listar Casos

```bash
curl "https://takedown-api.company.com/api/v1/cases?status=follow_up&limit=10" \
  -H "Authorization: Bearer <token>"
```

#### GET /cases/{id} - Detalhes do Caso

```bash
curl "https://takedown-api.company.com/api/v1/cases/tdk-abc-123" \
  -H "Authorization: Bearer <token>"
```

#### PATCH /cases/{id} - Atualizar Caso

```bash
curl -X PATCH "https://takedown-api.company.com/api/v1/cases/tdk-abc-123" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "priority": "critical",
    "assignee": "analyst@company.com"
  }'
```

#### GET /metrics - Métricas

```bash
curl "https://takedown-api.company.com/api/v1/metrics?period=30d" \
  -H "Authorization: Bearer <token>"
```

### WebHooks (Futuro)

```bash
# Configurar webhook para eventos
curl -X POST https://takedown-api.company.com/api/v1/webhooks \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "url": "https://your-app.com/webhook/takedown",
    "events": ["case_created", "case_resolved", "case_escalated"],
    "secret": "webhook_secret"
  }'
```

#### Webhook Payload

```json
{
  "event": "case_resolved",
  "timestamp": "2024-01-17T10:30:00Z",
  "case": {
    "case_id": "tdk-abc-123",
    "status": "closed",
    "resolution": "domain_suspended",
    "resolution_time_hours": 36.5
  }
}
```

## 📊 Códigos de Retorno

| Código | Descrição | Exemplo |
|--------|-----------|---------|
| `0` | Sucesso | Comando executado com sucesso |
| `1` | Erro geral | Argumentos inválidos, configuração incorreta |
| `2` | Erro de validação | IOC inválido, tags malformadas |
| `3` | Erro de conectividade | SMTP inacessível, DNS timeout |
| `4` | Erro de autenticação | Credenciais inválidas |
| `5` | Recurso não encontrado | Case ID inexistente |
| `6` | Timeout | Operação excedeu tempo limite |
| `7` | Erro de configuração | Arquivo de config inválido |

### Verificação de Exit Code

```bash
# Bash
./takedown -action=submit -ioc="test.com" -tags="test"
if [ $? -eq 0 ]; then
    echo "Sucesso"
else
    echo "Erro: $?"
fi

# Python
import subprocess
result = subprocess.run(['./takedown', '-action=list'], capture_output=True)
if result.returncode == 0:
    print("Success:", result.stdout.decode())
else:
    print("Error:", result.stderr.decode())
```

## 📄 Formatos de Output

### Text (Default)

```
Case ID: tdk-abc-123
Status: follow_up
Priority: high
Age: 27.8 hours
Target: GoDaddy.com, LLC
Next Action: 2024-01-17 06:00:00 UTC
```

### JSON

```json
{
  "case_id": "tdk-abc-123",
  "status": "follow_up",
  "priority": "high",
  "age_hours": 27.8,
  "target": "GoDaddy.com, LLC",
  "next_action_at": "2024-01-17T06:00:00Z"
}
```

### YAML

```yaml
case_id: tdk-abc-123
status: follow_up
priority: high
age_hours: 27.8
target: GoDaddy.com, LLC
next_action_at: "2024-01-17T06:00:00Z"
```

### CSV (para exports)

```csv
case_id,status,priority,age_hours,target,created_at
tdk-abc-123,follow_up,high,27.8,GoDaddy.com LLC,2024-01-15T10:30:00Z
tdk-def-456,submitted,critical,5.2,Example Hosting,2024-01-16T20:15:00Z
```

## 🔍 Filtros Avançados

### Sintaxe de Filtros

```bash
# Operadores de comparação
-status=submitted              # Igualdade
-age-gt=24h                   # Maior que
-age-lt=7d                    # Menor que
-priority-in=high,critical    # Lista de valores

# Operadores de texto
-target-contains=godaddy      # Contém texto
-tags-any=phishing,malware    # Qualquer das tags
-tags-all=phishing,brand:*    # Todas as tags (com wildcard)

# Operadores de data
-created-since=2024-01-01     # Desde data
-created-before=2024-01-31    # Antes de data
-updated-last=24h             # Última atualização
```

### Exemplos Avançados

```bash
# Casos atrasados de alta prioridade
takedown -action=list \
  -status=follow_up \
  -priority-in=high,critical \
  -age-gt=48h

# Casos de phishing dos últimos 30 dias
takedown -action=list \
  -tags-any=phishing \
  -created-since=30d \
  -output=csv

# Casos do GoDaddy que precisam follow-up
takedown -action=list \
  -status=follow_up \
  -target-contains=godaddy \
  -next-action-overdue=true
```

---

**Próximo**: [Guia de Desenvolvimento](../development/README.md)