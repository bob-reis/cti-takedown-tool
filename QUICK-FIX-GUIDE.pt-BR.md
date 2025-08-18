# 🚨 Quick Fix Guide - GitHub Actions & SonarCloud

## 🔍 **Problemas Identificados:**

### 1. **GitHub Actions Failing**
- Workflow usando action incorreta para Go
- Configuração SonarCloud incompleta
- Missing coverage file generation

### 2. **SonarCloud Setup Issues**
- Project key mismatch
- Organization mismatch  
- Go-specific configuration missing

## ⚡ **Soluções Rápidas:**

### 1. **Substituir o arquivo build.yml**

Substitua o conteúdo de `.github/workflows/build.yml` por:

```yaml
name: Build and Analysis

on:
  push:
    branches:
      - main
  pull_request:
    types: [opened, synchronize, reopened]

env:
  GO_VERSION: "1.22"

jobs:
  test:
    name: Test and Coverage
    runs-on: ubuntu-latest
    
    steps:
    - name: Checkout repository
      uses: actions/checkout@v4
      with:
        fetch-depth: 0
    
    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: ${{ env.GO_VERSION }}
    
    - name: Run tests with coverage
      run: |
        go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
    
    - name: Verify coverage file
      run: |
        if [ ! -f coverage.out ]; then
          echo "❌ Coverage file not found"
          exit 1
        fi
        echo "✅ Coverage file generated successfully"

  sonarcloud:
    name: SonarCloud Analysis
    runs-on: ubuntu-latest
    needs: test
    
    steps:
    - name: Checkout repository
      uses: actions/checkout@v4
      with:
        fetch-depth: 0
    
    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: ${{ env.GO_VERSION }}
    
    - name: Run tests with coverage for SonarCloud
      run: |
        go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
    
    - name: SonarCloud Scan
      uses: SonarSource/sonarcloud-github-action@master
      env:
        GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        SONAR_TOKEN: ${{ secrets.SONAR_TOKEN }}
```

### 2. **Atualizar sonar-project.properties**

```properties
sonar.projectKey=bob-reis_site-takedown
sonar.organization=bob-reis
sonar.projectName=CTI Takedown Tool

# Go coverage
sonar.go.coverage.reportPaths=coverage.out

# Source configuration
sonar.sources=.
sonar.tests=.
sonar.inclusions=**/*.go
sonar.exclusions=**/*_test.go,**/vendor/**,docs/**,*.md
sonar.test.inclusions=**/*_test.go
sonar.coverage.exclusions=**/*_test.go,cmd/**
```

### 3. **Verificar SonarCloud Setup**

1. **Acesse**: https://sonarcloud.io
2. **Login** com GitHub
3. **Import project**: `bob-reis/site-takedown`
4. **Verificar**:
   - Project Key: `bob-reis_site-takedown`
   - Organization: `bob-reis`

### 4. **Configurar GitHub Secrets**

Acesse: `https://github.com/bob-reis/site-takedown/settings/secrets/actions`

**Adicionar**:
```
Name: SONAR_TOKEN
Value: [seu_token_do_sonarcloud]
```

**Como obter SONAR_TOKEN**:
1. SonarCloud → My Account → Security
2. Generate Tokens
3. Name: "GitHub Actions"
4. Copiar o token gerado

## 🔄 **Após as Correções:**

### 1. **Commit e Push**
```bash
git add .
git commit -m "fix: correct GitHub Actions and SonarCloud configuration

- Replace sonarqube-scan-action with sonarcloud-github-action
- Add proper Go test coverage generation
- Update project keys to match GitHub repository
- Add comprehensive build and test pipeline"

git push origin main
```

### 2. **Verificar Execução**
- Acesse: Actions tab no GitHub
- Verifique se workflows estão executando
- SonarCloud deve receber coverage data

## 🎯 **Resultados Esperados:**

### ✅ **GitHub Actions**
- ✅ Tests executando com coverage
- ✅ SonarCloud scan funcionando
- ✅ Build artifacts gerados

### ✅ **SonarCloud**
- ✅ Project importado corretamente
- ✅ Coverage metrics exibidos
- ✅ Quality gate funcionando

## 🐛 **Troubleshooting Comum:**

### **Error: "SONAR_TOKEN not found"**
```bash
Solução: Adicionar SONAR_TOKEN nos GitHub Secrets
```

### **Error: "Project not found"**
```bash
Solução: Verificar project key e organization no SonarCloud
```

### **Error: "Coverage file not found"**
```bash
Solução: Verificar se go test está gerando coverage.out
```

### **Error: "Quality gate failed"**
```bash
Solução: Verificar métricas no SonarCloud dashboard
```

## 📱 **Verificação Rápida:**

1. **GitHub**: Actions tab → deve mostrar workflows rodando
2. **SonarCloud**: Dashboard → deve mostrar métricas
3. **Coverage**: SonarCloud → deve mostrar % de coverage
4. **Quality Gate**: SonarCloud → deve mostrar PASSED/FAILED

---

**🚀 Após essas correções, todo o pipeline CI/CD funcionará perfeitamente!**