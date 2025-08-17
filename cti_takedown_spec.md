
# Especificação do Módulo de Takedown — Plataforma de CTI

> **Objetivo:** Orquestrar e registrar requisições de remoção/suspensão de domínios/URLs maliciosos com evidência sólida, trilhas de auditoria e escalonamento automatizado, reduzindo MTTR sem gerar falsos positivos.

---

## 1) Visão Geral de Arquitetura

**Componentes**
- **Ingest:** Entrada de IOCs (domínios/URLs/IPs/ASN), telemetria e alertas (crawlers, detecções internas, feeds de terceiros).
- **Enriquecimento:** DNS ativo/passivo, RDAP/WHOIS (registrar/abuse), IP→ASN (hosting provável), certificado TLS, HTTP headers, screenshot/HAR.
- **Decisão (Policy Engine):** Regras por tipo (phishing, malware delivery, C2, typosquatting/brand). Scoring e severidade.
- **Ação (Connectors):** Registrar, Registry/ccTLD, Hosting/ISP, CDN, Buscadores/Warning lists, Blocklists, CERTs.
- **Orquestração:** State machine, filas, SLAs, *human-in-the-loop*.
- **Auditoria:** Evidências, comunicações, prazos, outcomes, cadeia de custódia.

**Princípios**
- **Plugável:** conectores por registrar/hosting/CDN via drivers.
- **Idempotente:** reenvios não duplicam casos.
- **Seguro:** defang de IOCs; coleta em sandbox isolado; logging verificado.

---

## 2) Fluxo Macro (Estados)

1. **Discovered** → 2. **Triage** (score/risk) → 3. **Evidence Pack** (coleta) →  
4. **Route** (registrar/hosting/cdn/ccTLD/buscadores/blocklists/CERT) →  
5. **Submit** (template certo) → 6. **Ack/Case ID** →  
7. **Follow-up** (SLA, lembretes, reenvio) → 8. **Outcome** (Suspenso / Removido / Negado / Escalar) →  
9. **Close** (lições, métricas).

**SLAs sugeridos**
- Registrar/Hosting: 24–72h para 1º retorno; relembrar a cada 48h; escalar em 5 dias úteis.
- CDN: 24–48h (normalmente encaminham ao origin).
- Buscadores/Warnings: repetir submissão em 24–72h se ainda ativo.

---

## 3) Evidence Pack (Checklist Mínimo)

- URLs **defangadas** (`hxxp://`, `example[.]com`) + caminho e parâmetros.
- **Screenshots** + **HAR** com timestamps (UTC e local) e **hash** dos arquivos.
- **DNS**: A/AAAA/CNAME/MX/TXT/SPF/DMARC/DKIM, TTLs, cadeias de CNAME.
- **IP/ASN** (hosting provável) + geolocalização de rede.
- **HTTP headers** (server, redirect chain) e **status**.
- **TLS**: emissor, CN/SAN, validade.
- Indicadores de **phishing** (marca alvo, formulários, endpoints), **malware** (hash de amostras) ou **C2** (padrões de beaconing).
- Links de reputação (URLhaus/Spamhaus/etc.) — **não depender exclusivamente** deles.
- **Impacto/Risco** + **prioridade** + **ação solicitada** (suspender domínio, remover conteúdo, bloquear NS, etc.).

**Observação:** Sempre opere em ambiente isolado. Nunca clique sem defang em rede corporativa.

---

## 4) Descobrir Quem Acionar (Automatizado)

1. **RDAP do domínio** → extrair `registrar` e contatos de **abuse** (e-mail/telefone/form).
2. **Resolver DNS** → IP → **ASN** (hosting provável) → contato **abuse** do provedor.
3. **CDN presente?** Reportar **CDN** e **hosting de origem** (quando identificável).
4. **ccTLD .BR**:
   - Conflito de nome/brand (typosquatting): **SACI-Adm** (WIPO/ABPI/Arbitragem).
   - Conteúdo malicioso: acionar **hosting/ISP**; para coordenação/incidentes, envolver **CERT.br**.

---

## 5) Playbooks por Cenário

### 5.1 Registrar (gTLD: .com/.net/.org…)
**Quando:** phishing/malware/C2 hospedado em domínio/subdomínio + violação de ToS/contratos.  
**Passos:**
1. RDAP → obter **registrar** e **abuse**.
2. Montar **Evidence Pack** (Seção 3).
3. **Enviar** via e-mail/formulário do registrar (solicitar **suspensão/bloqueio** do domínio ou do label/subdomínio responsável).
4. **Follow-up**: aguardar ACK/ID do caso; relembrar em 48h; **escalar** conforme Seção 7.

### 5.2 Hosting/ISP — *Remoção de Conteúdo*
**Quando:** conteúdo está em um servidor/VM/conta de cliente.  
**Passos:**
1. IP → ASN → contato **abuse** (muitas redes aceitam `abuse@domínio`).
2. Enviar Evidence Pack com foco em **conteúdo** (caminhos, payloads, logs, HAR).
3. Solicitar **remoção/isolamento** e notificação ao cliente.

### 5.3 CDN (ex.: Cloudflare/Akamai/Fastly)
- Usar formulário de abuso da CDN e, **em paralelo**, notificar o **hosting de origem**.
- A CDN geralmente **encaminha** ao origin; por isso, o contato direto com o host acelera.

### 5.4 ccTLD .BR
- **Disputa de nome/typosquatting/marca:** **SACI-Adm**.
- **Conteúdo malicioso:** acionar **hosting/ISP**; considerar **CERT.br** para coordenação.
- Registro.br não remove conteúdo do site; foca-se no nome de domínio e questões contratuais.

### 5.5 Buscadores / Warning Lists (mitigação de alcance)
- **Google**: reportar phishing/malware (Safe Browsing/Search).
- **Microsoft SmartScreen**: reportar site inseguro.
> Útil para reduzir tráfego enquanto a origem não foi mitigada. **Não substitui** o takedown na infraestrutura.

### 5.6 Blocklists / Threat-Sharing
- **Spamhaus DBL**: reputação de domínios.
- **abuse.ch / URLhaus**: submissão de URLs de malware (há API para automação).
- **APWG**: reporte de phishing (compartilhamento com ecossistema).

### 5.7 E-mail/Providers
- Uso de botões “report phishing” em clientes/serviços ajuda a derrubar campanhas.
- Encaminhar amostras padroniza combate em múltiplos provedores.

---

## 6) Domínios com Privacy/Proxy

**Linha mestra:** Privacy/Proxy **não impede** takedown. Você **não precisa** identificar o titular para mitigar o abuso.

**Fluxo:**
1. **RDAP** → coletar **registrar** e contato de **abuse**; se houver entidade privacy/proxy, usar o **relay** (o provedor encaminha sua notificação ao titular).
2. **Enviar ao registrar (abuse)** com Evidence Pack e **pedir**:
   - (a) Ação por violação de ToS/DNS Abuse (suspensão/bloqueio), e
   - (b) **Relay** da notificação ao registrant.
3. **Em paralelo**, acionar **hosting/ASN** (privacy não cobre infraestrutura).

**“Reveal” (desanonimização):**
- Solicitar **apenas** quando houver **base legal** clara (UDRP/URS, ordem judicial, investigação formal). Muitos provedores **apenas encaminham** (relay) sem revelar dados.

---

## 7) Escalonamento (ICANN vs. Outros)

**ICANN (gTLDs) — quando usar:**
- Registrar **sem canal de abuse** público, ou que **não responde**/ignora evidências.
- **Dados RDDS/RDAP** grosseiramente imprecisos e não corrigidos após reporte.
> A ICANN trata **conformidade contratual** do registrar. **Não derruba** sites/casos individuais.

**Não é ICANN:**
- **Conteúdo malicioso:** Hosting/ASN/CDN + Buscadores/Blocklists.
- **Marca/typosquatting:** **UDRP/URS** (gTLD) | **SACI-Adm** (.BR).

**Regra prática:**
1) Notificar **Hosting/CDN/Registrar** com Evidence Pack.  
2) Fazer **2 follow-ups** (48–72h).  
3) **gTLD** sem resposta/sem canal → **ICANN Contractual Compliance** (anexar tickets e prazos).  
4) **Marca:** UDRP/URS (gTLD) | **SACI-Adm** (.BR).

---

## 8) Dados & Schemas (Exemplos)

### 8.1 IOC (entrada)
```json
{
  "indicator_id": "ioc-2025-08-16-001",
  "type": "url",
  "value": "hxxps://login-acme-security[.]com/verify",
  "first_seen": "2025-08-16T12:03:00Z",
  "source": "phishing-detector",
  "tags": ["phishing", "brand:AcmeBank", "high"]
}
```

### 8.2 Evidence Pack (armazenamento)
```json
{
  "evidence_id": "ev-001",
  "ioc": "ioc-2025-08-16-001",
  "screenshots": ["s3://.../shot1.png"],
  "har": "s3://.../sess.har",
  "dns": {"A": ["203.0.113.10"], "MX": [], "TXT": ["v=spf1 ..."], "DMARC": "p=none"},
  "http": {"headers": {"server": "nginx"}, "status": 200, "chain": ["302 -> /auth"]},
  "tls": {"issuer": "R3", "cn": "*.example.com", "not_after": "2025-12-02"},
  "intel_refs": ["urlhaus:123456", "spamhaus-dbl:suspect"],
  "risk": {"score": 92, "rationale": "credential harvest + brand kit + live victims"}
}
```

### 8.3 AbuseContact (normalizado por RDAP)
```json
{
  "domain": "example.com",
  "registrar": {"name": "GoDaddy.com, LLC", "iana_id": 146},
  "abuse": {"email": "abuse@godaddy.com", "webform": "https://.../AbuseReport"},
  "hosting": {"asn": 64500, "name": "ExampleHost", "abuse": "abuse@examplehost.com"},
  "cdn": {"name": "Cloudflare", "webform": "https://.../reporting-abuse/"}
}
```

### 8.4 TakedownRequest (estado)
```json
{
  "case_id": "tdk-2025-08-16-777",
  "target": {"type": "registrar", "entity": "GoDaddy.com, LLC"},
  "evidence_id": "ev-001",
  "requested_action": "suspend_domain",
  "status": "submitted",
  "sla": {"first_response_hours": 48, "escalate_after_hours": 120},
  "history": [
    {"t": "2025-08-16T13:00Z", "event": "submitted", "channel": "webform", "ref": "GD-CASE-98765"}
  ]
}
```

---

## 9) Orquestração & Operação

- **State machine:** `discovered → triage → evidence → route → submitted → acked → followup → outcome → closed`.
- **Retentativas e lembretes:** tarefa a cada 24h; reencaminhar com histórico do ticket/ID de caso.
- **Escalação:** sem resposta → ICANN (gTLD) para **conformidade**, ou jurídico (UDRP/URS/SACI) para **marca**.
- **KPIs:** MTTA/MTTR, taxa de sucesso por tipo (registrar/host/CDN), % de reenvios, top TLDs/ASNs, aging.

---

## 10) Políticas de Decisão (Exemplos)

- **Phishing ativo:** Hosting + Registrar + Google/Microsoft warnings + Blocklists.
- **Malware delivery:** Hosting (remover payload) + URLhaus + Blocklists.
- **C2:** Hosting/ASN e, se aplicável, CERT nacional para coordenação.
- **Typosquatting/brand:** UDRP/URS (gTLD) ou **SACI-Adm** (.BR).

---

## 11) Segurança Operacional

- Nunca acessar IOCs a partir de rede corporativa; usar sandbox/headless isolado.
- Defangar IOCs em toda comunicação externa.
- Assinar e-mails com PGP da equipe (opcional, útil para alguns abuse desks).
- Logar tudo (headers, respostas, números de caso), com **carimbo de tempo confiável**.

---

## 12) Métricas & Logs

- **Por alvo:** registrar/host/cdn/buscador/blocklist.
- **Taxa de sucesso** por tipo (phishing/malware/C2/brand).
- **Tempo até ACK** e **tempo até mitigação**.
- **TLDs/ASNs mais recorrentes**; **hotlist** para priorização.
- **Reincidência** por cliente/ASN.

---

## 13) Templates (Copiar/Colar)

### 13.1 Registrar (PT)
```
Assunto: [Urgente] Solicitação de suspensão de domínio por phishing — {domain}

Prezados,

Identificamos atividade de phishing no domínio {domain}, registrado por {registrar}.
Solicitamos SUSPENSÃO conforme violação de ToS/DNS Abuse.

Resumo:
- Domínio/URLs: {urls_defanged}
- Marca alvo: {brand}
- Evidências: screenshots, HAR, DNS, HTTP headers, TLS, indicadores de coleta de credenciais.
- Impacto: {impact}
- Primeiro visto (UTC): {first_seen_utc}

Links de evidência (acesso restrito): {evidence_links}

Contato técnico para esclarecimentos: {name, email, telefone}.

Agradecemos confirmação de recebimento e número de caso.

Atenciosamente,
{assinatura}
```

### 13.2 Hosting/ISP (EN)
```
Subject: [Abuse] Malware/Phishing content hosted on your network — {ip} / ASN {asn}

Hello Abuse Team,

We detected malicious content being served from {ip} (ASN {asn}) associated with domain {domain}.
Please REMOVE the content and notify the customer.

Evidence:
- URLs (defanged): {urls}
- Screenshots & HAR: {links}
- DNS & HTTP details: {summary}
- First seen (UTC): {time}

Kindly provide a ticket ID and status.

Regards,
{signature}
```

### 13.3 Cloudflare/CDN (nota)
> Usar o formulário de abuso da CDN; normalmente o provedor **notifica o origin**. Em paralelo, **contate o host**.

### 13.4 Relay para Serviço de Privacy/Proxy (EN)
```
Subject: [Abuse Relay Request] Phishing/Malware hosted at {domain} — Please forward to registrant

Hello Privacy/Proxy Team,

We detected {phishing|malware|C2} on {domain}. Please FORWARD this notice to the registrant.
We also notified the registrar abuse desk.

Evidence (links, defanged):
- URLs: {list}
- Screenshots & HAR: {links}
- DNS/HTTP/TLS summary: {summary}
- First seen (UTC): {time}

Requested actions:
- Remove/suspend malicious content
- Registrar: suspend domain per ToS/DNS Abuse obligations

Thank you for relaying this notice and confirming receipt.

Regards,
{signature}
```

### 13.5 CERT nacional (PT)
```
Assunto: [Coordenação de Incidente] {phishing|malware|C2} — {domain} / ASN {asn}

Prezados,

Estamos coordenando takedown do alvo {domain} ({ip}, ASN {asn}) com indícios de {tipo}.
Solicitamos apoio na coordenação com a rede de origem caso necessário.

Resumo de evidências (links restritos): {evidence_links}
Primeiro visto (UTC): {time}
Impacto: {impact}

Contato técnico: {nome, email, telefone}.

Atenciosamente,
{assinatura}
```

---

## 14) Anexos Operacionais

### 14.1 Estrutura de Repositório Sugerida
```
/takedown
  /docs
    spec.md
    runbook.md
  /connectors
    registrar/
    hosting/
    cdn/
    search/
    blocklists/
  /policies
    routing.yaml
    sla.yaml
  /samples
    templates/
    evidence/
  /infra
    state-machine.md
    queueing.md
```

### 14.2 `sla.yaml` (exemplo)
```yaml
registrar:
  first_response_hours: 48
  escalate_after_hours: 120
hosting:
  first_response_hours: 48
  escalate_after_hours: 96
cdn:
  first_response_hours: 24
  escalate_after_hours: 72
search_warnings:
  retry_hours: 24
```

### 14.3 `routing.yaml` (exemplo)
```yaml
rules:
  - match: ["phishing", "brand:*"]
    actions: ["registrar", "hosting", "search_warnings", "blocklists"]
  - match: ["malware"]
    actions: ["hosting", "blocklists"]
  - match: ["c2"]
    actions: ["hosting", "cert"]
```

### 14.4 Definições de DoD (Definition of Done) por Ação
- **Registrar:** ACK com ID de caso e status atualizado; ou 2 follow-ups concluídos; ou escalado à ICANN (gTLD).
- **Hosting:** Confirmação de remoção/bloqueio; ou evidência de indisponibilidade do conteúdo.
- **CDN:** Caso encaminhado + origem notificada/mitigada.
- **Buscadores/Warnings:** Status de revisão recebido.
- **Blocklists:** IOC submetido e aceito.

---

## 15) Minha Opinião (Resumo Prático)

Automatize **descoberta de contatos via RDAP**, padronize **Evidence Packs**, modele uma **máquina de SLA** (reenvio/escalação) e trate **privacy** como detalhe operacional (use **relay** por padrão, **reveal** só com base legal). Para .BR, foque em **hosting/ISP** para conteúdo e **SACI-Adm** para marca. Enquanto isso, use **warnings** e **blocklists** para reduzir o impacto.
