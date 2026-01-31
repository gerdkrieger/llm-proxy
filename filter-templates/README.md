# 🔒 Enterprise Content Filter Templates

Vorkonfigurierte Filter-Templates für professionelle/firmenbezogene Anwendungsfälle.

---

## 📋 Übersicht

Diese Filter-Templates decken die wichtigsten **Datenschutz-, Sicherheits- und Compliance-Anforderungen** im Unternehmenskontext ab.

### 📦 Verfügbare Template-Dateien:

| Datei | Filter | Beschreibung |
|-------|--------|--------------|
| **enterprise-filters.csv** | ~100+ | Komplettes Enterprise-Set (alle Kategorien) |
| **pii-filters.csv** | ~15 | Personenbezogene Daten (PII) |
| **financial-filters.csv** | ~12 | Finanz- und Bankdaten |
| **security-filters.csv** | ~25 | API Keys, Passwords, Tokens |
| **technical-filters.csv** | ~15 | DB Connections, IPs, Hostnames |
| **compliance-filters.csv** | ~20 | Confidential, Legal, HR |
| **medical-filters.csv** | ~5 | HIPAA-relevante Gesundheitsdaten |

---

## 🎯 Filter-Kategorien

### 1. **PII (Personal Identifiable Information)** 🆔
**Priorität:** 95-100

- ✅ Email-Adressen
- ✅ Telefonnummern (US, DE, UK, International)
- ✅ Sozialversicherungsnummern (SSN)
- ✅ Steuernummern (Tax IDs)
- ✅ Führerschein-Nummern
- ✅ Reisepass-Nummern
- ✅ Personalausweis-Nummern

**Beispiel:**
```
john.doe@company.com → [EMAIL]
+49 123 456789 → [PHONE]
123-45-6789 → [SSN]
```

---

### 2. **Financial Data** 💳
**Priorität:** 95-100

- ✅ Kreditkartennummern (Visa, Mastercard, Amex, Discover)
- ✅ CVV/CVC Codes
- ✅ IBAN (Internationale Bankkonten)
- ✅ BIC/SWIFT Codes
- ✅ Kontonummern
- ✅ Routing Numbers
- ✅ Kryptowährung-Adressen (Bitcoin, Ethereum)

**Beispiel:**
```
4532 1234 5678 9010 → [CREDIT_CARD]
CVV: 123 → [CVV]
DE89370400440532013000 → [IBAN]
bc1qxy2kgdygjrsqtzq2n0yrf2493p83kkfjhx0wlh → [CRYPTO_ADDRESS]
```

---

### 3. **Security & Credentials** 🔐
**Priorität:** 98-100

**API Keys & Secrets:**
- ✅ Generic API Keys
- ✅ AWS Access Keys & Secrets
- ✅ Google API Keys
- ✅ GitHub Tokens (PAT, OAuth, Server)
- ✅ GitLab Tokens
- ✅ Slack Tokens
- ✅ Stripe Keys
- ✅ Twilio Credentials
- ✅ SendGrid Keys

**Authentication:**
- ✅ Passwords (Keywords & Patterns)
- ✅ JWT Tokens
- ✅ SSH Private Keys
- ✅ Bearer Tokens
- ✅ OAuth Access Tokens

**Beispiel:**
```
api_key: sk-abc123def456 → [***API_KEY***]
AKIAIOSFODNN7EXAMPLE → [***AWS_KEY***]
AIzaSyD-example123 → [***GOOGLE_API_KEY***]
ghp_1234567890abcdefghijklmnopqrstuvwxyz → [***GITHUB_TOKEN***]
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9... → [***JWT_TOKEN***]
```

---

### 4. **Database & Technical Secrets** 🗄️
**Priorität:** 95-98

- ✅ MongoDB Connection Strings
- ✅ PostgreSQL Connections
- ✅ MySQL Connections
- ✅ Redis Connections
- ✅ Credentials in URLs
- ✅ Private IP Addresses (10.x, 172.x, 192.168.x)
- ✅ Internal Hostnames (.internal, .local, .corp)
- ✅ Environment Variables (DB_PASSWORD, SECRET_KEY)

**Beispiel:**
```
mongodb://user:pass@host:27017/db → [***DB_CONNECTION***]
postgres://admin:secret@10.0.0.1:5432/prod → [***DB_CONNECTION***]
10.0.1.50 → [INTERNAL_IP]
server01.internal → [INTERNAL_HOST]
DB_PASSWORD=supersecret → [***DB_PASSWORD***]
```

---

### 5. **Confidential & Internal** 🔒
**Priorität:** 85-95

**Keywords:**
- ✅ confidential, vertraulich, geheim, secret
- ✅ top secret, streng geheim
- ✅ proprietary information
- ✅ trade secret
- ✅ NDA, not for distribution

**Internal Projects:**
- ✅ Project Codenames (anpassbar)
- ✅ Internal Operations

**HR/Legal:**
- ✅ salary, gehalt
- ✅ performance review
- ✅ attorney-client privilege
- ✅ pending litigation

**Beispiel:**
```
This is confidential → This is [CONFIDENTIAL]
Project Phoenix details → [INTERNAL_PROJECT] details
salary: $120,000 → [SALARY_INFO]
```

---

### 6. **Competitor & Business Intelligence** 🏢
**Priorität:** 80-90

- ✅ Competitor Names (anpassbar)
- ✅ acquisition target, merger talks
- ✅ earnings report, quarterly results
- ✅ layoff, cost cutting

---

### 7. **Medical & Health Data (HIPAA)** 🏥
**Priorität:** 98-100

- ✅ Medical Record Numbers (MRN)
- ✅ diagnosed with, medical condition
- ✅ prescription for
- ✅ Health Insurance IDs

---

### 8. **Additional Security Patterns** 🛡️
**Priorität:** 90-95

- ✅ Encryption Keys (Hex)
- ✅ Software License Keys
- ✅ UUIDs/GUIDs
- ✅ Session Tokens
- ✅ CSRF Tokens

---

## 🚀 Usage

### 1. **Komplettes Set importieren:**

```bash
# Via API
curl -X POST http://localhost:8080/admin/filters/bulk-import \
  -H "Content-Type: application/json" \
  -H "X-Admin-API-Key: YOUR_KEY" \
  -d @filter-templates/enterprise-filters.csv

# Via Admin UI
1. Öffne http://localhost:5173
2. Gehe zu "Filters"
3. Klicke "Bulk Import"
4. Lade enterprise-filters.csv hoch
```

### 2. **Nur bestimmte Kategorien:**

```bash
# Nur PII-Filter
cat filter-templates/pii-filters.csv

# Nur Security-Filter
cat filter-templates/security-filters.csv
```

### 3. **Eigene Anpassungen:**

Editiere die CSV-Dateien und passe an:
- **Prioritäten** (höher = wichtiger)
- **Patterns** (Regex-Muster)
- **Replacements** (Was angezeigt wird)
- **Beschreibungen**

---

## ⚙️ Anpassung an deine Firma

### Competitor Names:
```csv
# Ersetze mit echten Competitor-Namen
YourCompetitor,[COMPETITOR],word,85,Main competitor,false,true
RivalCompany,[COMPETITOR],word,85,Industry rival,false,true
```

### Internal Project Names:
```csv
# Ersetze mit echten Projekt-Codenamen
Project Alpha,[INTERNAL_PROJECT],phrase,90,Internal codename,false,true
Project Beta,[INTERNAL_PROJECT],phrase,90,Internal codename,false,true
```

### Internal Domains:
```csv
# Füge deine internen Domains hinzu
\.yourcompany\.internal\b,[INTERNAL_HOST],regex,95,Internal domain,false,true
\.corp\.yourcompany\.com\b,[INTERNAL_HOST],regex,95,Corporate domain,false,true
```

---

## 🎯 Best Practices

### 1. **Prioritäten richtig setzen:**
```
100 = Kritisch (Passwords, SSN, Credit Cards)
95-98 = Sehr wichtig (API Keys, IBAN, Medical Data)
90-94 = Wichtig (Internal IPs, Confidential)
85-89 = Normal (Competitor Names, UUIDs)
80-84 = Niedrig (Generic Keywords)
```

### 2. **Regex-Pattern testen:**
Nutze den Test-Button im Admin UI bevor du Filter aktivierst!

### 3. **Case Sensitivity:**
- `false` = case-insensitive (empfohlen für Keywords)
- `true` = case-sensitive (für spezifische Patterns wie API Keys)

### 4. **Schrittweise einführen:**
Starte mit einer Kategorie (z.B. Security), teste, dann erweitere.

---

## 📊 Performance

- **Filter-Anzahl:** ~100 Filter
- **Overhead:** <10ms pro Request (mit Caching)
- **Cache-TTL:** 5 Minuten (einstellbar)
- **Match-Recording:** Asynchron (keine Blocking)

---

## 🔧 Troubleshooting

### "Filter matcht zu viel"
→ Pattern zu generisch, verfeinere Regex

### "Filter matcht zu wenig"
→ Pattern zu spezifisch, mache Regex flexibler

### "Performance-Probleme"
→ Prüfe Regex-Komplexität, ReDoS-Protection ist aktiv

### "False Positives"
→ Nutze negative lookaheads in Regex oder erhöhe Spezifität

---

## 📖 Weitere Ressourcen

- **CONTENT_FILTERING.md** - API Dokumentation
- **BULK_IMPORT_GUIDE.md** - Import-Anleitung
- **TESTING_REPORT.md** - Test-Ergebnisse

---

## ⚖️ Compliance

Diese Templates helfen bei:
- ✅ **GDPR** (EU Datenschutz)
- ✅ **HIPAA** (US Healthcare)
- ✅ **PCI DSS** (Payment Card Security)
- ✅ **SOC 2** (Security Controls)
- ✅ **ISO 27001** (Information Security)

---

## 🆘 Support

Bei Fragen oder Problemen:
1. Prüfe die Logs: `tail -f /tmp/llm-proxy.log`
2. Teste Filter einzeln im Admin UI
3. Prüfe Filter Statistics Dashboard

---

**Viel Erfolg beim Absichern deiner LLM-Kommunikation!** 🔒🚀
