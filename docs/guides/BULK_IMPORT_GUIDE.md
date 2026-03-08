# 📦 Bulk-Import Guide - Content Filter

## ✅ **Ja, du kannst Filter per Admin UI einpflegen!**

Es gibt **3 einfache Methoden**, um viele Filter auf einmal zu importieren:

---

## **Methode 1: Web Interface mit Textarea** ⭐ EMPFOHLEN

### So geht's:

```bash
# 1. Server starten
cd /home/krieger/Sites/golang-projekte/llm-proxy
./bin/llm-proxy &

# 2. Erweitertes Web Interface öffnen
firefox filter-management-advanced.html
```

### Im Browser:

1. **Tab "Textarea Import"** öffnen
2. Filter **Zeile für Zeile** eingeben (komma-separiert)
3. Auf **"Filter importieren"** klicken

### Format-Optionen:

#### **Option A: Komma-separiert** (empfohlen)
```
pattern, replacement, type, priority, description
```

**Beispiel:**
```
badword, [FILTERED], word, 100, Offensive Sprache
damn, [*], word, 100, Kraftausdrücke
confidential information, [REDACTED], phrase, 95, Vertraulich
```

#### **Option B: Tab-separiert**
```
pattern	replacement	type	priority	description
badword	[FILTERED]	word	100	Offensive Sprache
```

#### **Option C: Einfache Wortliste**
```
badword
damn
shit
```
→ Wird automatisch als Word-Filter mit `[FILTERED]` Ersetzung erstellt!

---

## **Methode 2: CSV-Datei Upload** 📄

### So geht's:

1. **CSV-Datei erstellen** (oder Beispiel verwenden)
2. **Web Interface** öffnen: `firefox filter-management-advanced.html`
3. **Tab "CSV Upload"** wählen
4. **Datei hochladen** (Drag & Drop oder Datei-Auswahl)
5. Auf **"CSV importieren"** klicken

### CSV Format:

**Header (optional):**
```csv
pattern,replacement,filter_type,priority,description,case_sensitive,enabled
```

**Beispiel:**
```csv
pattern,replacement,filter_type,priority,description
badword,[FILTERED],word,100,Offensive Sprache
damn,[*],word,100,Kraftausdrücke
confidential information,[REDACTED],phrase,95,Vertraulich
\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b,[EMAIL],regex,90,Email-Adressen
```

### Beispiel-CSV verwenden:

```bash
# Öffne die mitgelieferte Beispiel-CSV im Web Interface
# Datei: example-filters.csv (12 vordefinierte Filter)
```

---

## **Methode 3: REST API (Bulk-Import Endpoint)** 🔧

Für fortgeschrittene Nutzer oder Automation:

```bash
curl -X POST http://localhost:8080/admin/filters/bulk-import \
  -H "X-Admin-API-Key: YOUR_ADMIN_API_KEY_HERE" \
  -H "Content-Type: application/json" \
  -d '{
    "filters": [
      {
        "pattern": "badword",
        "replacement": "[FILTERED]",
        "filter_type": "word",
        "priority": 100,
        "description": "Offensive Sprache",
        "case_sensitive": false,
        "enabled": true
      },
      {
        "pattern": "confidential",
        "replacement": "[REDACTED]",
        "filter_type": "word",
        "priority": 95,
        "enabled": true
      }
    ]
  }'
```

**Response:**
```json
{
  "success": [1, 2],
  "failed": [],
  "total": 2
}
```

---

## 🚀 **Schnellstart - 3 Schritte:**

### **Schritt 1: Server starten**
```bash
cd /home/krieger/Sites/golang-projekte/llm-proxy
./bin/llm-proxy &
```

### **Schritt 2: Interface öffnen**
```bash
firefox filter-management-advanced.html
```

### **Schritt 3: Filter eingeben**

**Einfaches Beispiel (copy & paste in Textarea):**
```
badword, [FILTERED], word, 100, Offensive Sprache
damn, [*], word, 100, Kraftausdrücke
confidential information, [REDACTED], phrase, 95, Vertraulich
Project Phoenix, [INTERNAL], phrase, 95, Interner Projektname
```

→ Klicke auf "Filter importieren" ✅

---

## 📋 **Format-Referenz**

### Pflichtfelder:
- `pattern` - Was gefiltert werden soll
- `replacement` - Womit ersetzt werden soll

### Optionale Felder:
- `filter_type` - `word` (Standard), `phrase`, oder `regex`
- `priority` - 0-999 (Standard: 100, höher = zuerst)
- `description` - Beschreibung
- `case_sensitive` - `true`/`false` (Standard: false)
- `enabled` - `true`/`false` (Standard: true)

---

## 💡 **Beispiele**

### **Schimpfwörter-Liste (einfach)**
```
badword
damn
shit
hell
crap
```
→ Alle werden zu Word-Filtern mit `[FILTERED]` Ersetzung

### **Vertrauliche Begriffe (mit Beschreibung)**
```
confidential information, [REDACTED], phrase, 95, Geheime Infos
Project Phoenix, [INTERNAL_PROJECT], phrase, 95, Internes Projekt
top secret, [CLASSIFIED], phrase, 95, Geheim
```

### **PII-Filter (Regex)**
```
\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b, [EMAIL], regex, 90, Email-Adressen
\b0[0-9]{2,4}[-\s]?[0-9]{3,8}\b, [PHONE], regex, 90, Telefonnummern
\b[0-9]{4}[\s-]?[0-9]{4}[\s-]?[0-9]{4}[\s-]?[0-9]{4}\b, [CREDIT_CARD], regex, 95, Kreditkarten
```

---

## ✅ **Vorteile des Bulk-Imports**

| Feature | Einzeln | Bulk |
|---------|---------|------|
| **Geschwindigkeit** | ⏱️ Langsam | ⚡ Schnell |
| **Fehleranfällig** | ❌ Tippfehler | ✅ Copy & Paste |
| **Übersicht** | 🤔 Schwer | ✅ Excel/CSV |
| **Wiederverwendbar** | ❌ Nein | ✅ CSV speichern |
| **Versionierung** | ❌ Nein | ✅ Git-freundlich |

---

## 📂 **Dateien**

| Datei | Zweck |
|-------|-------|
| `filter-management-advanced.html` | Web Interface mit Bulk-Import |
| `example-filters.csv` | Beispiel CSV-Datei (12 Filter) |
| `create-example-filters.sh` | Script für Beispiel-Filter |
| `BULK_IMPORT_GUIDE.md` | Diese Anleitung |

---

## 🔍 **Vergleich: Einzeln vs. Bulk**

### Einzeln erstellen (bisherig):
```bash
# 10 Filter = 10 API Calls
curl -X POST .../admin/filters -d '{...}'  # Filter 1
curl -X POST .../admin/filters -d '{...}'  # Filter 2
curl -X POST .../admin/filters -d '{...}'  # Filter 3
...
```

### Bulk-Import (neu):
```bash
# 10 Filter = 1 API Call + Web Interface
1. Excel öffnen
2. Filter-Liste erstellen
3. Als CSV speichern
4. In Web Interface hochladen
5. Fertig! ✅
```

---

## 🛠️ **Workflow für große Filter-Listen**

### Szenario: 100+ Filter verwalten

1. **Excel/Google Sheets öffnen**
2. **Spalten erstellen:**
   - Spalte A: Pattern
   - Spalte B: Replacement
   - Spalte C: Type
   - Spalte D: Priority
   - Spalte E: Description

3. **Filter eintragen** (mit Excel-Features wie Auto-Fill)

4. **Als CSV exportieren**

5. **In Web Interface hochladen**

6. **Versionierung** (optional):
   ```bash
   git add filters.csv
   git commit -m "Added 100 new content filters"
   ```

---

## 🎯 **Use Cases**

### **Use Case 1: Compliance-Vorgaben**
```csv
pattern,replacement,filter_type,priority,description
GDPR violations,***REDACTED***,phrase,100,Compliance
personal data,***REDACTED***,phrase,100,DSGVO
sensitive information,***CLASSIFIED***,phrase,100,Datenschutz
```

### **Use Case 2: Unternehmens-Glossar**
```csv
pattern,replacement,filter_type,priority,description
CompanyName,[COMPANY],word,90,Firmenname schützen
Project Alpha,[PROJECT],phrase,90,Internes Projekt
CEO Name,[CEO],word,95,Geschäftsführer
```

### **Use Case 3: Sprach-Säuberung**
```csv
pattern,replacement,filter_type,priority,description
badword1,[*],word,100,Offensive language
badword2,[*],word,100,Offensive language
badword3,[*],word,100,Offensive language
...
```

---

## 🚨 **Troubleshooting**

### Problem: Import schlägt fehl
**Lösung:**
1. Prüfe CSV-Format (komma-separiert)
2. Prüfe Pflichtfelder (pattern, replacement)
3. Validiere Regex-Patterns mit `/admin/filters/test`

### Problem: Filter werden nicht angewendet
**Lösung:**
```bash
# Cache aktualisieren
curl -X POST http://localhost:8080/admin/filters/refresh \
  -H "X-Admin-API-Key: YOUR_ADMIN_API_KEY_HERE"
```

### Problem: Zu viele Fehler beim Import
**Lösung:**
- Importiere Filter in kleineren Batches (z.B. 20 Filter pro Import)
- Prüfe die Error-Messages im Results-Bereich
- Teste einzelne problematische Filter mit `/admin/filters/test`

---

## 📊 **Import-Statistiken**

Das Web Interface zeigt nach dem Import:

✅ **Erfolg:** Anzahl erfolgreich erstellter Filter  
❌ **Fehler:** Liste fehlgeschlagener Filter mit Grund  
📊 **Gesamt:** Anzahl verarbeiteter Zeilen

**Beispiel:**
```
Import abgeschlossen! 8 erfolgreich, 2 fehlgeschlagen

Gesamt: 10 | Erfolg: 8 | Fehler: 2

✅ Erfolgreich:
✓ Filter ID 1 erstellt
✓ Filter ID 2 erstellt
...

❌ Fehlgeschlagen:
✗ Filter 9 (test***): Invalid regex pattern
✗ Filter 10 (): pattern is required
```

---

## 📚 **Weitere Ressourcen**

- **Vollständige API Dokumentation:** `CONTENT_FILTERING.md`
- **Einzelfilter-Erstellung:** `filter-management.html`
- **Schnellstart:** `QUICK_START_FILTERS.md`
- **Test-Script:** `# Use Admin UI at http://localhost:3005`

---

## 🎉 **Zusammenfassung**

**JA, du kannst Filter per Admin UI einpflegen!**

✅ **Textarea:** Einfach Copy & Paste von Listen  
✅ **CSV-Upload:** Excel → Export → Upload  
✅ **API:** Für Automation und Scripts  

**Empfehlung:** Nutze das Web Interface für bis zu 50 Filter, CSV für größere Listen!

---

**Happy Bulk-Importing! 🚀**
