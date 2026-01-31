# 📋 Filter Categories - Quick Reference

Übersicht über die Filter-Kategorien im Admin UI.

---

## 🎯 Wie Kategorien funktionieren

Die Kategorien werden **automatisch erkannt** basierend auf dem **Replacement-Text** des Filters.

Zum Beispiel:
- Filter mit `[EMAIL]` → Kategorie: **PII**
- Filter mit `[***API_KEY***]` → Kategorie: **Security**
- Filter mit `[CREDIT_CARD]` → Kategorie: **Financial**

---

## 📊 Verfügbare Kategorien

### 🆔 **PII (Personal Identifiable Information)**

**Erkennungsmuster:**
- Replacements enthalten: `EMAIL`, `PHONE`, `SSN`, `TAX_ID`, `PASSPORT`, `DRIVER`

**Beispiele:**
```
\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b,[EMAIL],regex
\b\d{3}-\d{2}-\d{4}\b,[SSN],regex
\b0[1-9]\d{1,4}[-\s]?\d{3,8}\b,[PHONE],regex
```

**Was gefiltert wird:**
- Email-Adressen
- Telefonnummern (alle Formate)
- Sozialversicherungsnummern
- Steuernummern
- Führerschein-Nummern
- Reisepass-Nummern
- Personalausweis-Nummern

---

### 💳 **Financial Data**

**Erkennungsmuster:**
- Replacements enthalten: `CREDIT_CARD`, `CVV`, `IBAN`, `BANK`, `CRYPTO`

**Beispiele:**
```
\b\d{4}[-\s]?\d{4}[-\s]?\d{4}[-\s]?\d{4}\b,[CREDIT_CARD],regex
\bCVV:?\s*\d{3,4}\b,[CVV],regex
\b[A-Z]{2}\d{2}[A-Z0-9]{10,30}\b,[IBAN],regex
```

**Was gefiltert wird:**
- Kreditkartennummern
- CVV/CVC Codes
- IBAN Kontonummern
- BIC/SWIFT Codes
- Routing Numbers
- Kryptowährung-Adressen

---

### 🔐 **Security & Credentials**

**Erkennungsmuster:**
- Replacements enthalten: `API`, `KEY`, `TOKEN`, `PASSWORD`, `SECRET`, `AWS`, `GITHUB`, `JWT`

**Beispiele:**
```
\bapi[_-]?key[:\s=]+[a-zA-Z0-9_-]{20,}\b,[***API_KEY***],regex
\bAKIA[0-9A-Z]{16}\b,[***AWS_KEY***],regex
\bghp_[a-zA-Z0-9]{36}\b,[***GITHUB_TOKEN***],regex
password,[***PASSWORD***],word
```

**Was gefiltert wird:**
- API Keys (Generic, AWS, Google, Stripe, etc.)
- GitHub/GitLab Tokens
- JWT Tokens
- SSH Private Keys
- OAuth/Bearer Tokens
- Passwords
- Slack/Twilio/SendGrid Credentials

---

### 🗄️ **Technical Secrets**

**Erkennungsmuster:**
- Replacements enthalten: `DB`, `CONNECTION`, `IP`, `HOST`, `LOCALHOST`

**Beispiele:**
```
\bmongodb(\+srv)?://[^\s'"]+,[***DB_CONNECTION***],regex
\b10\.\d{1,3}\.\d{1,3}\.\d{1,3}\b,[INTERNAL_IP],regex
\b[a-z0-9-]+\.(internal|local|lan|corp)\b,[INTERNAL_HOST],regex
```

**Was gefiltert wird:**
- Database Connection Strings
- Private IP Addresses (10.x, 172.x, 192.168.x)
- Internal Hostnames
- Environment Variables (DB_PASSWORD, SECRET_KEY)
- Docker Credentials

---

### 🔒 **Confidential & Internal**

**Erkennungsmuster:**
- Replacements enthalten: `CONFIDENTIAL`, `REDACTED`, `CLASSIFIED`, `INTERNAL`, `PROPRIETARY`

**Beispiele:**
```
confidential,[CONFIDENTIAL],word
top secret,[CLASSIFIED],phrase
Project Phoenix,[INTERNAL_PROJECT],phrase
proprietary information,[PROPRIETARY],phrase
```

**Was gefiltert wird:**
- Confidential Keywords
- Internal Project Names
- Proprietary Information
- Trade Secrets
- NDA-relevante Begriffe
- Legal Privilege
- HR-Daten (Salary, Reviews)

---

## 🔍 Filter-Funktionen im Admin UI

### 1. **Nach Type filtern:**

```
Filter Type: [All Types ▼]
             Word
             Phrase
             Regex
```

- **Word:** Exakte Wort-Übereinstimmung (case-insensitive)
- **Phrase:** Multi-Wort-Übereinstimmung
- **Regex:** Pattern-Matching mit Regular Expressions

---

### 2. **Nach Kategorie filtern:**

```
Category: [All Categories ▼]
          🆔 PII (Personal Info)
          💳 Financial Data
          🔐 Security & Credentials
          🗄️ Technical Secrets
          🔒 Confidential
```

Zeigt nur Filter der ausgewählten Kategorie.

---

### 3. **Suche:**

```
Search: [pattern, replacement...]
```

Durchsucht:
- Pattern
- Replacement
- Description

Echtzeit-Filterung während du tippst.

---

### 4. **Sortierung:**

```
Sort By: [Priority (High to Low) ▼]
         ID (Ascending)
         Pattern (A-Z)
```

- **Priority:** Wichtigste zuerst (100 → 80)
- **ID:** Chronologisch nach Erstellung
- **Pattern:** Alphabetisch

---

## 💡 Tipps & Best Practices

### Filter kombinieren:

```
Type: Regex
Category: Security
Search: github
Sort By: Priority
```

Zeigt nur:
- ✅ Regex-Filter
- ✅ Security-Kategorie
- ✅ Mit "github" im Pattern
- ✅ Sortiert nach Priority

### Quick Filters:

**Alle API Keys sehen:**
```
Category: Security
Search: api
```

**Alle Kreditkarten-Filter:**
```
Category: Financial
Search: credit
```

**Alle Email-Filter:**
```
Category: PII
Search: email
```

**Alle hohe Priorität:**
```
Sort By: Priority (High to Low)
→ Zeigt Filter mit Priority 95-100 zuerst
```

---

## 🎨 Category Badges

Filter werden automatisch mit farbigen Type-Badges markiert:

- 🔵 **Word** - Blau
- 🟢 **Phrase** - Grün  
- 🟣 **Regex** - Lila

---

## 📊 Statistiken

Die Filter-Statistik zeigt:

```
┌─────────────┬─────────────┬─────────────┬─────────────┐
│ Total: 112  │ Enabled: 110│ Matches: 45 │ Cache: 120s │
└─────────────┴─────────────┴─────────────┴─────────────┘

Filters by Type: Word: 45 | Phrase: 35 | Regex: 32
```

**Showing 15 of 112 filters**
- Zeigt wie viele Filter aktuell gefiltert sind
- Aktive Filter werden hervorgehoben

---

## 🔄 Filter zurücksetzen

**Einzeln:**
```
• Search: "api" [Clear]
• Type: regex [Clear]
• Category: security [Clear]
```

**Alle auf einmal:**
- Setze alle Dropdowns auf "All"
- Lösche Suchtext

---

## 🚀 Workflows

### Workflow 1: Alle PII-Filter prüfen

1. Category: **PII**
2. Durchgehe die Liste
3. Teste verdächtige Filter
4. Editiere bei Bedarf

### Workflow 2: Alle Regex-Pattern optimieren

1. Type: **Regex**
2. Sort By: **Priority**
3. Teste jeden Filter
4. Optimiere Performance

### Workflow 3: Neue Kategorie hinzufügen

1. Erstelle Filter mit konsistentem Replacement
2. z.B. alle mit `[MEDICAL_...]`
3. Suche dann nach: `medical`
4. Alle erscheinen zusammen

---

## 📖 Siehe auch:

- **filter-templates/README.md** - Template-Übersicht
- **CONTENT_FILTERING.md** - API Dokumentation
- **BULK_IMPORT_GUIDE.md** - Import-Anleitung

---

**Viel Erfolg beim Organisieren deiner Filter!** 🎯
