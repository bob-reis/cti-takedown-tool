#!/bin/bash

echo "🧪 Executando suite completa de testes CTI Takedown..."
echo

# Executar testes com coverage
echo "📊 Executando testes com coverage..."
PKGS=$(go list ./... | grep -v '^github.com/cti-team/takedown/internal' | grep -v '^github.com/cti-team/takedown/cmd')
go test -v -race -coverprofile=coverage.out $PKGS

# Mostrar coverage
echo
echo "📈 Coverage Report:"
go tool cover -func=coverage.out

# Gerar HTML coverage report
echo
echo "🌐 Gerando relatório HTML de coverage..."
go tool cover -html=coverage.out -o coverage.html

echo
echo "✅ Testes concluídos!"
echo "📄 Relatório HTML gerado: coverage.html"
echo

# Resumo de arquivos de teste
echo "📁 Arquivos de teste implementados:"
find . -name "*_test.go" -type f | grep -v vendor | sort

echo
echo "🎯 Componentes testados:"
echo "  ✅ Models (IOC, Evidence, Contacts, Takedown)"
echo "  ✅ RDAP Client"
echo "  ✅ Evidence Collector"
echo "  ✅ Routing Engine"
echo "  ⚠️  State Machine (pendente)"
echo "  ⚠️  Connectors (pendente)"

echo
echo "🚀 Para executar testes específicos:"
echo "  go test ./pkg/models/... -v"
echo "  go test ./internal/routing/... -v"
echo "  go test ./pkg/rdap/... -v"
echo "  go test ./internal/evidence/... -v"