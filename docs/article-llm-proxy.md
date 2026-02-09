# LLM-Proxy: Warum jedes Unternehmen einen intelligenten Gateway zwischen sich und den KI-Anbietern braucht

## Das Problem: KI ohne Kontrolle

Die meisten Unternehmen nutzen heute KI-Modelle von OpenAI, Anthropic oder anderen Anbietern. Die Integration ist simpel: API-Key rein, fertig. Doch genau diese Einfachheit wird zum Risiko.

- **Wer hat wann welches Modell genutzt — und was wurde gesendet?**
- **Was passiert, wenn ein Mitarbeiter vertrauliche Kundendaten in den Prompt schreibt?**
- **Wie kontrolliert man Kosten, wenn 15 Teams gleichzeitig GPT-5 nutzen?**
- **Was wenn der API-Key leakt — wer hat ihn, wo steckt er überall?**

Die ehrliche Antwort in den meisten Betrieben: Niemand weiss es.

---

## Die Lösung: Ein zentraler LLM-Gateway

**LLM-Proxy** ist ein selbst gehosteter Gateway-Server, der sich zwischen Ihre internen Tools und die KI-Anbieter schaltet. Jede Anfrage läuft durch diesen Proxy — mit voller Kontrolle, Transparenz und Sicherheit.

Die Software ist in Go geschrieben, läuft als Docker-Container und bietet eine vollständige Admin-Oberfläche zur Verwaltung.

---

## Was LLM-Proxy konkret leistet

### 1. Zentrales API-Key-Management mit Verschlüsselung

Anstatt API-Keys in Konfigurationsdateien, Umgebungsvariablen oder schlimmstenfalls im Quellcode zu verteilen, werden alle Provider-Keys zentral in der Datenbank gespeichert — verschlüsselt mit **AES-256-GCM**.

- Keys werden über die Admin-Oberfläche hinzugefügt und verwaltet
- Nach dem Speichern ist der Klartext nie wieder sichtbar — nur ein Hint (z.B. `...r5PZ`)
- Keys können einzeln aktiviert, deaktiviert oder gelöscht werden
- Gewichtung und Rate-Limits pro Key konfigurierbar
- Fallback auf Konfigurations-Keys, wenn keine DB-Keys vorhanden sind

**Vorteil für den Betrieb:** Kein Mitarbeiter muss jemals einen Provider-API-Key sehen oder besitzen. Die Keys liegen an genau einer Stelle, verschlüsselt, mit Audit-Trail.

### 2. Multi-Provider mit einem einzigen Endpunkt

LLM-Proxy spricht die **OpenAI-kompatible API** — der De-facto-Standard, den praktisch jedes Tool versteht. Intern routet der Proxy die Anfragen an den richtigen Provider:

- **Anthropic Claude** (Opus, Sonnet, Haiku — alle Generationen)
- **OpenAI** (GPT-5, GPT-4.1, o-Serie Reasoning-Modelle, und weitere)

Ihre internen Tools — ob OpenWebUI, Cursor, eigene Applikationen oder Skripte — verbinden sich mit einem einzigen Endpunkt. Den Provider-Wechsel erledigt der Proxy.

**Vorteil für den Betrieb:** Sie sind nicht an einen Anbieter gebunden. Modellwechsel, Preisvergleiche oder Failover passieren zentral, ohne dass ein einziges internes Tool angepasst werden muss.

### 3. Client-Authentifizierung und Zugriffskontrolle

Jedes Tool, jeder Nutzer, jede Abteilung bekommt einen eigenen **OAuth-Client** mit dediziertem API-Key:

- Clients werden über die Admin-UI erstellt und verwaltet
- Jeder Client hat eigene Scopes (Lese- und Schreibrechte)
- Clients können einzeln gesperrt werden — sofort, ohne Seiteneffekte
- Dreistufige Authentifizierung: Statische API-Keys, DB-basierte Client-Secrets (bcrypt-gehasht), OAuth-Tokens

**Vorteil für den Betrieb:** Wenn ein Mitarbeiter das Unternehmen verlässt oder ein Tool ausgemustert wird, deaktivieren Sie den Client mit einem Klick. Kein Key-Rotation-Drama über 20 Systeme hinweg.

### 4. Content-Filtering und Datenschutz

Das integrierte **Content-Filter-System** prüft jede Anfrage, bevor sie den Provider erreicht:

- **Wort-, Phrase- und Regex-Filter** mit konfigurierbaren Ersetzungen
- **Kategorien und Prioritäten** für strukturierte Filter-Verwaltung
- **Bulk-Import** für unternehmensweite Filterregeln
- **Live-Test** von Filtern direkt in der Admin-UI
- Blockierte Inhalte werden protokolliert — mit vollständigem Match-Log

**Vorteil für den Betrieb:** Vertrauliche Firmennamen, Kundendaten, interne Projektnamen oder Compliance-relevante Begriffe werden abgefangen, bevor sie jemals einen externen Server erreichen. Das ist nicht optional — für regulierte Branchen ist es Pflicht.

### 5. Live-Monitoring und Request-Logging

Jede einzelne API-Anfrage wird in Echtzeit protokolliert und ist über die Admin-UI einsehbar:

- **Live Monitor** mit Echtzeit-Aktualisierung
- Suchfunktion über alle Felder (Modell, Client, Status, Pfad)
- Quick-Filter: Nur API-Requests, nur Fehler, Health-Checks ausblenden
- Detail-Ansicht pro Request: Headers, Body, Response, Timing
- Client-Identifikation: Welcher Client hat welche Anfrage gesendet

**Vorteil für den Betrieb:** Vollständige Transparenz. Sie sehen in Echtzeit, wer was an welchen Provider sendet, wie lange es dauert, und ob Fehler auftreten. Unverzichtbar für Debugging, Kostenanalyse und Compliance-Audits.

### 6. Model-Management

Nicht jedes Team braucht Zugriff auf jedes Modell. Über die Admin-UI steuern Sie pro Provider:

- Welche Modelle aktiv und sichtbar sind
- Modelle einzeln ein-/ausschalten
- Vollständige Modell-Kataloge für Claude und OpenAI (inkl. neueste Generationen)

**Vorteil für den Betrieb:** Das Marketing-Team braucht kein o3-pro Reasoning-Modell für 100x den Preis. Sie aktivieren pro Kontext genau die Modelle, die sinnvoll und budgetkonform sind.

### 7. Caching und Kostenoptimierung

Identische Anfragen werden zwischengespeichert (Redis):

- Konfigurierbarer TTL und Cache-Grösse
- Cache-Statistiken in der Admin-UI
- Gezieltes Invalidieren nach Modell
- Komplett-Clear bei Bedarf

**Vorteil für den Betrieb:** Wiederkehrende Anfragen (z.B. System-Prompts, Template-Generierungen) kosten nur einmal. Je nach Nutzungsprofil spart das 20-40% der API-Kosten.

### 8. Metriken und Observability

Eingebauter **Prometheus-Endpunkt** für nahtlose Integration in bestehende Monitoring-Stacks:

- Request-Counts, Latenzen, Fehlerraten
- Provider-Health-Status
- Cache-Hit-Raten
- Integration mit Grafana, Datadog oder jedem Prometheus-kompatiblen Tool

---

## Die Admin-Oberfläche

LLM-Proxy wird mit einer vollständigen **Web-Admin-UI** ausgeliefert:

- **Dashboard** — Systemstatus, Provider-Health, aktive Modelle auf einen Blick
- **Providers** — Provider-Konfiguration, API-Key-Management, Model-Auswahl, Connection-Tests
- **Live Monitor** — Echtzeit-Request-Log mit Filtern und Detail-Ansicht
- **Clients** — OAuth-Client-Verwaltung mit Secret-Reset
- **Content Filters** — Filter-CRUD, Bulk-Import, Live-Testing
- **Settings** — Systemweite Konfiguration
- **Nutzungsstatistiken** — Auswertungen nach Client, Modell und Zeitraum

Die UI läuft als separater Container (nginx) und kommuniziert über die Admin-API mit dem Backend. Kein zusätzlicher Server, kein Framework-Lock-in.

---

## Technische Architektur

| Komponente | Technologie |
|---|---|
| Backend | Go (kompiliert, single binary) |
| Datenbank | PostgreSQL |
| Cache | Redis |
| Admin UI | Svelte + Tailwind CSS |
| Deployment | Docker (Multi-Container) |
| Verschlüsselung | AES-256-GCM |
| Auth | OAuth 2.0 + bcrypt + API-Keys |
| Metriken | Prometheus |
| Reverse Proxy | nginx (Admin UI) |

Die gesamte Infrastruktur läuft auf einem einzigen Server — ein DigitalOcean Droplet reicht. Kein Kubernetes, kein Cloud-Lock-in, keine versteckten Kosten.

---

## Für wen ist LLM-Proxy?

- **KMU und Mittelstand**, die KI-Tools intern einsetzen und Kontrolle über Datenflüsse brauchen
- **Agenturen und IT-Dienstleister**, die mehrere KI-Provider für verschiedene Kunden managen
- **Regulierte Branchen** (Finanzen, Gesundheit, Recht), die nachweisen müssen, welche Daten an externe APIs gehen
- **Entwicklungsteams**, die einen einheitlichen API-Endpunkt für alle LLM-Integrationen wollen
- **Jedes Unternehmen**, das mehr als einen API-Key im Umlauf hat und nachts ruhig schlafen möchte

---

## Fazit

KI-APIs direkt einzubinden ist wie ein Firmennetzwerk ohne Firewall zu betreiben — es funktioniert, bis es das nicht mehr tut.

**LLM-Proxy** gibt Ihnen die Kontrolle zurück: Wer nutzt was, was wird gesendet, was darf raus, was kostet es. Alles an einer Stelle, selbst gehostet, vollständig transparent.

Die Software ist Open Source und kann sofort deployed werden. Ein Server, ein Docker-Compose, fertig.

---

*LLM-Proxy wird entwickelt von Krieger Engineering. Bei Fragen zur Integration oder zum Deployment kontaktieren Sie uns direkt.*
