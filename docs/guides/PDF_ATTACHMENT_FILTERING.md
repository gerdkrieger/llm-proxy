# PDF & Attachment Filtering Guide

## 📄 Übersicht

LLM-Proxy kann **PDFs und Bild-Attachments automatisch filtern** und vertrauliche Informationen schwärzen (redaction), bevor sie an den LLM-Provider gesendet werden.

### Unterstützte Dateitypen
- ✅ PDF-Dokumente (.pdf)
- ✅ Bilder (JPEG, PNG, GIF, WebP)
- ✅ Base64-kodierte Attachments in Chat-Messages

### Wie es funktioniert

```
1. PDF/Bild → 2. OCR (Texterkennung) → 3. Filter anwenden → 4. Visuell schwärzen → 5. Gefiltertes Dokument
```

---

## 🔧 Technische Architektur

### Komponenten

| Komponente | Zweck | Technologie |
|------------|-------|-------------|
| **RedactionService** | Kern-Service für Dokument-Schwärzung | Go |
| **AttachmentService** | Attachment-Analyse im Chat | Go |
| **FilterService** | Content Filter (bestehend) | Go |
| **OCR Engine** | Texterkennung in PDFs/Bildern | Tesseract OCR |
| **PDF Processor** | PDF-zu-Bild Konvertierung | pdftoppm |
| **Visual Redaction** | Schwarze Boxen über Text | Ghostscript + ImageMagick |

### Workflow

```go
// 1. Chat-Message kommt mit PDF Attachment rein
POST /v1/chat/completions
{
  "messages": [{
    "role": "user",
    "content": [
      { "type": "text", "text": "Analysiere dieses Dokument" },
      { "type": "image_url", "image_url": { "url": "data:application/pdf;base64,..." } }
    ]
  }]
}

// 2. AttachmentService analysiert das Attachment
service.AnalyzeAttachments(ctx, messages)

// 3. RedactionService führt OCR durch
ocrWords := performOCR(pdfFile)

// 4. Content Filter erkennen PII
piiMatches := detectPIIInOCR(ctx, ocrWords)
// Beispiel: "john@example.com" → EMAIL
// Beispiel: "4532 1234 5678 9012" → CREDIT_CARD

// 5. Visuell schwärzen
redactedPDF := visuallyRedact(inputFile, piiMatches)

// 6. Gefiltertes PDF zurück in Message
message.content[1].image_url.url = "data:application/pdf;base64,[REDACTED_PDF]"
```

---

## 🚀 Verwendung

### Im Chat (OpenAI-kompatible API)

```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer your_client_key" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-3-sonnet",
    "messages": [
      {
        "role": "user",
        "content": [
          {
            "type": "text",
            "text": "Bitte analysiere dieses Dokument"
          },
          {
            "type": "image_url",
            "image_url": {
              "url": "data:application/pdf;base64,JVBERi0xLjQK..."
            }
          }
        ]
      }
    ]
  }'
```

**Was passiert:**
1. PDF wird erkannt (Base64 + MIME-Type)
2. OCR extrahiert Text mit Koordinaten
3. Content Filter finden PII (Email, Telefon, etc.)
4. PDF wird geschwärzt (schwarze Boxen)
5. Geschwärztes PDF geht an Claude/OpenAI
6. Filter Matches werden in DB geloggt

---

## 🎯 Welche Inhalte werden gefiltert?

### Automatisch erkannte PII-Typen

Alle konfigurierten Content Filter werden angewendet:

| Filter-Typ | Beispiel | Ersetzung |
|------------|----------|-----------|
| **Email** | `john@example.com` | `[EMAIL_ENTFERNT]` |
| **Telefon** | `0123-456789` | `[TELEFON_ENTFERNT]` |
| **Kreditkarte** | `4532 1234 5678 9012` | `[KREDITKARTE_ENTFERNT]` |
| **Vertraulich** | `confidential information` | `[VERTRAULICH_ENTFERNT]` |
| **Projekt-Namen** | `Project Phoenix` | `[INTERNES_PROJEKT]` |
| **Custom Patterns** | Eigene Regex/Word-Filter | Eigene Replacements |

Alle Filter aus der Admin UI (`/admin/filters`) werden verwendet!

---

## 📋 Code-Beispiele

### 1. Python-Client mit PDF

```python
import requests
import base64

# PDF laden und Base64 kodieren
with open("document.pdf", "rb") as f:
    pdf_base64 = base64.b64encode(f.read()).decode("utf-8")

response = requests.post(
    "http://localhost:8080/v1/chat/completions",
    headers={
        "Authorization": "Bearer your_client_key",
        "Content-Type": "application/json"
    },
    json={
        "model": "claude-3-sonnet",
        "messages": [{
            "role": "user",
            "content": [
                {
                    "type": "text",
                    "text": "Fasse dieses Dokument zusammen"
                },
                {
                    "type": "image_url",
                    "image_url": {
                        "url": f"data:application/pdf;base64,{pdf_base64}"
                    }
                }
            ]
        }]
    }
)

print(response.json())
```

### 2. JavaScript/Node.js mit PDF

```javascript
const fs = require('fs');
const axios = require('axios');

// PDF laden
const pdfBuffer = fs.readFileSync('document.pdf');
const pdfBase64 = pdfBuffer.toString('base64');

axios.post('http://localhost:8080/v1/chat/completions', {
  model: 'claude-3-sonnet',
  messages: [{
    role: 'user',
    content: [
      { type: 'text', text: 'Analysiere dieses PDF' },
      {
        type: 'image_url',
        image_url: {
          url: `data:application/pdf;base64,${pdfBase64}`
        }
      }
    ]
  }]
}, {
  headers: {
    'Authorization': 'Bearer your_client_key',
    'Content-Type': 'application/json'
  }
}).then(response => {
  console.log(response.data);
});
```

### 3. cURL mit Bild-Attachment

```bash
# Bild Base64 kodieren
IMAGE_B64=$(base64 -w 0 image.png)

curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer your_key" \
  -H "Content-Type: application/json" \
  -d "{
    \"model\": \"gpt-4-vision\",
    \"messages\": [{
      \"role\": \"user\",
      \"content\": [
        {\"type\": \"text\", \"text\": \"Was ist auf diesem Bild?\"},
        {
          \"type\": \"image_url\",
          \"image_url\": {\"url\": \"data:image/png;base64,$IMAGE_B64\"}
        }
      ]
    }]
  }"
```

---

## 🔍 Wie funktioniert OCR-basierte Schwärzung?

### Schritt 1: PDF zu Bild konvertieren

```bash
# pdftoppm konvertiert PDF Seite 1 zu PNG (300 DPI)
pdftoppm -png -f 1 -l 1 -singlefile -r 300 input.pdf output
```

**Resultat:** `output.png` mit hoher Auflösung für OCR

### Schritt 2: Tesseract OCR

```bash
# Tesseract extrahiert Text mit Koordinaten (hOCR Format)
tesseract input.png output -l eng+deu hocr
```

**Resultat:** `output.hocr` XML mit:
```xml
<span class='ocrx_word' title='bbox 120 450 280 480'>john@example.com</span>
<span class='ocrx_word' title='bbox 120 500 350 530'>confidential</span>
```

**Koordinaten:**
- `bbox 120 450 280 480` = X:120, Y:450, Width:160, Height:30

### Schritt 3: PII Erkennung

```go
// Content Filter werden auf OCR-Text angewandt
text := "john@example.com confidential information"

filters := filterService.GetEnabledFilters(ctx)
// Filter 1: Email Regex → Matches "john@example.com"
// Filter 2: Phrase "confidential information" → Matches

// OCR-Koordinaten werden den Matches zugeordnet
match1 := PIIMatch{
    Type: "EMAIL",
    Text: "john@example.com",
    X: 120, Y: 450, Width: 160, Height: 30
}
```

### Schritt 4: Visuelles Schwärzen

#### Für PDFs (Ghostscript):

```bash
# PostScript Code generieren für schwarze Boxen
cat > redaction.ps << EOF
%!PS-Adobe-3.0
0 0 0 setrgbcolor
newpath
120 450 moveto
280 450 lineto
280 480 lineto
120 480 lineto
closepath
fill
EOF

# Schwärzung auf PDF anwenden
gs -dBATCH -dNOPAUSE -sDEVICE=pdfwrite \
   -sOutputFile=redacted.pdf \
   input.pdf redaction.ps
```

#### Für Bilder (ImageMagick):

```bash
# Schwarze Box zeichnen
convert input.png \
  -fill black \
  -draw "rectangle 120,450 280,480" \
  redacted.png
```

**Resultat:** Geschwärztes Dokument mit schwarzen Boxen über PII

---

## 🛠️ Dependencies

### Erforderliche System-Tools

| Tool | Zweck | Installation |
|------|-------|--------------|
| **tesseract** | OCR (Texterkennung) | `apt install tesseract-ocr` |
| **tesseract-data** | Sprachen (DE+EN) | `apt install tesseract-ocr-deu tesseract-ocr-eng` |
| **pdftoppm** | PDF→Bild | `apt install poppler-utils` |
| **ghostscript** | PDF Schwärzung | `apt install ghostscript` |
| **convert** | Bild Schwärzung | `apt install imagemagick` |

### Prüfen ob installiert:

```bash
which tesseract    # /usr/bin/tesseract
which pdftoppm     # /usr/bin/pdftoppm
which gs           # /usr/bin/gs
which convert      # /usr/bin/convert

# Sprachen prüfen
tesseract --list-langs
# deu (Deutsch)
# eng (Englisch)
```

### Installation (Ubuntu/Debian):

```bash
sudo apt update
sudo apt install -y \
  tesseract-ocr \
  tesseract-ocr-deu \
  tesseract-ocr-eng \
  poppler-utils \
  ghostscript \
  imagemagick
```

---

## 📊 Monitoring & Logging

### Filter Matches werden geloggt

```sql
SELECT * FROM filter_matches
WHERE filter_id = 0  -- 0 = Attachment Redaction
ORDER BY created_at DESC;
```

**Beispiel-Eintrag:**
```
id: 123
filter_id: 0  (NULL für Attachment-Redactions)
original_text: "john@example.com"
replacement: "[EMAIL_ENTFERNT]"
filter_type: "regex"
match_count: 1
created_at: 2026-02-06 12:34:56
```

### Logs prüfen:

```bash
# Backend Logs
docker logs llm-proxy-backend | grep -i "redact\|ocr\|attachment"

# Beispiel Output:
# INFO Starting redaction for file: document.pdf (type: pdf)
# INFO OCR extracted 245 words from document
# INFO Found 3 PII matches, proceeding with redaction
# INFO Successfully redacted 3 locations in document
```

### Admin UI - Filter Statistics

```bash
curl -H "X-Admin-API-Key: your_key" \
  http://localhost:8080/admin/filters/stats

{
  "total_filters": 6,
  "enabled_filters": 6,
  "total_matches": 42,
  "by_type": {
    "word": 12,
    "phrase": 8,
    "regex": 22
  }
}
```

---

## 🧪 Testing

### 1. Test mit Beispiel-PDF

```bash
# Erstelle Test-PDF mit PII
cat > test.html << 'EOF'
<html>
<body>
  <h1>Vertrauliches Dokument</h1>
  <p>Email: john.doe@example.com</p>
  <p>Telefon: 0123-456789</p>
  <p>Kreditkarte: 4532 1234 5678 9012</p>
  <p>Project Phoenix - confidential information</p>
</body>
</html>
EOF

# HTML zu PDF konvertieren
wkhtmltopdf test.html test.pdf

# PDF Base64 kodieren
PDF_B64=$(base64 -w 0 test.pdf)

# An LLM-Proxy senden
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer your_key" \
  -H "Content-Type: application/json" \
  -d "{
    \"model\": \"claude-3-sonnet\",
    \"messages\": [{
      \"role\": \"user\",
      \"content\": [
        {\"type\": \"text\", \"text\": \"Extrahiere alle Informationen\"},
        {
          \"type\": \"image_url\",
          \"image_url\": {\"url\": \"data:application/pdf;base64,$PDF_B64\"}
        }
      ]
    }]
  }"
```

**Erwartetes Ergebnis:**
- Email wird geschwärzt
- Telefon wird geschwärzt
- Kreditkarte wird geschwärzt
- "Project Phoenix" wird geschwärzt
- "confidential information" wird geschwärzt

### 2. Filter Matches prüfen

```bash
# Prüfe DB
docker exec llm-proxy-postgres psql -U proxy_user -d llm_proxy \
  -c "SELECT original_text, replacement FROM filter_matches WHERE filter_id = 0 ORDER BY created_at DESC LIMIT 10;"
```

### 3. Geschwärztes PDF herunterladen

Das geschwärzte PDF ist in der Response als Base64:

```javascript
// Response
{
  "choices": [{
    "message": {
      "content": [...],
      // Geschwärztes PDF ist hier (wenn zurückgegeben)
    }
  }]
}
```

---

## 🔒 Sicherheitsaspekte

### Was wird NICHT an LLM-Provider gesendet?

✅ **Vertrauliche Informationen werden geschwärzt:**
- Emails
- Telefonnummern
- Kreditkarten
- Interne Projektnamen
- Custom-definierte sensible Daten

### Temporäre Dateien

- Temporäre Dateien werden in `/tmp/llm-proxy-redaction/` erstellt
- Werden nach Verarbeitung automatisch gelöscht
- Original-Dokument wird NICHT persistent gespeichert

```go
defer os.Remove(tempFile)        // Original
defer os.Remove(redactedFile)    // Geschwärzte Version
```

### Filter-Bypass-Verhinderung

- OCR erkennt Text auch in Bildern
- Visuelles Schwärzen (nicht nur Text-Ersetzung)
- Screenshots von PDFs werden ebenfalls gefiltert

---

## ⚙️ Konfiguration

### Content Filter verwalten

Alle Filter werden über die Admin UI verwaltet:

```bash
# Admin UI öffnen
open http://localhost:3005

# Oder via API
curl -H "X-Admin-API-Key: your_key" \
  http://localhost:8080/admin/filters
```

**Filter erstellen:**

```bash
curl -X POST http://localhost:8080/admin/filters \
  -H "X-Admin-API-Key: your_key" \
  -H "Content-Type: application/json" \
  -d '{
    "pattern": "\\b[0-9]{4}[\\s-]?[0-9]{4}[\\s-]?[0-9]{4}[\\s-]?[0-9]{4}\\b",
    "replacement": "[KREDITKARTE_ENTFERNT]",
    "description": "Kreditkartennummern",
    "filter_type": "regex",
    "case_sensitive": false,
    "enabled": true,
    "priority": 95
  }'
```

### OCR-Sprachen anpassen

```go
// In redaction_service.go, Zeile ~160
cmd := exec.Command("tesseract", imageFile, outputBase, 
    "-l", "eng+deu+fra",  // Englisch + Deutsch + Französisch
    "hocr")
```

---

## 🐛 Troubleshooting

### Problem: OCR funktioniert nicht

```bash
# Prüfe Tesseract
tesseract --version
tesseract --list-langs

# Falls fehlt:
sudo apt install tesseract-ocr tesseract-ocr-deu tesseract-ocr-eng
```

### Problem: PDF-Konvertierung schlägt fehl

```bash
# Prüfe pdftoppm
which pdftoppm

# Falls fehlt:
sudo apt install poppler-utils
```

### Problem: Keine Schwärzung sichtbar

**Mögliche Ursachen:**
1. Keine Filter aktiv → Prüfe `/admin/filters`
2. OCR-Qualität schlecht → Erhöhe PDF-Auflösung (300 DPI)
3. Text ist in Bild eingebettet → OCR sollte es erkennen

```bash
# Logs prüfen
docker logs llm-proxy-backend | grep "redaction\|OCR"
```

### Problem: Performance langsam

**OCR ist rechenintensiv:**
- PDF mit 10 Seiten: ~5-10 Sekunden
- Großes Bild: ~2-3 Sekunden

**Optimierungen:**
- Nur erste Seite von PDFs verarbeiten (aktuell implementiert)
- Cache für bereits verarbeitete Dokumente (TODO)
- Asynchrone Verarbeitung (TODO)

---

## ⚠️ Bekannte Limitierungen

### PDF-Unterstützung mit Claude API

**Problem:** Claude's API akzeptiert nur Bild-MIME-Types (`image/png`, `image/jpeg`, `image/gif`, `image/webp`), aber keine PDFs (`application/pdf`).

**Technische Details:**
- PDFs die als `data:application/pdf;base64,...` gesendet werden, werden von Claude abgelehnt
- Fehlermeldung: `"Input should be 'image/jpeg', 'image/png', 'image/gif' or 'image/webp'"`
- Dies ist eine Einschränkung der Claude API, nicht des LLM-Proxy

**Aktuelles Verhalten:**
```json
// ❌ Funktioniert NICHT mit Claude API
{
  "type": "image_url",
  "image_url": {
    "url": "data:application/pdf;base64,JVBERi0xLjQK..."
  }
}

// ✅ Funktioniert mit Claude API  
{
  "type": "image_url",
  "image_url": {
    "url": "data:image/png;base64,iVBORw0KGgoA..."
  }
}
```

**Workaround (aktuell):**

Benutzer müssen PDFs als Bilder senden:

```bash
# PDF zu PNG konvertieren
pdftoppm -png -f 1 -l 1 -singlefile -r 300 input.pdf output

# Als Base64 kodieren
base64 output.png > output.base64

# Als image/png senden (nicht application/pdf)
curl -X POST https://llmproxy.aitrail.ch/v1/chat/completions \
  -H "Authorization: Bearer YOUR_KEY" \
  -d '{
    "model": "claude-3-haiku-20240307",
    "messages": [{
      "role": "user",
      "content": [
        {"type": "text", "text": "Analysiere dieses Dokument"},
        {"type": "image_url", "image_url": {"url": "data:image/png;base64,..."}}
      ]
    }]
  }'
```

**Zukünftige Lösung (Enhancement):**

Der LLM-Proxy könnte automatisch PDFs zu Bildern konvertieren:

```go
// Pseudo-Code für zukünftige Implementierung
func (s *Service) analyzeImageData(ctx context.Context, dataURL string) {
    // 1. Erkenne application/pdf MIME-Type
    if strings.Contains(dataURL, "application/pdf") {
        // 2. Extrahiere PDF-Daten
        pdfData := extractBase64Data(dataURL)
        
        // 3. Konvertiere zu PNG
        pngData := s.redactionService.convertPDFToImage(pdfData)
        
        // 4. Update MIME-Type
        dataURL = "data:image/png;base64," + base64.Encode(pngData)
    }
    
    // 5. OCR und Redaktion wie gewohnt
    result := s.redactionService.RedactDocument(ctx, imageData, filename)
}
```

**Code-Änderungen erforderlich:**
- Datei: `internal/application/attachment/service.go`
- Funktion: `analyzeImageData()` (Zeilen 141-219)
- Aufwand: ca. 2-4 Stunden

**Status:** Enhancement geplant, nicht kritisch (Workaround verfügbar)

**Betroffene Anbieter:**
- ✅ **Claude (Anthropic):** Nur Bilder, keine PDFs
- ⚠️  **OpenAI (GPT-4 Vision):** Unterstützt ebenfalls nur Bilder
- ℹ️  **Andere Anbieter:** Müssen individuell geprüft werden

### Alternative: PDFs als Text-Extraktion

Für reine Text-PDFs ohne visuelle Redaktion:

```bash
# PDF-Text extrahieren (ohne OCR)
pdftotext input.pdf - | jq -Rs '{
  "model": "claude-3-haiku-20240307",
  "messages": [{
    "role": "user", 
    "content": "Analysiere diesen Text:\n\n" + .
  }]
}'
```

**Vorteile:**
- ✅ Funktioniert mit allen LLM-Anbietern
- ✅ Keine Bildkonvertierung notwendig
- ✅ Content Filter funktionieren auf reinem Text

**Nachteile:**
- ❌ Keine visuelle Schwärzung (nur Text-Replacement)
- ❌ Formatierung geht verloren
- ❌ Keine OCR bei gescannten PDFs

---

## 📚 Weiterführende Dokumentation

- **Content Filtering:** [CONTENT_FILTERING.md](CONTENT_FILTERING.md)
- **Filter Management:** [FILTER_MANAGEMENT_GUIDE.md](FILTER_MANAGEMENT_GUIDE.md)
- **Admin API:** [ADMIN_API.md](ADMIN_API.md)

---

## 🎯 Zusammenfassung

**LLM-Proxy bietet vollautomatische PDF & Attachment-Filterung:**

1. ✅ **OCR-basierte Texterkennung** (Tesseract)
2. ✅ **Content Filter Anwendung** (bestehende Filter)
3. ✅ **Visuelle Schwärzung** (Ghostscript/ImageMagick)
4. ✅ **Logging & Monitoring** (Filter Matches DB)
5. ✅ **Produktionsfertig** (Dependencies installiert)

**Unterstützt:**
- PDF-Dokumente
- Bild-Attachments (JPEG, PNG, etc.)
- Base64-kodierte Dateien
- Multimodale Chat-Messages

**Nächste Schritte:**
1. Content Filter in Admin UI konfigurieren
2. PDF-Attachment im Chat senden
3. Geschwärzte Version wird an LLM gesendet
4. Filter Matches in DB prüfen

---

**Last Updated:** 2026-02-06  
**Status:** ✅ Produktionsfertig  
**Dependencies:** ✅ Alle installiert
