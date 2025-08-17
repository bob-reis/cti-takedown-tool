#!/bin/bash

# Exemplos de uso da ferramenta CTI Takedown

echo "=== CTI Takedown Tool - Exemplos ==="
echo

# Compilar se necessário
if [ ! -f "./takedown" ]; then
    echo "Compilando projeto..."
    go build -o takedown cmd/takedown/main.go
    echo "✓ Compilação concluída"
    echo
fi

echo "1. Submetendo URL de phishing:"
./takedown -action=submit -ioc="https://fake-bank-security.com/verify" -tags="phishing,brand:ExampleBank,high"
echo

echo "2. Submetendo domínio de malware:"
./takedown -action=submit -ioc="malware-distribution.evil" -type=domain -tags="malware,critical"
echo

echo "3. Submetendo C2 infrastructure:"
./takedown -action=submit -ioc="c2-command.bad" -type=domain -tags="c2,critical"
echo

echo "4. Listando todos os casos:"
./takedown -action=list
echo

echo "=== Integração com Frontend ==="
echo "Para integrar com o botão do frontend, use:"
echo 'curl -X POST http://localhost:8080/takedown -d {"ioc":"dominio.com","tags":["phishing","brand:BankName"]}'
echo

echo "=== Modo Daemon ==="
echo "Para executar como daemon:"
echo "./takedown -daemon &"
echo

echo "=== Configuração ==="
echo "Configure SMTP em configs/smtp.yaml"
echo "Ajuste SLAs em configs/sla/default.yaml" 
echo "Customize templates em configs/templates/"