# 🚀 GitHub Setup & CI/CD Guide

Este documento contém as instruções para configurar o projeto no GitHub com integração completa de CI/CD, SonarCloud e automação de quality gates.

## 📋 Arquivos Criados para GitHub

### 🔧 Configuração Principal
- `.gitignore` - Ignora arquivos desnecessários (binários, logs, secrets, etc.)
- `.golangci.yml` - Configuração completa do linter para Go
- `sonar-project.properties` - Configuração do SonarCloud

### 🤖 GitHub Actions Workflows
- `.github/workflows/ci.yml` - Pipeline principal de CI/CD
- `.github/workflows/pr-validation.yml` - Validação obrigatória de Pull Requests
- `.github/workflows/sonarcloud.yml` - Integração com SonarCloud

### 📝 Templates
- `.github/PULL_REQUEST_TEMPLATE.md` - Template para Pull Requests
- `.github/ISSUE_TEMPLATE/bug_report.md` - Template para reports de bugs
- `.github/ISSUE_TEMPLATE/feature_request.md` - Template para solicitações de features

## 🛠️ Setup do Repositório GitHub

### 1. Criar Repositório

```bash
# Inicializar repositório local
git init
git add .
git commit -m "Initial commit: CTI Takedown Tool v1.0.0"

# Adicionar remote origin
git remote add origin https://github.com/cti-team/takedown.git
git branch -M main
git push -u origin main
```

### 2. Configurar Secrets

Acesse `Settings > Secrets and variables > Actions` e adicione:

#### Secrets Obrigatórios:
```bash
# SonarCloud
SONAR_TOKEN=your_sonarcloud_token

# Docker Hub (opcional)
DOCKER_USERNAME=your_docker_username
DOCKER_PASSWORD=your_docker_password

# Slack Notifications (opcional)
SLACK_WEBHOOK=your_slack_webhook_url
```

#### Como obter os tokens:

**SONAR_TOKEN:**
1. Acesse [SonarCloud](https://sonarcloud.io)
2. Login com GitHub
3. Vá em `My Account > Security > Generate Tokens`
4. Crie token com nome "GitHub Actions"
5. Copie o token gerado

**DOCKER_USERNAME/PASSWORD:**
1. Crie conta no [Docker Hub](https://hub.docker.com)
2. Vá em `Account Settings > Security > Access Tokens`
3. Crie novo token para CI/CD

### 3. Configurar Branch Protection

Acesse `Settings > Branches` e configure para `main`:

```yaml
Branch protection rules:
✅ Require a pull request before merging
  ✅ Require approvals: 1
  ✅ Dismiss stale PR approvals when new commits are pushed
  ✅ Require review from code owners

✅ Require status checks to pass before merging
  ✅ Require branches to be up to date before merging
  Required status checks:
    - Test and Validate
    - Security Analysis  
    - SonarCloud Scan
    - PR Gate - Quality Checks

✅ Require conversation resolution before merging
✅ Restrict pushes that create files larger than 100MB
✅ Allow force pushes: ❌
✅ Allow deletions: ❌
```

## 🔍 SonarCloud Setup

### 1. Configurar Projeto

1. Acesse [SonarCloud](https://sonarcloud.io)
2. Click `+` > `Analyze new project`
3. Selecione `cti-team/takedown`
4. Configure:
   - **Project Key**: `cti-team_takedown`
   - **Organization**: `cti-team`
   - **Display Name**: `CTI Takedown Tool`

### 2. Quality Gate

Configure quality gate personalizado:

```yaml
Conditions:
- Coverage: > 80%
- Duplicated Lines: < 3%
- Maintainability Rating: A
- Reliability Rating: A  
- Security Rating: A
- Security Hotspots: 0
- New Code Coverage: > 80%
- New Duplicated Lines: < 3%
```

### 3. Configuração do Projeto

O arquivo `sonar-project.properties` já está configurado com:
- Exclusões de arquivos de teste e documentação
- Configuração específica para Go
- Path de coverage reports
- Configurações de qualidade

## 🔄 CI/CD Pipelines

### Pipeline Principal (ci.yml)

**Triggers:**
- Push para `main` e `develop`
- Pull Requests para `main` e `develop`

**Jobs:**
1. **Test**: Testes unitários, linting, coverage
2. **Security**: Scan de segurança com Gosec e Nancy
3. **SonarCloud**: Análise de qualidade de código
4. **Build Release**: Build multi-platform (main apenas)
5. **Docker**: Build de imagem Docker (main apenas)
6. **Performance**: Benchmarks (main apenas)
7. **Notify**: Notificações Slack (main apenas)

### Pipeline de PR (pr-validation.yml)

**Validações Obrigatórias:**
- ✅ Formatação de código (`gofmt`)
- ✅ Linting (`go vet`, `staticcheck`, `golangci-lint`)
- ✅ Testes unitários (100% deve passar)
- ✅ Coverage mínimo (80%)
- ✅ Build successful
- ✅ Validação de configuração
- ✅ Detecção de dados sensíveis
- ✅ Verificação de documentação

**Features Especiais:**
- Comentários automáticos no PR com status
- Análise de impacto de mudanças
- Validação de arquivos críticos
- Bloqueio automático se falhar

### Pipeline SonarCloud (sonarcloud.yml)

**Análises:**
- Qualidade de código
- Coverage de testes
- Detecção de bugs e vulnerabilidades
- Code smells e duplicação
- Security hotspots

## 📊 Quality Gates

### Criteria de Qualidade

**Para Merge em Main:**
1. ✅ Todos os testes passando
2. ✅ Coverage ≥ 80%
3. ✅ Zero vulnerabilidades críticas
4. ✅ SonarCloud Quality Gate: PASSED
5. ✅ Pelo menos 1 aprovação de review
6. ✅ Todos os comentários resolvidos

**Para Releases:**
1. ✅ Todos os critérios de merge
2. ✅ Benchmarks de performance executados
3. ✅ Build multi-platform successful
4. ✅ Docker image criada
5. ✅ Documentation atualizada

## 🚀 Fluxo de Desenvolvimento

### Processo Padrão

```bash
# 1. Criar feature branch
git checkout -b feature/amazing-feature

# 2. Desenvolver e testar localmente
go test ./...
golangci-lint run ./...
./test.sh

# 3. Commit e push
git add .
git commit -m "feat: add amazing feature"
git push origin feature/amazing-feature

# 4. Criar Pull Request
# - Usar template automático
# - Aguardar validações automáticas
# - Corrigir issues se houver
# - Solicitar review

# 5. Merge após aprovação
# - Automático após approvals
# - Squash and merge recomendado
```

### Exemplo de Workflow

1. **Desenvolvimento Local**
   ```bash
   # Testar antes de push
   make test        # ou go test ./...
   make lint        # ou golangci-lint run
   make build       # ou go build
   ```

2. **Push para GitHub**
   - GitHub Actions executará automaticamente
   - Validação de PR bloqueará merge se falhar
   - SonarCloud analisará qualidade

3. **Review Process**
   - Reviewer verifica código
   - Automated checks devem passar
   - Merge após aprovação

4. **Deploy Automático**
   - Merge para `main` trigga build de release
   - Docker image criada automaticamente
   - Artefatos disponíveis para download

## 🔧 Troubleshooting

### Issues Comuns

**1. SonarCloud Token Inválido**
```bash
Error: Invalid authentication token
Solução: Verificar SONAR_TOKEN em GitHub Secrets
```

**2. Coverage Abaixo do Mínimo**
```bash
Error: Coverage 75% is below minimum threshold of 80%
Solução: Adicionar mais testes ou ajustar threshold
```

**3. Linter Failures**
```bash
Error: golangci-lint found issues
Solução: Corrigir issues ou adicionar //nolint se necessário
```

**4. Branch Protection**
```bash
Error: Required status check "Test and Validate" is expected
Solução: Aguardar conclusão dos checks ou corrigir falhas
```

### Debug Local

```bash
# Simular CI localmente
docker run --rm -v $(pwd):/app -w /app golang:1.22 ./test.sh

# Executar SonarCloud localmente
sonar-scanner \
  -Dsonar.projectKey=cti-team_takedown \
  -Dsonar.sources=. \
  -Dsonar.host.url=https://sonarcloud.io \
  -Dsonar.login=$SONAR_TOKEN
```

## 📈 Métricas e Monitoring

### Dashboards Disponíveis

1. **GitHub Actions**: `Actions` tab - Status de builds
2. **SonarCloud**: Quality metrics e trends
3. **GitHub Insights**: Activity, contributors, traffic
4. **Releases**: Download metrics e adoption

### KPIs Importantes

- **Build Success Rate**: > 95%
- **Test Coverage**: > 80%
- **PR Merge Time**: < 24h
- **Issues Resolution**: < 7 days
- **Security Vulnerabilities**: 0 critical

---

**🎉 Setup Completo!** 

Seu repositório agora tem CI/CD profissional com quality gates rigorosos, integração SonarCloud e automação completa de desenvolvimento.

Para mais informações, consulte a [documentação principal](README.md).