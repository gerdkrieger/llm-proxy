# Content Filter Management - Umfassende Anleitung

**Version**: 1.0  
**Zuletzt aktualisiert**: 30. Januar 2026  
**Status**: Production Ready

---

## 📋 Inhaltsverzeichnis

1. [Quick Start](#quick-start)
2. [Filter erstellen](#filter-erstellen)
3. [Filter bearbeiten](#filter-bearbeiten)
4. [Bulk Import](#bulk-import)
5. [Template Categories](#template-categories)
6. [Common Use Cases](#common-use-cases)
7. [Testing](#testing)
8. [Troubleshooting](#troubleshooting)
9. [API Reference](#api-reference)

---

## 🚀 Quick Start

### Zugriff auf das Admin UI

```bash
# Admin UI öffnen
URL: http://localhost:5173

# Navigation
1. Einloggen mit Admin API Key
2. Sidebar: "🔒 Filters" klicken
3. Filter Management Dashboard öffnet sich
```

### Erster Filter in 60 Sekunden

```
1. Click "➕ Add Filter"
2. Pattern: test@example.com
3. Replacement: [EMAIL] (aus Dropdown wählen)
4. Type: regex
5. Priority: 100
6. Description: Email addresses
7. Click "Create" ✅
```

**Fertig!** Der Filter ist aktiv und filtert alle Chat-Requests.

---

## 📝 Filter erstellen

### Mit Replacement Template (Empfohlen)

**Vorteile:**
- ✅ Keine Tippfehler
- ✅ Standardisierte Werte
- ✅ 60+ vorgefertigte Templates
- ✅ Nach Kategorien organisiert

**Schritte:**

1. **Filter-Modal öffnen**
   - Click "➕ Add Filter" Button
   
2. **Pattern eingeben**
   ```
   Beispiele:
   - Email: [a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}
   - AWS Key: AKIA[0-9A-Z]{16}
   - Credit Card: \b\d{4}[\s-]?\d{4}[\s-]?\d{4}[\s-]?\d{4}\b
   ```

3. **Replacement Template wählen**
   - Click Dropdown "Replacement Template"
   - Navigiere zu passender Kategorie:
     - 🆔 PII - Persönliche Daten
     - 💳 Financial - Finanzdaten
     - 🔐 Security - API Keys, Tokens
     - 🗄️ Technical - DB, IPs
     - 🔒 Confidential - Vertrauliches
   - Wähle Template (z.B. `[EMAIL] - Email Address`)
   - Preview erscheint: "Will replace matches with: [EMAIL]"

4. **Filter Type wählen**
   - **word**: Einfache Wörter (z.B. `password`)
   - **phrase**: Multi-Wort (z.B. `social security number`)
   - **regex**: Komplexe Patterns (z.B. Email-Regex)

5. **Priority setzen**
   ```
   Empfohlene Werte:
   100 = Kritisch (Credentials, Keys)
   90  = Hoch (PII, Finanzdaten)
   80  = Mittel (Technical Secrets)
   70  = Niedrig (Confidential Terms)
   ```

6. **Description hinzufügen**
   - Kurze Beschreibung des Filters
   - Hilft bei Wartung und Debugging

7. **Optionen setzen**
   - ☑️ Case Sensitive (optional)
   - ☑️ Enabled (Standard: an)

8. **Speichern**
   - Click "Create"
   - Filter erscheint in der Liste
   - Ist sofort aktiv

### Mit Custom Replacement

**Wann verwenden:**
- Firmenspezifische Terme
- Einzigartige Replacement-Werte
- Temporäre Filter

**Schritte:**

1-2. Wie oben

3. **Custom Replacement**
   - Dropdown auf "✏️ Custom (type your own)" lassen
   - Text-Input erscheint
   - Eigenen Wert eingeben: z.B. `[COMPANY_SECRET]`

4-8. Wie oben

---

## ✏️ Filter bearbeiten

### Filter mit Template Replacement

**Der Editor erkennt automatisch**, ob ein Filter ein Template verwendet:

1. **Filter finden und öffnen**
   - In Liste: Click "Edit" Button
   - Oder: Search verwenden, dann Edit

2. **Template wird erkannt**
   - Dropdown zeigt aktuelles Template (z.B. `[EMAIL] - Email Address`)
   - Preview zeigt aktuellen Wert
   - ✅ Smart Detection funktioniert automatisch

3. **Template ändern**
   - Option 1: Anderes Template wählen
   - Option 2: Zu "Custom" wechseln für eigenen Wert

4. **Speichern**
   - Click "Update"
   - Filter wird sofort aktualisiert
   - Cache wird automatisch refreshed

### Filter mit Custom Replacement

1. **Filter öffnen**
   - Click "Edit" auf Filter mit custom value

2. **Custom Mode erkannt**
   - Dropdown zeigt "✏️ Custom"
   - Text-Input zeigt aktuellen Wert
   - Kann bearbeitet werden

3. **Ändern**
   - Option 1: Custom Wert editieren
   - Option 2: Template aus Dropdown wählen

4. **Speichern**
   - Click "Update"

---

## 📦 Bulk Import

### Vorgefertigte Enterprise Templates importieren

**100+ Production-Ready Filters verfügbar!**

**Dateien:**
- `filter-templates/enterprise-filters.csv` - Komplette Suite
- Oder einzelne CSV-Dateien erstellen

**Import-Schritte:**

1. **Bulk Import öffnen**
   - Click "📦 Bulk Import" Button
   - Modal öffnet sich

2. **CSV-Datei öffnen**
   - Terminal: `cat filter-templates/enterprise-filters.csv`
   - Oder: In Editor öffnen

3. **Content kopieren**
   - Alle Zeilen markieren
   - Copy (Ctrl+C)

4. **In Textarea einfügen**
   - Paste (Ctrl+V) in großes Textfeld
   - Format wird automatisch erkannt

5. **Import starten**
   - Click "Import" Button
   - Progress wird angezeigt

6. **Ergebnis prüfen**
   ```
   ✓ Success: 95 filters imported
   ✗ Failed: 5 filters
   
   Failed Items:
   • Line 12: Invalid regex pattern
   • Line 34: Duplicate pattern
   ...
   ```

7. **Fehler beheben** (falls welche)
   - Fehlerhafte Zeilen korrigieren
   - Erneut importieren

### CSV Format

```csv
pattern,replacement,type,priority,description,case_sensitive,enabled
test@example.com,[EMAIL],regex,100,Email addresses,false,true
AKIA[0-9A-Z]{16},[***AWS_KEY***],regex,100,AWS Access Keys,false,true
password,[***PASSWORD***],word,100,Password keyword,false,true
```

**Felder:**
- `pattern` - Pattern zum Matchen (required)
- `replacement` - Replacement value (required)
- `type` - word|phrase|regex (required)
- `priority` - 0-100 (default: 100)
- `description` - Beschreibung (optional)
- `case_sensitive` - true|false (default: false)
- `enabled` - true|false (default: true)

---

## 🎯 Template Categories

### 🆔 PII - Personal Identifiable Information (8)

| Template | Verwendung | Beispiel Pattern |
|----------|------------|------------------|
| `[EMAIL]` | Email-Adressen | `[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}` |
| `[PHONE]` | Telefonnummern | `\+?[1-9]\d{1,14}` |
| `[SSN]` | Social Security Number | `\d{3}-\d{2}-\d{4}` |
| `[TAX_ID]` | Steuer-ID | `\d{2}-\d{7}` |
| `[PASSPORT]` | Reisepass-Nummer | `[A-Z]{1,2}\d{6,9}` |
| `[DRIVER_LICENSE]` | Führerschein | `[A-Z]\d{7,8}` |
| `[NATIONAL_ID]` | Personalausweis | `\d{9}` |
| `[MRN]` | Medical Record Number | `MRN\d{7}` |

### 💳 Financial Data (7)

| Template | Verwendung | Beispiel Pattern |
|----------|------------|------------------|
| `[CREDIT_CARD]` | Kreditkarten | `\b\d{4}[\s-]?\d{4}[\s-]?\d{4}[\s-]?\d{4}\b` |
| `[CVV]` | CVV/CVC Code | `\b\d{3,4}\b` |
| `[IBAN]` | IBAN | `[A-Z]{2}\d{2}[A-Z0-9]+` |
| `[BIC]` | BIC/SWIFT | `[A-Z]{6}[A-Z0-9]{2}([A-Z0-9]{3})?` |
| `[BANK_ACCOUNT]` | Kontonummer | `\d{8,12}` |
| `[ROUTING_NUMBER]` | Routing Number | `\d{9}` |
| `[CRYPTO_ADDRESS]` | Crypto Wallet | `[13][a-km-zA-HJ-NP-Z1-9]{25,34}` |

### 🔐 Security & Credentials (16)

| Template | Verwendung | Beispiel Pattern |
|----------|------------|------------------|
| `[***API_KEY***]` | Generic API Key | `[A-Za-z0-9_-]{32,}` |
| `[***AWS_KEY***]` | AWS Access Key | `AKIA[0-9A-Z]{16}` |
| `[***AWS_SECRET***]` | AWS Secret Key | `[A-Za-z0-9/+=]{40}` |
| `[***GOOGLE_API_KEY***]` | Google API Key | `AIza[0-9A-Za-z-_]{35}` |
| `[***GITHUB_TOKEN***]` | GitHub Token | `gh[pousr]_[A-Za-z0-9]{36}` |
| `[***GITLAB_TOKEN***]` | GitLab Token | `glpat-[A-Za-z0-9_-]{20}` |
| `[***JWT_TOKEN***]` | JWT Token | `eyJ[A-Za-z0-9-_=]+\.[A-Za-z0-9-_=]+\.[A-Za-z0-9-_.+/=]*` |
| `[***SSH_PRIVATE_KEY***]` | SSH Private Key | `-----BEGIN (RSA|DSA|EC|OPENSSH) PRIVATE KEY-----` |
| `[***BEARER_TOKEN***]` | Bearer Token | `Bearer [A-Za-z0-9_-]{20,}` |
| `[***PASSWORD***]` | Password Keyword | `password` (word) |
| `[***SLACK_TOKEN***]` | Slack Token | `xox[baprs]-[0-9]{10,13}-[a-zA-Z0-9-]+` |
| `[***STRIPE_KEY***]` | Stripe Key | `sk_live_[0-9a-zA-Z]{24}` |

### 🗄️ Technical Secrets (9)

| Template | Verwendung | Beispiel Pattern |
|----------|------------|------------------|
| `[***DB_CONNECTION***]` | DB Connection String | `mongodb://.*` oder `postgres://.*` |
| `[***DB_PASSWORD***]` | DB Password | In Connection Strings |
| `[INTERNAL_IP]` | Internal IPs | `10\.\d{1,3}\.\d{1,3}\.\d{1,3}` |
| `[INTERNAL_HOST]` | Internal Hostnames | `.*\.internal\b` |
| `[LOCALHOST]` | Localhost | `localhost` |
| `[***SECRET_KEY***]` | Generic Secret | `secret_key.*=.*` |
| `[***ENCRYPTION_KEY***]` | Encryption Key | 256-bit hex strings |

### 🔒 Confidential (10)

| Template | Verwendung | Beispiel Pattern |
|----------|------------|------------------|
| `[CONFIDENTIAL]` | Keyword | `confidential` |
| `[REDACTED]` | Generic Redaction | Various |
| `[CLASSIFIED]` | Classified Info | `classified` |
| `[INTERNAL_PROJECT]` | Project Names | `Project [A-Z][a-z]+` |
| `[PROPRIETARY]` | Proprietary Info | `proprietary` |
| `[TRADE_SECRET]` | Trade Secrets | Firm-specific |
| `[SALARY_INFO]` | Salary Info | Salary patterns |
| `[COMPETITOR]` | Competitor Names | Specific names |

### 🛡️ Additional (4)

| Template | Verwendung | Beispiel Pattern |
|----------|------------|------------------|
| `[UUID]` | UUIDs | `[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}` |
| `[LICENSE_KEY]` | Software Licenses | Firm-specific |
| `[SESSION_TOKEN]` | Session Tokens | `sess_[A-Za-z0-9]{32}` |
| `[CSRF_TOKEN]` | CSRF Tokens | 32+ character tokens |

**Total: 54 vordefinierte Templates**

---

## 💡 Common Use Cases

### Use Case 1: Filter AWS Credentials
```
Pattern:      AKIA[0-9A-Z]{16}
Template:     [***AWS_KEY***]
Type:         regex
Priority:     100
Description:  AWS Access Key IDs
Case Sensitive: false
Enabled:      true
```

### Use Case 2: Filter Email Addresses
```
Pattern:      [a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}
Template:     [EMAIL]
Type:         regex
Priority:     90
Description:  Email addresses (all formats)
Case Sensitive: false
Enabled:      true
```

### Use Case 3: Filter Credit Cards
```
Pattern:      \b\d{4}[\s-]?\d{4}[\s-]?\d{4}[\s-]?\d{4}\b
Template:     [CREDIT_CARD]
Type:         regex
Priority:     100
Description:  Credit card numbers (all formats)
Case Sensitive: false
Enabled:      true
```

### Use Case 4: Filter Internal Project Names
```
Pattern:      Project Nexus
Template:     [INTERNAL_PROJECT]
Type:         phrase
Priority:     80
Description:  Internal project codename
Case Sensitive: false
Enabled:      true
```

### Use Case 5: Custom Company Secret
```
Pattern:      CompanySecret2024
Template:     Custom → [COMPANY_SECRET]
Type:         word
Priority:     70
Description:  Company confidential keyword
Case Sensitive: true
Enabled:      true
```

---

## 🧪 Testing

### Filter testen

**Bevor ein Filter produktiv geht, sollte er getestet werden:**

1. **Test-Modal öffnen**
   - In Filter-Liste: Click "Test" Button
   - Test-Modal öffnet sich

2. **Test-Text eingeben**
   ```
   Beispiel:
   "My email is john.doe@example.com and my AWS key is AKIAIOSFODNN7EXAMPLE"
   ```

3. **Test ausführen**
   - Click "Run Test"
   - Ergebnis wird angezeigt:
     ```
     Original: "My email is john.doe@example.com and my AWS key is AKIAIOSFODNN7EXAMPLE"
     Filtered: "My email is [EMAIL] and my AWS key is [***AWS_KEY***]"
     Matches: 2
     ```

4. **Ergebnis prüfen**
   - ✅ Pattern matched?
   - ✅ Replacement korrekt?
   - ✅ Keine false positives?

5. **Anpassen falls nötig**
   - Pattern verfeinern
   - Erneut testen
   - Wiederholen bis perfekt

### Bulk Test

**Mehrere Filter gleichzeitig testen:**

```bash
# Test-Script verwenden
./test-all-filters.sh

# Oder manuell via API
curl -X POST http://localhost:8080/admin/chat/test \
  -H "X-API-Key: admin_dev_key_..." \
  -d '{"messages": [{"role": "user", "content": "Test text here"}]}'
```

---

## 🐛 Troubleshooting

### Problem: Filter matched nicht

**Mögliche Ursachen:**

1. **Pattern falsch**
   - ✅ Regex Syntax prüfen
   - ✅ Online Regex Tester verwenden (regex101.com)
   - ✅ Backslashes escapen: `\\d` statt `\d`

2. **Filter Type falsch**
   - word: Matched nur ganze Wörter
   - phrase: Matched exakte Phrase
   - regex: Regex-Pattern

3. **Case Sensitivity**
   - Wenn Case Sensitive an: "Password" ≠ "password"
   - Wenn aus: beide matchen

4. **Priority zu niedrig**
   - Filter mit höherer Priority werden zuerst ausgeführt
   - Evtl. wird Text schon vorher ersetzt

5. **Filter disabled**
   - Status prüfen (grün = enabled)

**Lösung:**
- Test-Funktion verwenden
- Pattern Step-by-Step debuggen
- Dokumentation lesen

### Problem: Preview zeigt nicht das richtige Template

**Ursache:**
- Dropdown noch auf "Custom"

**Lösung:**
- Template aus Dropdown wählen
- Preview erscheint automatisch

### Problem: Bulk Import schlägt fehl

**Mögliche Ursachen:**

1. **CSV Format falsch**
   - Kommas als Separator verwenden
   - Keine Anführungszeichen um Werte
   - Kein Header in Daten (wird automatisch erkannt)

2. **Ungültige Regex**
   - Regex Syntax prüfen
   - Backslashes escapen

3. **Duplicate Patterns**
   - Pattern existiert schon
   - Entweder ändern oder alten löschen

**Lösung:**
- Fehlerhafte Zeilen korrigieren
- Erneut importieren
- Nur fehlerhafte Zeilen importieren

### Problem: Admin UI lädt nicht

**Lösung:**
```bash
# Services prüfen
docker compose -f docker-compose.dev.yml up -d

# Oder manuell
cd admin-ui && npm run dev
cd .. && ./bin/llm-proxy server
```

---

## 📚 API Reference

### Endpoints

**Basis URL**: `http://localhost:8080/admin`

#### Create Filter
```http
POST /admin/filters
Content-Type: application/json
X-API-Key: admin_dev_key_...

{
  "pattern": "test@example.com",
  "replacement": "[EMAIL]",
  "filter_type": "regex",
  "priority": 100,
  "description": "Email addresses",
  "case_sensitive": false,
  "enabled": true
}
```

#### List Filters
```http
GET /admin/filters
X-API-Key: admin_dev_key_...
```

#### Get Filter
```http
GET /admin/filters/{id}
X-API-Key: admin_dev_key_...
```

#### Update Filter
```http
PUT /admin/filters/{id}
Content-Type: application/json
X-API-Key: admin_dev_key_...

{
  "pattern": "updated@example.com",
  "replacement": "[EMAIL]",
  ...
}
```

#### Delete Filter
```http
DELETE /admin/filters/{id}
X-API-Key: admin_dev_key_...
```

#### Bulk Import
```http
POST /admin/filters/bulk-import
Content-Type: application/json
X-API-Key: admin_dev_key_...

{
  "filters": [
    {
      "pattern": "...",
      "replacement": "...",
      ...
    }
  ]
}
```

#### Test Filter
```http
POST /admin/filters/{id}/test
Content-Type: application/json
X-API-Key: admin_dev_key_...

{
  "text": "Test text with test@example.com"
}
```

#### Get Statistics
```http
GET /admin/filters/stats
X-API-Key: admin_dev_key_...
```

#### Refresh Cache
```http
POST /admin/filters/refresh
X-API-Key: admin_dev_key_...
```

**Vollständige API-Dokumentation**: `CONTENT_FILTERING.md`

---

## 🔗 Weitere Ressourcen

### Dokumentation
- `README.md` - Projekt Overview
- `CONTENT_FILTERING.md` - Vollständige API Reference
- `BULK_IMPORT_GUIDE.md` - Detaillierter Bulk Import Guide
- `QUICK_START_FILTERS.md` - Quick Start Guide
- `STARTUP_GUIDE.md` - System starten/stoppen

### Templates
- `filter-templates/enterprise-filters.csv` - 100+ Production Filters
- `filter-templates/README.md` - Template Dokumentation
- `filter-templates/CATEGORIES.md` - Kategorie-System

### Scripts
- `test-all-filters.sh` - Alle Filter testen
- `create-example-filters.sh` - Beispiel-Filter erstellen
- `start-all.sh` - Alle Services starten
- `stop-all.sh` - Alle Services stoppen

---

## ✅ Best Practices

### Filter Design

1. **Spezifisch sein**
   - Lieber mehrere spezifische Filter als ein generischer
   - Reduziert false positives

2. **Prioritäten setzen**
   - Kritische Filter (Credentials): 100
   - Wichtige Filter (PII): 90
   - Normale Filter: 80-70

3. **Testen, testen, testen**
   - Immer Test-Funktion verwenden
   - Mit realistischen Daten testen
   - Edge Cases berücksichtigen

4. **Dokumentieren**
   - Description-Feld ausfüllen
   - Zweck des Filters klar machen

5. **Templates verwenden**
   - Konsistenz
   - Keine Tippfehler
   - Einfacher zu warten

### Wartung

1. **Regelmäßig prüfen**
   - Match-Statistiken checken
   - Ungenutzte Filter deaktivieren
   - Veraltete Filter löschen

2. **Performance monitoren**
   - Cache-Age im Auge behalten
   - Bei >100 Filtern: Optimierung prüfen

3. **Backup**
   - Filter regelmäßig exportieren
   - CSV-Backups anlegen

---

## 📞 Support

Bei Fragen oder Problemen:

1. Dokumentation lesen (diese Datei)
2. `TROUBLESHOOTING.md` konsultieren
3. Logs prüfen: `logs/llm-proxy.log`
4. Issue auf GitHub erstellen

---

**Version**: 1.0  
**Status**: Production Ready  
**Last Updated**: 30. Januar 2026

**Happy Filtering! 🎉**
