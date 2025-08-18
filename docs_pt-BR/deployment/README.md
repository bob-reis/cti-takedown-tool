# 🚀 Guia de Deployment e Produção

Este guia detalha como fazer o deploy do CTI Takedown Tool em ambientes de produção, incluindo configurações de alta disponibilidade, monitoring, backup e manutenção.

## 📋 Índice

- [Ambientes](#ambientes)
- [Estratégias de Deploy](#estratégias-de-deploy)
- [Configuração de Produção](#configuração-de-produção)
- [Alta Disponibilidade](#alta-disponibilidade)
- [Monitoring e Alertas](#monitoring-e-alertas)
- [Backup e Recovery](#backup-e-recovery)
- [Manutenção](#manutenção)
- [Security Hardening](#security-hardening)

## 🌍 Ambientes

### Ambiente de Staging

```bash
# Configuração para staging
export TAKEDOWN_ENV=staging
export TAKEDOWN_CONFIG_DIR=/opt/takedown/configs/staging
export TAKEDOWN_LOG_LEVEL=info
export TAKEDOWN_WORKERS=3

# Build para staging
CGO_ENABLED=0 go build -o takedown-staging -ldflags "-X main.version=$(git describe --tags)" cmd/takedown/main.go
```

### Ambiente de Produção

```bash
# Configuração para produção
export TAKEDOWN_ENV=production
export TAKEDOWN_CONFIG_DIR=/opt/takedown/configs/production
export TAKEDOWN_LOG_LEVEL=warn
export TAKEDOWN_WORKERS=10

# Build otimizado para produção
CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static" -X main.version=$(git describe --tags) -X main.buildTime=$(date -u +"%Y-%m-%dT%H:%M:%SZ") -s -w' -o takedown cmd/takedown/main.go
```

## 🐳 Estratégias de Deploy

### 1. Docker Container

#### Dockerfile Otimizado

```dockerfile
# Multi-stage build para otimização
FROM golang:1.22-alpine AS builder

# Instalar dependências de build
RUN apk add --no-cache git ca-certificates tzdata

# Criar usuário não-root
RUN addgroup -g 10001 takedown && \
    adduser -D -u 10001 -G takedown takedown

WORKDIR /src

# Copy go mod e sum para cache de dependências
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build da aplicação
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static" -s -w' -o takedown cmd/takedown/main.go

# Imagem final mínima
FROM scratch

# Copy certificados SSL
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy timezone data
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy usuário
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group

# Copy aplicação e configs
COPY --from=builder /src/takedown /usr/local/bin/takedown
COPY --from=builder /src/configs /opt/takedown/configs

# Usar usuário não-root
USER 10001:10001

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD ["/usr/local/bin/takedown", "-action=health"]

# Configuração padrão
ENV TAKEDOWN_CONFIG_DIR=/opt/takedown/configs
ENV TAKEDOWN_LOG_LEVEL=info

EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/takedown"]
CMD ["-daemon"]
```

#### Docker Compose para Produção

```yaml
# docker-compose.prod.yml
version: '3.8'

services:
  takedown:
    build:
      context: .
      dockerfile: Dockerfile
    image: takedown:latest
    container_name: takedown-prod
    restart: unless-stopped
    
    # Configuração de recursos
    mem_limit: 512m
    mem_reservation: 256m
    cpus: '1.0'
    
    # Variáveis de ambiente
    environment:
      - TAKEDOWN_ENV=production
      - TAKEDOWN_WORKERS=10
      - TAKEDOWN_LOG_LEVEL=warn
      - SMTP_HOST=${SMTP_HOST}
      - SMTP_USER=${SMTP_USER}
      - SMTP_PASS=${SMTP_PASS}
    
    # Volumes para persistência
    volumes:
      - /opt/takedown/data:/app/data
      - /opt/takedown/logs:/app/logs
      - /opt/takedown/configs:/opt/takedown/configs:ro
    
    # Rede
    networks:
      - takedown-network
    
    # Health check
    healthcheck:
      test: ["CMD", "/usr/local/bin/takedown", "-action=health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s
    
    # Logging
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"

  # Opcional: NGINX como reverse proxy
  nginx:
    image: nginx:alpine
    container_name: takedown-nginx
    restart: unless-stopped
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf:ro
      - /opt/ssl:/etc/nginx/ssl:ro
    networks:
      - takedown-network
    depends_on:
      - takedown

networks:
  takedown-network:
    driver: bridge
```

### 2. Kubernetes Deployment

#### Namespace e ServiceAccount

```yaml
# k8s/namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: takedown-system
  labels:
    name: takedown-system
    app.kubernetes.io/name: takedown
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: takedown
  namespace: takedown-system
```

#### ConfigMap e Secrets

```yaml
# k8s/configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: takedown-config
  namespace: takedown-system
data:
  production.yaml: |
    workers: 20
    timeout: 300s
    log_level: warn
    queue_size: 1000
    
    features:
      auto_escalation: true
      bulk_processing: true
      ml_scoring: false
---
apiVersion: v1
kind: Secret
metadata:
  name: takedown-secrets
  namespace: takedown-system
type: Opaque
stringData:
  smtp-host: "smtp.company.com"
  smtp-user: "takedown@company.com"
  smtp-pass: "secure_password"
```

#### Deployment

```yaml
# k8s/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: takedown
  namespace: takedown-system
  labels:
    app: takedown
    version: v1.0.0
spec:
  replicas: 3
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 1
  selector:
    matchLabels:
      app: takedown
  template:
    metadata:
      labels:
        app: takedown
        version: v1.0.0
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "8080"
        prometheus.io/path: "/metrics"
    spec:
      serviceAccountName: takedown
      securityContext:
        runAsNonRoot: true
        runAsUser: 10001
        runAsGroup: 10001
        fsGroup: 10001
      
      containers:
      - name: takedown
        image: takedown:v1.0.0
        imagePullPolicy: IfNotPresent
        
        ports:
        - containerPort: 8080
          name: http
          protocol: TCP
        
        env:
        - name: TAKEDOWN_ENV
          value: "production"
        - name: TAKEDOWN_CONFIG_DIR
          value: "/opt/takedown/configs"
        - name: TAKEDOWN_WORKERS
          value: "20"
        - name: SMTP_HOST
          valueFrom:
            secretKeyRef:
              name: takedown-secrets
              key: smtp-host
        - name: SMTP_USER
          valueFrom:
            secretKeyRef:
              name: takedown-secrets
              key: smtp-user
        - name: SMTP_PASS
          valueFrom:
            secretKeyRef:
              name: takedown-secrets
              key: smtp-pass
        
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        
        livenessProbe:
          exec:
            command:
            - /usr/local/bin/takedown
            - -action=health
          initialDelaySeconds: 30
          periodSeconds: 60
          timeoutSeconds: 10
          failureThreshold: 3
        
        readinessProbe:
          exec:
            command:
            - /usr/local/bin/takedown
            - -action=health
          initialDelaySeconds: 10
          periodSeconds: 30
          timeoutSeconds: 5
          failureThreshold: 3
        
        volumeMounts:
        - name: config
          mountPath: /opt/takedown/configs
          readOnly: true
        - name: data
          mountPath: /app/data
        - name: logs
          mountPath: /app/logs
      
      volumes:
      - name: config
        configMap:
          name: takedown-config
      - name: data
        persistentVolumeClaim:
          claimName: takedown-data
      - name: logs
        persistentVolumeClaim:
          claimName: takedown-logs
```

#### Services e Ingress

```yaml
# k8s/service.yaml
apiVersion: v1
kind: Service
metadata:
  name: takedown
  namespace: takedown-system
  labels:
    app: takedown
spec:
  selector:
    app: takedown
  ports:
  - port: 8080
    targetPort: 8080
    protocol: TCP
    name: http
  type: ClusterIP
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: takedown
  namespace: takedown-system
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
spec:
  tls:
  - hosts:
    - takedown.company.com
    secretName: takedown-tls
  rules:
  - host: takedown.company.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: takedown
            port:
              number: 8080
```

### 3. Systemd Service (Linux)

```ini
# /etc/systemd/system/takedown.service
[Unit]
Description=CTI Takedown Tool
Documentation=https://github.com/cti-team/takedown
After=network.target
Wants=network.target

[Service]
Type=simple
User=takedown
Group=takedown
WorkingDirectory=/opt/takedown
ExecStart=/opt/takedown/bin/takedown -daemon -config=/opt/takedown/configs/production.yaml
ExecReload=/bin/kill -HUP $MAINPID
Restart=on-failure
RestartSec=5
StandardOutput=journal
StandardError=journal
SyslogIdentifier=takedown

# Configurações de ambiente
Environment=TAKEDOWN_ENV=production
Environment=TAKEDOWN_LOG_LEVEL=warn
Environment=TAKEDOWN_WORKERS=10

# Configurações de segurança
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/opt/takedown/data /opt/takedown/logs

# Limites de recursos
LimitNOFILE=65536
LimitNPROC=4096

[Install]
WantedBy=multi-user.target
```

```bash
# Instalar e configurar service
sudo systemctl daemon-reload
sudo systemctl enable takedown.service
sudo systemctl start takedown.service

# Verificar status
sudo systemctl status takedown.service
```

## ⚙️ Configuração de Produção

### 1. Estrutura de Diretórios

```bash
# Criar estrutura de produção
sudo mkdir -p /opt/takedown/{bin,configs,data,logs,backups}
sudo chown -R takedown:takedown /opt/takedown

# Estrutura recomendada
/opt/takedown/
├── bin/
│   └── takedown                    # Binário principal
├── configs/
│   ├── production.yaml            # Config principal
│   ├── smtp.yaml                  # Config SMTP
│   ├── sla/                       # SLAs por ambiente
│   ├── routing/                   # Regras de routing
│   └── templates/                 # Templates de email
├── data/
│   ├── state.db                   # Estado persistente
│   └── evidence/                  # Evidências coletadas
├── logs/
│   ├── takedown.log              # Log principal
│   ├── access.log                # Log de acesso
│   └── error.log                 # Log de erros
└── backups/
    ├── configs/                   # Backup de configs
    └── data/                      # Backup de dados
```

### 2. Configuração de Produção

```yaml
# /opt/takedown/configs/production.yaml
# Configuração principal de produção
server:
  environment: "production"
  workers: 20
  queue_size: 2000
  timeout: 300s
  max_concurrent_cases: 500
  
# Logging otimizado
logging:
  level: "warn"
  file: "/opt/takedown/logs/takedown.log"
  max_size: 100  # MB
  max_backups: 7
  max_age: 30    # dias
  format: "json"
  
# Performance tuning
performance:
  evidence_cache_ttl: "1h"
  rdap_cache_ttl: "24h"
  dns_timeout: "10s"
  http_timeout: "30s"
  
# Rate limiting
rate_limiting:
  enabled: true
  requests_per_minute: 100
  burst_size: 200
  
# Configurações de rede
network:
  dns_servers: ["1.1.1.1", "8.8.8.8"]
  http_proxy: ""
  https_proxy: ""
  no_proxy: "localhost,127.0.0.1,.company.com"
  
# Features flags
features:
  auto_escalation: true
  bulk_processing: true
  ml_scoring: false
  webhook_notifications: true
  
# Métricas e monitoring
metrics:
  enabled: true
  endpoint: "/metrics"
  port: 8080
  
# Health checks
health:
  enabled: true
  endpoint: "/health"
  port: 8080
```

### 3. Configuração de Proxy (NGINX)

```nginx
# /etc/nginx/sites-available/takedown
upstream takedown_backend {
    least_conn;
    server 127.0.0.1:8080 max_fails=3 fail_timeout=30s;
    server 127.0.0.1:8081 max_fails=3 fail_timeout=30s backup;
}

server {
    listen 80;
    server_name takedown.company.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name takedown.company.com;
    
    # SSL configuration
    ssl_certificate /etc/ssl/certs/takedown.company.com.pem;
    ssl_certificate_key /etc/ssl/private/takedown.company.com.key;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-RSA-AES256-GCM-SHA512:DHE-RSA-AES256-GCM-SHA512:ECDHE-RSA-AES256-GCM-SHA384;
    ssl_prefer_server_ciphers off;
    
    # Security headers
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header Referrer-Policy "no-referrer-when-downgrade" always;
    add_header Content-Security-Policy "default-src 'self' http: https: data: blob: 'unsafe-inline'" always;
    
    # Logging
    access_log /var/log/nginx/takedown_access.log;
    error_log /var/log/nginx/takedown_error.log;
    
    # Health check endpoint
    location /health {
        proxy_pass http://takedown_backend;
        access_log off;
    }
    
    # Metrics endpoint (restrito)
    location /metrics {
        allow 10.0.0.0/8;
        allow 172.16.0.0/12;
        allow 192.168.0.0/16;
        deny all;
        
        proxy_pass http://takedown_backend;
    }
    
    # API endpoints
    location /api/ {
        proxy_pass http://takedown_backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # Rate limiting
        limit_req zone=api burst=20 nodelay;
        
        # Timeouts
        proxy_connect_timeout 30s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }
    
    # Static files (futuro dashboard)
    location / {
        root /opt/takedown/web;
        try_files $uri $uri/ /index.html;
        expires 1h;
    }
}

# Rate limiting zones
http {
    limit_req_zone $binary_remote_addr zone=api:10m rate=10r/m;
}
```

## 🏗️ Alta Disponibilidade

### 1. Load Balancer Setup

```yaml
# HAProxy configuration
# /etc/haproxy/haproxy.cfg
global
    daemon
    user haproxy
    group haproxy
    log stdout local0
    
defaults
    mode http
    timeout connect 5000ms
    timeout client 50000ms
    timeout server 50000ms
    option httplog
    
frontend takedown_frontend
    bind *:80
    bind *:443 ssl crt /etc/ssl/certs/takedown.pem
    redirect scheme https if !{ ssl_fc }
    
    # Health check
    acl health_check path_beg /health
    use_backend takedown_health if health_check
    
    default_backend takedown_backend
    
backend takedown_backend
    balance roundrobin
    option httpchk GET /health
    
    server takedown1 10.0.1.10:8080 check inter 30s
    server takedown2 10.0.1.11:8080 check inter 30s
    server takedown3 10.0.1.12:8080 check inter 30s backup
    
backend takedown_health
    option httpchk GET /health
    server takedown1 10.0.1.10:8080 check
    server takedown2 10.0.1.11:8080 check
```

### 2. Database Clustering (Futuro)

```yaml
# Redis Cluster para state sharing
version: '3.8'

services:
  redis-master:
    image: redis:7-alpine
    command: redis-server --appendonly yes --replica-read-only no
    volumes:
      - redis-master-data:/data
    networks:
      - redis-cluster
    
  redis-replica:
    image: redis:7-alpine
    command: redis-server --slaveof redis-master 6379 --appendonly yes
    depends_on:
      - redis-master
    networks:
      - redis-cluster
    
  redis-sentinel:
    image: redis:7-alpine
    command: redis-sentinel /etc/redis/sentinel.conf
    volumes:
      - ./sentinel.conf:/etc/redis/sentinel.conf
    depends_on:
      - redis-master
      - redis-replica
    networks:
      - redis-cluster

volumes:
  redis-master-data:

networks:
  redis-cluster:
    driver: bridge
```

## 📊 Monitoring e Alertas

### 1. Prometheus Configuration

```yaml
# prometheus.yml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

rule_files:
  - "takedown_rules.yml"

scrape_configs:
  - job_name: 'takedown'
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: '/metrics'
    scrape_interval: 30s
    
alerting:
  alertmanagers:
    - static_configs:
        - targets:
          - alertmanager:9093
```

### 2. Alerting Rules

```yaml
# takedown_rules.yml
groups:
- name: takedown.rules
  rules:
  
  # Alto número de casos pendentes
  - alert: TakedownHighPendingCases
    expr: takedown_cases_pending > 100
    for: 5m
    labels:
      severity: warning
    annotations:
      summary: "Alto número de casos pendentes"
      description: "{{ $value }} casos pendentes há mais de 5 minutos"
  
  # Taxa de falha alta
  - alert: TakedownHighFailureRate
    expr: rate(takedown_cases_failed_total[5m]) > 0.1
    for: 10m
    labels:
      severity: critical
    annotations:
      summary: "Taxa de falha alta no Takedown"
      description: "Taxa de falha: {{ $value | humanizePercentage }}"
  
  # Timeout de SMTP
  - alert: TakedownSMTPDown
    expr: takedown_smtp_errors_total > 5
    for: 2m
    labels:
      severity: critical
    annotations:
      summary: "Problemas de conectividade SMTP"
      description: "{{ $value }} erros SMTP nos últimos 2 minutos"
  
  # Uso de memória alto
  - alert: TakedownHighMemoryUsage
    expr: process_resident_memory_bytes{job="takedown"} > 512*1024*1024
    for: 15m
    labels:
      severity: warning
    annotations:
      summary: "Alto uso de memória no Takedown"
      description: "Uso de memória: {{ $value | humanizeBytes }}"
```

### 3. Grafana Dashboard

```json
{
  "dashboard": {
    "title": "CTI Takedown Tool",
    "panels": [
      {
        "title": "Cases por Status",
        "type": "stat",
        "targets": [
          {
            "expr": "takedown_cases_total",
            "legendFormat": "{{status}}"
          }
        ]
      },
      {
        "title": "Throughput de Cases",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(takedown_cases_processed_total[5m])",
            "legendFormat": "Cases/segundo"
          }
        ]
      },
      {
        "title": "SLA Compliance",
        "type": "singlestat",
        "targets": [
          {
            "expr": "takedown_sla_compliance_percentage",
            "legendFormat": "SLA %"
          }
        ]
      }
    ]
  }
}
```

## 💾 Backup e Recovery

### 1. Backup Strategy

```bash
#!/bin/bash
# /opt/takedown/scripts/backup.sh

BACKUP_DIR="/opt/takedown/backups"
DATE=$(date +%Y%m%d_%H%M%S)
RETENTION_DAYS=30

# Backup de configurações
tar -czf "${BACKUP_DIR}/configs_${DATE}.tar.gz" -C /opt/takedown configs/

# Backup de dados
tar -czf "${BACKUP_DIR}/data_${DATE}.tar.gz" -C /opt/takedown data/

# Backup de logs (últimos 7 dias)
find /opt/takedown/logs -name "*.log" -mtime -7 | \
    tar -czf "${BACKUP_DIR}/logs_${DATE}.tar.gz" -T -

# Cleanup de backups antigos
find "${BACKUP_DIR}" -name "*.tar.gz" -mtime +${RETENTION_DAYS} -delete

# Upload para S3 (opcional)
if [ -n "$S3_BUCKET" ]; then
    aws s3 cp "${BACKUP_DIR}/configs_${DATE}.tar.gz" "s3://${S3_BUCKET}/takedown/backups/"
    aws s3 cp "${BACKUP_DIR}/data_${DATE}.tar.gz" "s3://${S3_BUCKET}/takedown/backups/"
fi

echo "Backup completed: ${DATE}"
```

### 2. Restore Procedure

```bash
#!/bin/bash
# /opt/takedown/scripts/restore.sh

BACKUP_FILE=$1
BACKUP_DIR="/opt/takedown/backups"

if [ -z "$BACKUP_FILE" ]; then
    echo "Usage: $0 <backup_file>"
    echo "Available backups:"
    ls -la ${BACKUP_DIR}/*.tar.gz
    exit 1
fi

# Parar serviço
sudo systemctl stop takedown

# Fazer backup dos dados atuais
mv /opt/takedown/data /opt/takedown/data.bak.$(date +%Y%m%d_%H%M%S)

# Restaurar dados
tar -xzf "${BACKUP_DIR}/${BACKUP_FILE}" -C /opt/takedown/

# Ajustar permissões
sudo chown -R takedown:takedown /opt/takedown/data

# Reiniciar serviço
sudo systemctl start takedown

# Verificar status
sudo systemctl status takedown

echo "Restore completed from: ${BACKUP_FILE}"
```

### 3. Automated Backup (Cron)

```bash
# /etc/cron.d/takedown-backup
# Backup diário às 2:00 AM
0 2 * * * takedown /opt/takedown/scripts/backup.sh > /opt/takedown/logs/backup.log 2>&1

# Backup de configs toda hora (apenas se houver mudanças)
0 * * * * takedown /opt/takedown/scripts/config_backup.sh > /dev/null 2>&1
```

## 🔧 Manutenção

### 1. Log Rotation

```bash
# /etc/logrotate.d/takedown
/opt/takedown/logs/*.log {
    daily
    missingok
    rotate 30
    compress
    delaycompress
    notifempty
    create 0644 takedown takedown
    postrotate
        /bin/kill -HUP `cat /var/run/takedown.pid 2> /dev/null` 2> /dev/null || true
    endscript
}
```

### 2. Health Monitoring Script

```bash
#!/bin/bash
# /opt/takedown/scripts/health_monitor.sh

HEALTH_URL="http://localhost:8080/health"
LOG_FILE="/opt/takedown/logs/health_monitor.log"
PID_FILE="/var/run/takedown.pid"

# Verificar se processo está rodando
if [ -f "$PID_FILE" ]; then
    PID=$(cat $PID_FILE)
    if ! ps -p $PID > /dev/null 2>&1; then
        echo "$(date): Process not running, restarting..." >> $LOG_FILE
        sudo systemctl restart takedown
        exit 1
    fi
else
    echo "$(date): PID file not found, service may be down" >> $LOG_FILE
    sudo systemctl restart takedown
    exit 1
fi

# Verificar health endpoint
HTTP_STATUS=$(curl -s -o /dev/null -w "%{http_code}" $HEALTH_URL)

if [ "$HTTP_STATUS" != "200" ]; then
    echo "$(date): Health check failed (HTTP $HTTP_STATUS), restarting..." >> $LOG_FILE
    sudo systemctl restart takedown
    exit 1
fi

# Verificar métricas básicas
PENDING_CASES=$(curl -s $HEALTH_URL | jq '.pending_cases // 0')
if [ "$PENDING_CASES" -gt 1000 ]; then
    echo "$(date): High pending cases ($PENDING_CASES), alerting..." >> $LOG_FILE
    # Enviar alerta (webhook, email, etc.)
fi

echo "$(date): Health check OK" >> $LOG_FILE
```

### 3. Update Strategy

```bash
#!/bin/bash
# /opt/takedown/scripts/update.sh

NEW_VERSION=$1
BACKUP_DIR="/opt/takedown/backups"
CURRENT_VERSION=$(./takedown --version 2>/dev/null || echo "unknown")

if [ -z "$NEW_VERSION" ]; then
    echo "Usage: $0 <new_version>"
    exit 1
fi

echo "Updating from $CURRENT_VERSION to $NEW_VERSION"

# 1. Backup atual
echo "Creating backup..."
tar -czf "${BACKUP_DIR}/pre_update_$(date +%Y%m%d_%H%M%S).tar.gz" \
    -C /opt/takedown bin/ configs/ data/

# 2. Download nova versão
echo "Downloading new version..."
wget -O /tmp/takedown-${NEW_VERSION} \
    "https://github.com/cti-team/takedown/releases/download/${NEW_VERSION}/takedown-linux-amd64"

# 3. Validar binário
echo "Validating binary..."
chmod +x /tmp/takedown-${NEW_VERSION}
if ! /tmp/takedown-${NEW_VERSION} --version; then
    echo "Invalid binary"
    exit 1
fi

# 4. Aplicar update
echo "Applying update..."
sudo systemctl stop takedown
cp /opt/takedown/bin/takedown /opt/takedown/bin/takedown.backup
cp /tmp/takedown-${NEW_VERSION} /opt/takedown/bin/takedown
chmod +x /opt/takedown/bin/takedown

# 5. Testar configuração
echo "Testing configuration..."
if ! /opt/takedown/bin/takedown -action=validate-config; then
    echo "Configuration validation failed, rolling back..."
    cp /opt/takedown/bin/takedown.backup /opt/takedown/bin/takedown
    sudo systemctl start takedown
    exit 1
fi

# 6. Restart serviço
echo "Restarting service..."
sudo systemctl start takedown

# 7. Verificar saúde
sleep 10
if ! curl -s http://localhost:8080/health > /dev/null; then
    echo "Health check failed, rolling back..."
    sudo systemctl stop takedown
    cp /opt/takedown/bin/takedown.backup /opt/takedown/bin/takedown
    sudo systemctl start takedown
    exit 1
fi

echo "Update completed successfully to version $NEW_VERSION"
rm /tmp/takedown-${NEW_VERSION}
```

## 🔒 Security Hardening

### 1. Sistema Operacional

```bash
#!/bin/bash
# Security hardening script

# Atualizar sistema
sudo apt update && sudo apt upgrade -y

# Configurar firewall
sudo ufw default deny incoming
sudo ufw default allow outgoing
sudo ufw allow ssh
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw --force enable

# Configurar fail2ban
sudo apt install fail2ban -y
sudo systemctl enable fail2ban
sudo systemctl start fail2ban

# Desabilitar serviços desnecessários
sudo systemctl disable avahi-daemon
sudo systemctl disable cups
sudo systemctl disable bluetooth

# Configurar limites de sistema
cat >> /etc/security/limits.conf << EOF
takedown soft nofile 65536
takedown hard nofile 65536
takedown soft nproc 4096
takedown hard nproc 4096
EOF
```

### 2. Application Security

```yaml
# Configurações de segurança da aplicação
security:
  # Rate limiting
  rate_limiting:
    enabled: true
    requests_per_minute: 60
    burst_size: 100
    
  # Input validation
  input_validation:
    max_ioc_length: 2048
    max_tags: 10
    allowed_ioc_types: ["url", "domain", "ip"]
    
  # Output sanitization
  output_sanitization:
    auto_defang: true
    strip_html: true
    escape_markdown: true
    
  # TLS configuration
  tls:
    min_version: "1.2"
    cipher_suites:
      - "TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384"
      - "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256"
    
  # Headers de segurança
  security_headers:
    x_frame_options: "SAMEORIGIN"
    x_content_type_options: "nosniff"
    x_xss_protection: "1; mode=block"
    strict_transport_security: "max-age=31536000"
    content_security_policy: "default-src 'self'"
```

### 3. Secrets Management

```bash
# Usar HashiCorp Vault ou AWS Secrets Manager
# Exemplo com AWS Secrets Manager

# Criar secrets
aws secretsmanager create-secret \
    --name "takedown/smtp/credentials" \
    --description "SMTP credentials for Takedown Tool" \
    --secret-string '{"host":"smtp.company.com","user":"takedown@company.com","pass":"secure_password"}'

# Script para recuperar secrets
#!/bin/bash
# /opt/takedown/scripts/get_secrets.sh

SECRET_ID="takedown/smtp/credentials"
AWS_REGION="us-east-1"

# Recuperar secret
SECRET_VALUE=$(aws secretsmanager get-secret-value \
    --secret-id $SECRET_ID \
    --region $AWS_REGION \
    --query SecretString \
    --output text)

# Exportar como variáveis de ambiente
eval $(echo $SECRET_VALUE | jq -r 'to_entries[] | "export SMTP_\(.key | ascii_upcase)=\(.value)"')

# Executar aplicação
exec /opt/takedown/bin/takedown "$@"
```

---

**Próximo**: [Troubleshooting Guide](../troubleshooting/README.md)