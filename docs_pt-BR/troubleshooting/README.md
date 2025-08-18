# 🔍 Troubleshooting Guide

Este guia abrangente ajuda a diagnosticar e resolver problemas comuns no CTI Takedown Tool, incluindo debugging avançado, logs de sistema e procedimentos de recuperação.

## 📋 Índice

- [Problemas Comuns](#problemas-comuns)
- [Diagnóstico de Sistema](#diagnóstico-de-sistema)
- [Análise de Logs](#análise-de-logs)
- [Problemas de Conectividade](#problemas-de-conectividade)
- [Problemas de Performance](#problemas-de-performance)
- [Recovery Procedures](#recovery-procedures)
- [Debug Avançado](#debug-avançado)

## 🚨 Problemas Comuns

### 1. **Falha no Build**

#### Erro: "go: command not found"
```bash
# Problema: Go não instalado ou não no PATH

# Solução 1: Instalar Go
curl -LO https://golang.org/dl/go1.22.1.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.22.1.linux-amd64.tar.gz

# Adicionar ao PATH
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# Verificar instalação
go version
```

#### Erro: "package github.com/google/uuid: cannot find module"
```bash
# Problema: Dependências não baixadas

# Solução: Baixar dependências
go mod download
go mod tidy

# Se problema persistir, limpar cache
go clean -modcache
go mod download
```

#### Erro: "build constraints exclude all Go files"
```bash
# Problema: Tags de build ou arquitetura incompatível

# Verificar arquitetura
go env GOOS GOARCH

# Build explícito
GOOS=linux GOARCH=amd64 go build -o takedown cmd/takedown/main.go
```

### 2. **Problemas de Configuração**

#### Erro: "config file not found"
```bash
# Verificar localização do arquivo
ls -la configs/
ls -la /opt/takedown/configs/

# Criar config básico se ausente
mkdir -p configs
cat > configs/config.yaml << EOF
smtp:
  host: "localhost"
  port: 587
  from: "test@localhost"
workers: 5
log_level: "info"
EOF
```

#### Erro: "invalid YAML configuration"
```bash
# Validar sintaxe YAML
python3 -c "import yaml; yaml.safe_load(open('configs/config.yaml'))"

# Ou usar yamllint
sudo apt install yamllint
yamllint configs/config.yaml

# Validar com a aplicação
./takedown -action=validate-config -config=configs/config.yaml
```

### 3. **Problemas de SMTP**

#### Erro: "dial tcp: connection refused"
```bash
# Diagnosticar conectividade
telnet smtp.company.com 587
nc -zv smtp.company.com 587

# Verificar configuração de proxy
echo $http_proxy $https_proxy

# Testar com debug
./takedown -action=test-smtp -debug

# Verificar logs do sistema
journalctl -u postfix -f  # se usando Postfix local
```

#### Erro: "535 authentication failed"
```bash
# Verificar credenciais
echo "AUTH PLAIN $(echo -ne '\0username\0password' | base64)" | \
nc smtp.company.com 587

# Testar com openssl
openssl s_client -connect smtp.company.com:587 -starttls smtp

# Verificar se 2FA está habilitado
# Para Gmail: usar app-specific password
```

#### Erro: "certificate verify failed"
```bash
# Desabilitar verificação SSL temporariamente (só para debug)
export TAKEDOWN_SMTP_INSECURE=true
./takedown -action=test-smtp

# Verificar certificados
openssl s_client -connect smtp.company.com:587 -starttls smtp -showcerts

# Atualizar certificados do sistema
sudo apt update && sudo apt install ca-certificates
```

### 4. **Problemas de DNS/Network**

#### Erro: "no such host"
```bash
# Verificar resolução DNS
nslookup suspicious-domain.com
dig suspicious-domain.com @8.8.8.8

# Testar com diferentes DNS
export DNS_SERVERS=1.1.1.1,8.8.8.8
./takedown -action=submit -ioc="test.com" -tags="test"

# Verificar /etc/resolv.conf
cat /etc/resolv.conf

# Limpar cache DNS local
sudo systemctl restart systemd-resolved
```

#### Erro: "i/o timeout"
```bash
# Aumentar timeout
export HTTP_TIMEOUT=60s
export DNS_TIMEOUT=30s

# Verificar MTU
ip link show | grep mtu

# Testar conectividade básica
ping -c 4 8.8.8.8
curl -I http://example.com
```

### 5. **Problemas de Permissões**

#### Erro: "permission denied"
```bash
# Verificar proprietário e permissões
ls -la takedown
stat takedown

# Ajustar permissões
chmod +x takedown
sudo chown takedown:takedown /opt/takedown -R

# Verificar SELinux (se aplicável)
getenforce
sudo setsebool -P httpd_can_network_connect 1
```

#### Erro: "cannot write to log file"
```bash
# Criar diretório de logs
sudo mkdir -p /opt/takedown/logs
sudo chown takedown:takedown /opt/takedown/logs

# Verificar espaço em disco
df -h /opt/takedown

# Verificar inodes
df -i /opt/takedown
```

## 🔍 Diagnóstico de Sistema

### 1. **Health Check Completo**

```bash
#!/bin/bash
# /opt/takedown/scripts/diagnostic.sh

echo "=== CTI Takedown Tool Diagnostic ==="
echo "Date: $(date)"
echo

# 1. Verificar binário
echo "1. Binary Check:"
if [ -f "./takedown" ]; then
    echo "✓ Binary exists"
    echo "  Version: $(./takedown --version 2>/dev/null || echo 'Unable to get version')"
    echo "  Size: $(ls -lh ./takedown | awk '{print $5}')"
    echo "  Permissions: $(ls -la ./takedown | awk '{print $1, $3, $4}')"
else
    echo "✗ Binary not found"
fi
echo

# 2. Verificar configuração
echo "2. Configuration Check:"
if [ -f "configs/config.yaml" ]; then
    echo "✓ Config file exists"
    if ./takedown -action=validate-config > /dev/null 2>&1; then
        echo "✓ Config is valid"
    else
        echo "✗ Config validation failed"
        ./takedown -action=validate-config
    fi
else
    echo "✗ Config file not found"
fi
echo

# 3. Verificar conectividade
echo "3. Connectivity Check:"

# DNS
if nslookup google.com > /dev/null 2>&1; then
    echo "✓ DNS resolution working"
else
    echo "✗ DNS resolution failed"
fi

# HTTP
if curl -s --max-time 10 http://example.com > /dev/null; then
    echo "✓ HTTP connectivity working"
else
    echo "✗ HTTP connectivity failed"
fi

# SMTP (se configurado)
SMTP_HOST=$(grep "host:" configs/smtp.yaml 2>/dev/null | awk '{print $2}' | tr -d '"')
SMTP_PORT=$(grep "port:" configs/smtp.yaml 2>/dev/null | awk '{print $2}')

if [ -n "$SMTP_HOST" ] && [ -n "$SMTP_PORT" ]; then
    if nc -z "$SMTP_HOST" "$SMTP_PORT" 2>/dev/null; then
        echo "✓ SMTP connectivity working ($SMTP_HOST:$SMTP_PORT)"
    else
        echo "✗ SMTP connectivity failed ($SMTP_HOST:$SMTP_PORT)"
    fi
else
    echo "- SMTP not configured"
fi
echo

# 4. Verificar recursos do sistema
echo "4. System Resources:"
echo "  CPU cores: $(nproc)"
echo "  Memory: $(free -h | grep '^Mem:' | awk '{print $2 " total, " $7 " available"}')"
echo "  Disk space: $(df -h . | tail -1 | awk '{print $4 " available"}')"
echo "  Load average: $(uptime | awk -F'load average:' '{print $2}')"
echo

# 5. Verificar processos relacionados
echo "5. Process Check:"
if pgrep -f "takedown" > /dev/null; then
    echo "✓ Takedown process running"
    echo "  PIDs: $(pgrep -f 'takedown' | tr '\n' ' ')"
    echo "  Memory usage: $(ps -o pid,pmem,rss,cmd -C takedown --no-headers 2>/dev/null || echo 'Unable to get memory info')"
else
    echo "- No takedown process running"
fi
echo

# 6. Verificar logs recentes
echo "6. Recent Logs:"
if [ -f "logs/takedown.log" ]; then
    echo "✓ Log file exists"
    echo "  Size: $(ls -lh logs/takedown.log | awk '{print $5}')"
    echo "  Last 3 entries:"
    tail -n 3 logs/takedown.log 2>/dev/null | sed 's/^/    /'
else
    echo "- No log file found"
fi

echo
echo "=== End Diagnostic ==="
```

### 2. **Performance Analysis**

```bash
#!/bin/bash
# Performance monitoring script

echo "=== Performance Analysis ==="

# CPU usage
echo "CPU Usage:"
top -bn1 | grep "takedown" | awk '{print "  PID " $1 ": " $9 "% CPU"}'

# Memory usage
echo "Memory Usage:"
ps -o pid,pmem,rss,vsz,cmd -C takedown --no-headers 2>/dev/null | \
while read pid pmem rss vsz cmd; do
    echo "  PID $pid: ${pmem}% memory, ${rss}KB RSS, ${vsz}KB VSZ"
done

# File descriptors
echo "File Descriptors:"
for pid in $(pgrep takedown); do
    fd_count=$(ls /proc/$pid/fd 2>/dev/null | wc -l)
    echo "  PID $pid: $fd_count open file descriptors"
done

# Network connections
echo "Network Connections:"
netstat -tulpn 2>/dev/null | grep takedown | while read line; do
    echo "  $line"
done

# Disk I/O
echo "Disk I/O:"
if command -v iotop >/dev/null 2>&1; then
    iotop -p $(pgrep takedown | tr '\n' ',' | sed 's/,$//') -n 1 -q 2>/dev/null | tail -n +3
else
    echo "  iotop not available"
fi
```

## 📋 Análise de Logs

### 1. **Estrutura de Logs**

```bash
# Localização dos logs
/opt/takedown/logs/
├── takedown.log          # Log principal da aplicação
├── access.log           # Log de acesso (futuro)
├── error.log            # Log de erros críticos
├── smtp.log             # Log específico de SMTP
└── debug.log            # Log de debug (quando habilitado)
```

### 2. **Análise de Logs Comuns**

```bash
# Verificar erros críticos
grep -i "error\|fatal\|panic" /opt/takedown/logs/takedown.log | tail -20

# Casos que falharam
grep "status.*failed" /opt/takedown/logs/takedown.log

# Problemas de SMTP
grep -i "smtp\|mail" /opt/takedown/logs/takedown.log | grep -i "error\|fail"

# Timeouts
grep -i "timeout\|deadline" /opt/takedown/logs/takedown.log

# Estatísticas por hora
awk '{print $1 " " $2}' /opt/takedown/logs/takedown.log | cut -c1-13 | sort | uniq -c

# Top errors
grep -i "error" /opt/takedown/logs/takedown.log | awk '{print $NF}' | sort | uniq -c | sort -nr

# Performance analysis
grep "evidence.*completed" /opt/takedown/logs/takedown.log | \
awk '{print $NF}' | sed 's/ms//' | \
awk '{sum+=$1; count++} END {printf "Average evidence collection: %.2fms\n", sum/count}'
```

### 3. **Log Analysis Scripts**

```bash
#!/bin/bash
# /opt/takedown/scripts/log_analysis.sh

LOG_FILE="/opt/takedown/logs/takedown.log"
PERIOD=${1:-"1h"}  # Default: last hour

echo "=== Log Analysis for last $PERIOD ==="

# Convert period to minutes
case $PERIOD in
    *h) MINUTES=$((${PERIOD%h} * 60)) ;;
    *m) MINUTES=${PERIOD%m} ;;
    *d) MINUTES=$((${PERIOD%d} * 1440)) ;;
    *) MINUTES=60 ;;
esac

# Get logs from specified period
SINCE=$(date -d "$MINUTES minutes ago" '+%Y-%m-%d %H:%M')
awk -v since="$SINCE" '$0 >= since' "$LOG_FILE" > /tmp/recent_logs.txt

echo "Total log entries: $(wc -l < /tmp/recent_logs.txt)"
echo

# Error analysis
echo "Error Summary:"
grep -i "error\|fail" /tmp/recent_logs.txt | \
awk -F: '{print $NF}' | sort | uniq -c | sort -nr | head -10
echo

# Case status summary
echo "Case Status Summary:"
grep "status.*updated" /tmp/recent_logs.txt | \
awk '{print $(NF-1)}' | sort | uniq -c
echo

# Target analysis
echo "Target Distribution:"
grep "target.*type" /tmp/recent_logs.txt | \
awk '{print $NF}' | sort | uniq -c
echo

# Performance metrics
echo "Performance Metrics:"
grep "duration" /tmp/recent_logs.txt | \
awk '{print $NF}' | sed 's/[^0-9.]//g' | \
awk '{sum+=$1; count++; if($1>max||max=="") max=$1; if($1<min||min=="") min=$1} 
     END {printf "  Average: %.2fs\n  Min: %.2fs\n  Max: %.2fs\n", sum/count, min, max}'

rm /tmp/recent_logs.txt
```

## 🌐 Problemas de Conectividade

### 1. **Debugging de DNS**

```bash
# Teste completo de DNS
function debug_dns() {
    local domain=$1
    
    echo "=== DNS Debug for $domain ==="
    
    # Resolução básica
    echo "1. Basic resolution:"
    nslookup "$domain"
    
    # Teste com diferentes DNS servers
    echo "2. Different DNS servers:"
    for dns in 8.8.8.8 1.1.1.1 208.67.222.222; do
        echo "  Using $dns:"
        dig @"$dns" "$domain" +short
    done
    
    # Verificar tipos de record
    echo "3. Record types:"
    for type in A AAAA CNAME MX TXT NS; do
        result=$(dig @8.8.8.8 "$domain" "$type" +short)
        if [ -n "$result" ]; then
            echo "  $type: $result"
        fi
    done
    
    # Trace DNS path
    echo "4. DNS trace:"
    dig +trace "$domain" | tail -10
}

# Uso: debug_dns suspicious-domain.com
```

### 2. **Debugging de HTTP**

```bash
# Teste completo de HTTP
function debug_http() {
    local url=$1
    
    echo "=== HTTP Debug for $url ==="
    
    # Teste básico de conectividade
    echo "1. Basic connectivity:"
    curl -I --max-time 10 "$url"
    
    # Teste com diferentes User-Agents
    echo "2. Different User-Agents:"
    for ua in "Mozilla/5.0" "curl/7.0" "CTI-Takedown/1.0"; do
        echo "  Using '$ua':"
        curl -s -o /dev/null -w "  Status: %{http_code}, Time: %{time_total}s\n" \
             -H "User-Agent: $ua" --max-time 10 "$url"
    done
    
    # Verificar certificado SSL
    echo "3. SSL Certificate:"
    echo | openssl s_client -connect "${url#https://}:443" -servername "${url#https://}" 2>/dev/null | \
    openssl x509 -noout -subject -dates -issuer 2>/dev/null
    
    # Teste de redirecionamentos
    echo "4. Redirects:"
    curl -L -s -o /dev/null -w "URL: %{url_effective}\nRedirect count: %{num_redirects}\n" "$url"
}

# Uso: debug_http https://suspicious-site.com
```

### 3. **Proxy e Firewall Issues**

```bash
# Verificar configurações de proxy
echo "Proxy Configuration:"
env | grep -i proxy

# Testar conectividade com proxy
if [ -n "$http_proxy" ]; then
    echo "Testing with proxy:"
    curl --proxy "$http_proxy" -I http://example.com
fi

# Verificar regras de firewall
echo "Firewall Rules:"
sudo iptables -L OUTPUT -n | grep -E "(80|443|25|587)"

# Verificar portas abertas
echo "Open Ports:"
netstat -tulpn | grep -E ":(80|443|25|587|8080)\s"

# Teste de conectividade para diferentes portas
for port in 80 443 25 587; do
    echo "Testing port $port:"
    nc -zv google.com $port 2>&1 | head -1
done
```

## ⚡ Problemas de Performance

### 1. **High Memory Usage**

```bash
# Análise de uso de memória
echo "=== Memory Analysis ==="

# Uso geral do sistema
free -h

# Uso específico do takedown
ps -o pid,pmem,rss,vsz,cmd -C takedown --no-headers

# Memory leaks detection
echo "Monitoring memory for 60 seconds..."
for i in {1..12}; do
    rss=$(ps -o rss= -C takedown 2>/dev/null | awk '{sum+=$1} END {print sum+0}')
    echo "$(date '+%H:%M:%S'): ${rss}KB"
    sleep 5
done

# Verificar swap usage
echo "Swap usage:"
swapon --show
```

**Soluções para High Memory:**
```bash
# 1. Reduzir número de workers
export TAKEDOWN_WORKERS=3

# 2. Implementar memory limits no systemd
sudo systemctl edit takedown.service
# Adicionar:
[Service]
MemoryLimit=512M
MemoryAccounting=yes

# 3. Configurar memory profiling
export TAKEDOWN_PPROF=true
# Analisar com: go tool pprof http://localhost:6060/debug/pprof/heap
```

### 2. **High CPU Usage**

```bash
# Análise de CPU
echo "=== CPU Analysis ==="

# CPU usage atual
top -bn1 | grep takedown

# CPU usage histórico
sar -u 1 10  # 10 samples, 1 second apart

# Verificar goroutines (se pprof habilitado)
curl -s http://localhost:6060/debug/pprof/goroutine?debug=1 | head -20

# Process tree
pstree -p $(pgrep takedown) 2>/dev/null
```

**Soluções para High CPU:**
```bash
# 1. Reduzir workers e batch size
export TAKEDOWN_WORKERS=2
export BATCH_SIZE=10

# 2. Implementar CPU limits
sudo systemctl edit takedown.service
# Adicionar:
[Service]
CPUQuota=50%

# 3. Verificar infinite loops
strace -p $(pgrep takedown) -c  # Count system calls
```

### 3. **Slow Response Times**

```bash
# Benchmark de performance
function benchmark_takedown() {
    echo "=== Performance Benchmark ==="
    
    # Test evidence collection
    time ./takedown -action=submit -ioc="test.example.com" -tags="test" -dry-run
    
    # Test status queries
    time for i in {1..10}; do
        ./takedown -action=list -limit=1 >/dev/null
    done
    
    # Test config validation
    time ./takedown -action=validate-config
}

# Network latency test
function test_network_latency() {
    echo "=== Network Latency Test ==="
    
    # DNS latency
    time nslookup google.com >/dev/null
    
    # HTTP latency
    curl -w "DNS: %{time_namelookup}s, Connect: %{time_connect}s, Total: %{time_total}s\n" \
         -o /dev/null -s http://example.com
    
    # SMTP latency
    time nc -z smtp.company.com 587
}
```

## 🔧 Recovery Procedures

### 1. **Service Recovery**

```bash
#!/bin/bash
# Emergency recovery script

echo "=== Emergency Recovery Procedure ==="

# 1. Stop service
echo "1. Stopping service..."
sudo systemctl stop takedown 2>/dev/null || true
pkill -f takedown 2>/dev/null || true
sleep 5

# 2. Backup current state
echo "2. Creating emergency backup..."
BACKUP_DIR="/opt/takedown/emergency_backup_$(date +%Y%m%d_%H%M%S)"
mkdir -p "$BACKUP_DIR"
cp -r /opt/takedown/data "$BACKUP_DIR/" 2>/dev/null || true
cp -r /opt/takedown/logs "$BACKUP_DIR/" 2>/dev/null || true

# 3. Check for corrupted files
echo "3. Checking for corruption..."
if [ -f "/opt/takedown/data/state.db" ]; then
    file /opt/takedown/data/state.db
    # Se necessário, restaurar backup
    # cp /opt/takedown/backups/latest/data/state.db /opt/takedown/data/
fi

# 4. Reset permissions
echo "4. Resetting permissions..."
sudo chown -R takedown:takedown /opt/takedown/
chmod +x /opt/takedown/bin/takedown

# 5. Validate configuration
echo "5. Validating configuration..."
if ! /opt/takedown/bin/takedown -action=validate-config; then
    echo "Configuration invalid, restoring from backup..."
    cp /opt/takedown/backups/latest/configs/* /opt/takedown/configs/
fi

# 6. Start service
echo "6. Starting service..."
sudo systemctl start takedown

# 7. Verify health
echo "7. Verifying health..."
sleep 10
if curl -s http://localhost:8080/health >/dev/null; then
    echo "✓ Service recovered successfully"
else
    echo "✗ Recovery failed, manual intervention required"
    exit 1
fi

echo "Recovery completed. Backup created at: $BACKUP_DIR"
```

### 2. **Data Recovery**

```bash
#!/bin/bash
# Data recovery procedures

BACKUP_DIR="/opt/takedown/backups"
DATA_DIR="/opt/takedown/data"

echo "=== Data Recovery ==="

# List available backups
echo "Available backups:"
ls -la "$BACKUP_DIR"/*.tar.gz | tail -10

# Function to restore from backup
restore_backup() {
    local backup_file=$1
    
    if [ ! -f "$backup_file" ]; then
        echo "Backup file not found: $backup_file"
        return 1
    fi
    
    echo "Restoring from: $backup_file"
    
    # Stop service
    sudo systemctl stop takedown
    
    # Backup current data
    mv "$DATA_DIR" "${DATA_DIR}.corrupted.$(date +%Y%m%d_%H%M%S)"
    
    # Extract backup
    tar -xzf "$backup_file" -C /opt/takedown/
    
    # Fix permissions
    sudo chown -R takedown:takedown "$DATA_DIR"
    
    # Start service
    sudo systemctl start takedown
    
    echo "Restore completed"
}

# Usage: restore_backup /opt/takedown/backups/data_20240115_120000.tar.gz
```

### 3. **Configuration Recovery**

```bash
# Generate default configuration
function generate_default_config() {
    local config_file="/opt/takedown/configs/emergency.yaml"
    
    cat > "$config_file" << EOF
# Emergency configuration
server:
  workers: 1
  queue_size: 100
  timeout: 60s

logging:
  level: "debug"
  file: "/opt/takedown/logs/emergency.log"

smtp:
  host: "localhost"
  port: 25
  from: "emergency@localhost"

features:
  auto_escalation: false
  bulk_processing: false
EOF

    echo "Emergency config created: $config_file"
}

# Test minimal configuration
function test_minimal_config() {
    echo "Testing with minimal configuration..."
    
    ./takedown -config=/opt/takedown/configs/emergency.yaml \
               -action=validate-config
    
    if [ $? -eq 0 ]; then
        echo "✓ Minimal config works"
        return 0
    else
        echo "✗ Even minimal config fails"
        return 1
    fi
}
```

## 🐛 Debug Avançado

### 1. **Profiling com pprof**

```bash
# Habilitar profiling
export TAKEDOWN_PPROF=true
./takedown -daemon &

# CPU profiling
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# Memory profiling
go tool pprof http://localhost:6060/debug/pprof/heap

# Goroutine analysis
go tool pprof http://localhost:6060/debug/pprof/goroutine

# Comandos úteis no pprof:
# top10          - Top 10 funções por CPU/memory
# list function  - Source code de função específica
# web           - Visualização em browser
# exit          - Sair
```

### 2. **Trace Analysis**

```bash
# Coletar trace
curl http://localhost:6060/debug/pprof/trace?seconds=10 > trace.out

# Analisar trace
go tool trace trace.out

# Comandos no trace viewer:
# Goroutine analysis - Ver goroutines
# Network blocking   - I/O blocking
# Syscall blocking   - System calls
# Scheduler latency  - Go scheduler
```

### 3. **Debugging com Delve**

```bash
# Instalar Delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Debug da aplicação
dlv exec ./takedown -- -action=submit -ioc="test.com" -tags="test"

# Comandos úteis no Delve:
# break main.main     - Breakpoint na função main
# continue           - Continuar execução
# next              - Próxima linha
# step              - Step into função
# print variable    - Imprimir variável
# stack             - Stack trace
# goroutines        - Listar goroutines
# quit              - Sair
```

### 4. **System Call Tracing**

```bash
# Trace system calls
strace -p $(pgrep takedown) -f -e trace=network,file -o strace.log

# Analisar network calls
grep -E "(socket|connect|bind|listen)" strace.log

# Analisar file operations
grep -E "(open|read|write|close)" strace.log

# Verificar performance de syscalls
strace -p $(pgrep takedown) -c  # Count calls
```

### 5. **Logs Estruturados**

```bash
# Configurar logging estruturado para debug
export TAKEDOWN_LOG_FORMAT=json
export TAKEDOWN_LOG_LEVEL=debug

# Analisar logs com jq
tail -f /opt/takedown/logs/takedown.log | jq '.'

# Filtrar por nível de log
jq 'select(.level == "error")' /opt/takedown/logs/takedown.log

# Filtrar por componente
jq 'select(.component == "evidence")' /opt/takedown/logs/takedown.log

# Análise temporal
jq -r '[.timestamp, .level, .message] | @tsv' /opt/takedown/logs/takedown.log
```

---

**Links Úteis:**
- 📖 [Documentação Principal](../README.md)
- 🔧 [Guia de Instalação](../installation/README.md)
- 🚀 [Deploy em Produção](../deployment/README.md)
- 👨‍💻 [Guia de Desenvolvimento](../development/README.md)