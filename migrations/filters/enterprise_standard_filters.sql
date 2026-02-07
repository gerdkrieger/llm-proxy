-- =============================================================================
-- ENTERPRISE STANDARD CONTENT FILTERS
-- =============================================================================
-- Datum: 2026-02-07
-- Zweck: Umfassende Filter für Unternehmensdaten gemäß DSGVO/GDPR
-- =============================================================================

-- Hinweis: Führe dieses Script aus um alle Standard-Filter zu erstellen
-- Bestehende Filter (IDs 1-6) werden nicht überschrieben

BEGIN;

-- =============================================================================
-- KATEGORIE 1: PERSONENBEZOGENE DATEN (DSGVO Art. 4 Nr. 1)
-- =============================================================================

-- Filter ID 7: Deutsche IBAN
INSERT INTO content_filters (id, filter_type, pattern, replacement, description, case_sensitive, enabled, priority, created_by)
VALUES (
    7,
    'regex',
    '\bDE[0-9]{2}\s?[0-9]{4}\s?[0-9]{4}\s?[0-9]{4}\s?[0-9]{4}\s?[0-9]{2}\b',
    '[IBAN-REDACTED]',
    'Filtert deutsche IBAN (Bankkontonummern)',
    false,
    true,
    90,
    'system'
) ON CONFLICT (id) DO NOTHING;

-- Filter ID 8: Internationale IBAN
INSERT INTO content_filters (id, filter_type, pattern, replacement, description, case_sensitive, enabled, priority, created_by)
VALUES (
    8,
    'regex',
    '\b[A-Z]{2}[0-9]{2}\s?[A-Z0-9]{4}\s?[A-Z0-9]{4}\s?[A-Z0-9]{4}\s?[A-Z0-9]{4}\s?[A-Z0-9]{0,2}\b',
    '[IBAN-REDACTED]',
    'Filtert internationale IBAN',
    false,
    true,
    90,
    'system'
) ON CONFLICT (id) DO NOTHING;

-- Filter ID 9: Kreditkartennummern
INSERT INTO content_filters (id, filter_type, pattern, replacement, description, case_sensitive, enabled, priority, created_by)
VALUES (
    9,
    'regex',
    '\b(?:4[0-9]{12}(?:[0-9]{3})?|5[1-5][0-9]{14}|3[47][0-9]{13}|3(?:0[0-5]|[68][0-9])[0-9]{11}|6(?:011|5[0-9]{2})[0-9]{12}|(?:2131|1800|35\d{3})\d{11})\b',
    '[KREDITKARTE-REDACTED]',
    'Filtert Kreditkartennummern (Visa, Mastercard, Amex, Discover, JCB)',
    false,
    true,
    100,
    'system'
) ON CONFLICT (id) DO NOTHING;

-- Filter ID 10: Sozialversicherungsnummer (Deutschland)
INSERT INTO content_filters (id, filter_type, pattern, replacement, description, case_sensitive, enabled, priority, created_by)
VALUES (
    10,
    'regex',
    '\b[0-9]{2}\s?[0-9]{6}\s?[A-Z]\s?[0-9]{3}\b',
    '[SV-NUMMER-REDACTED]',
    'Filtert deutsche Sozialversicherungsnummer',
    false,
    true,
    95,
    'system'
) ON CONFLICT (id) DO NOTHING;

-- Filter ID 11: Personalausweisnummer (Deutschland)
INSERT INTO content_filters (id, filter_type, pattern, replacement, description, case_sensitive, enabled, priority, created_by)
VALUES (
    11,
    'regex',
    '\b[CFGHJKLMNPRTVWXYZ][CFGHJKLMNPRTVWXYZ0-9]{8}[0-9]\b',
    '[AUSWEIS-REDACTED]',
    'Filtert deutsche Personalausweisnummer',
    false,
    true,
    95,
    'system'
) ON CONFLICT (id) DO NOTHING;

-- Filter ID 12: Reisepassnummer (Deutschland)
INSERT INTO content_filters (id, filter_type, pattern, replacement, description, case_sensitive, enabled, priority, created_by)
VALUES (
    12,
    'regex',
    '\b[CFGHJK][0-9]{8}\b',
    '[PASS-REDACTED]',
    'Filtert deutsche Reisepassnummer',
    false,
    true,
    95,
    'system'
) ON CONFLICT (id) DO NOTHING;

-- Filter ID 13: Geburtsdatum (verschiedene Formate)
INSERT INTO content_filters (id, filter_type, pattern, replacement, description, case_sensitive, enabled, priority, created_by)
VALUES (
    13,
    'regex',
    '\b(?:0[1-9]|[12][0-9]|3[01])\.(?:0[1-9]|1[012])\.(?:19|20)[0-9]{2}\b',
    '[GEBURTSDATUM-REDACTED]',
    'Filtert Geburtsdatum im Format DD.MM.YYYY',
    false,
    true,
    85,
    'system'
) ON CONFLICT (id) DO NOTHING;

-- Filter ID 14: Postleitzahlen (Deutschland)
INSERT INTO content_filters (id, filter_type, pattern, replacement, description, case_sensitive, enabled, priority, created_by)
VALUES (
    14,
    'regex',
    '\b[0-9]{5}\b(?=\s+[A-ZÄÖÜ])',
    '[PLZ-REDACTED]',
    'Filtert deutsche Postleitzahlen (wenn vor Stadtnamen)',
    false,
    true,
    60,
    'system'
) ON CONFLICT (id) DO NOTHING;

-- Filter ID 15: Straßenadressen (Deutschland)
INSERT INTO content_filters (id, filter_type, pattern, replacement, description, case_sensitive, enabled, priority, created_by)
VALUES (
    15,
    'regex',
    '\b[A-ZÄÖÜ][a-zäöüß]+(?:straße|str\.|strasse|weg|platz|allee|gasse)\s+[0-9]{1,4}[a-z]?\b',
    '[ADRESSE-REDACTED]',
    'Filtert deutsche Straßenadressen mit Hausnummer',
    false,
    true,
    75,
    'system'
) ON CONFLICT (id) DO NOTHING;

-- =============================================================================
-- KATEGORIE 2: FINANZDATEN
-- =============================================================================

-- Filter ID 16: Gehaltsinformationen (Euro)
INSERT INTO content_filters (id, filter_type, pattern, replacement, description, case_sensitive, enabled, priority, created_by)
VALUES (
    16,
    'regex',
    '\b(?:Gehalt|Lohn|Vergütung|Einkommen|Brutto|Netto):\s*(?:€|EUR)?\s*[0-9]{1,3}(?:\.[0-9]{3})*(?:,[0-9]{2})?\b',
    '[GEHALT-REDACTED]',
    'Filtert Gehaltsangaben in Euro',
    false,
    true,
    90,
    'system'
) ON CONFLICT (id) DO NOTHING;

-- Filter ID 17: Umsatzzahlen
INSERT INTO content_filters (id, filter_type, pattern, replacement, description, case_sensitive, enabled, priority, created_by)
VALUES (
    17,
    'regex',
    '\b(?:Umsatz|Gewinn|Verlust|Bilanz):\s*(?:€|EUR)?\s*[0-9]{1,3}(?:\.[0-9]{3})*(?:,[0-9]{2})?\s*(?:Mio|Mrd|Million|Milliarde)?\b',
    '[UMSATZ-REDACTED]',
    'Filtert Umsatz- und Bilanzzahlen',
    false,
    true,
    85,
    'system'
) ON CONFLICT (id) DO NOTHING;

-- Filter ID 18: Steuernummer (Deutschland)
INSERT INTO content_filters (id, filter_type, pattern, replacement, description, case_sensitive, enabled, priority, created_by)
VALUES (
    18,
    'regex',
    '\b[0-9]{2,3}/[0-9]{3}/[0-9]{5}\b',
    '[STEUERNUMMER-REDACTED]',
    'Filtert deutsche Steuernummer',
    false,
    true,
    90,
    'system'
) ON CONFLICT (id) DO NOTHING;

-- Filter ID 19: USt-IdNr. (Deutschland)
INSERT INTO content_filters (id, filter_type, pattern, replacement, description, case_sensitive, enabled, priority, created_by)
VALUES (
    19,
    'regex',
    '\bDE[0-9]{9}\b',
    '[UST-ID-REDACTED]',
    'Filtert deutsche Umsatzsteuer-Identifikationsnummer',
    false,
    true,
    90,
    'system'
) ON CONFLICT (id) DO NOTHING;

-- =============================================================================
-- KATEGORIE 3: KUNDENDATEN
-- =============================================================================

-- Filter ID 20: Kundennummer (generisch)
INSERT INTO content_filters (id, filter_type, pattern, replacement, description, case_sensitive, enabled, priority, created_by)
VALUES (
    20,
    'regex',
    '\b(?:Kunden-?Nr\.?|Kundennummer|KD-?Nr\.?):\s*[A-Z0-9]{6,12}\b',
    '[KUNDENNR-REDACTED]',
    'Filtert Kundennummern',
    false,
    true,
    80,
    'system'
) ON CONFLICT (id) DO NOTHING;

-- Filter ID 21: Auftragsnummer
INSERT INTO content_filters (id, filter_type, pattern, replacement, description, case_sensitive, enabled, priority, created_by)
VALUES (
    21,
    'regex',
    '\b(?:Auftrags-?Nr\.?|Auftragsnummer|Bestell-?Nr\.?|Bestellnummer|Order-?ID):\s*[A-Z0-9]{6,15}\b',
    '[AUFTRAGSNR-REDACTED]',
    'Filtert Auftrags- und Bestellnummern',
    false,
    true,
    75,
    'system'
) ON CONFLICT (id) DO NOTHING;

-- Filter ID 22: Rechnungsnummer
INSERT INTO content_filters (id, filter_type, pattern, replacement, description, case_sensitive, enabled, priority, created_by)
VALUES (
    22,
    'regex',
    '\b(?:Rechnungs-?Nr\.?|Rechnungsnummer|Invoice-?ID):\s*[A-Z0-9]{6,15}\b',
    '[RECHNUNGSNR-REDACTED]',
    'Filtert Rechnungsnummern',
    false,
    true,
    80,
    'system'
) ON CONFLICT (id) DO NOTHING;

-- Filter ID 23: Personennamen (Titel + Name)
INSERT INTO content_filters (id, filter_type, pattern, replacement, description, case_sensitive, enabled, priority, created_by)
VALUES (
    23,
    'regex',
    '\b(?:Herr|Frau|Dr\.|Prof\.|Dipl\.-Ing\.)\s+[A-ZÄÖÜ][a-zäöüß]+(?:\s+[A-ZÄÖÜ][a-zäöüß]+){1,2}\b',
    '[PERSON-REDACTED]',
    'Filtert Personennamen mit Titel',
    false,
    true,
    70,
    'system'
) ON CONFLICT (id) DO NOTHING;

-- =============================================================================
-- KATEGORIE 4: UNTERNEHMENSDATEN
-- =============================================================================

-- Filter ID 24: Handelsregisternummer
INSERT INTO content_filters (id, filter_type, pattern, replacement, description, case_sensitive, enabled, priority, created_by)
VALUES (
    24,
    'regex',
    '\bHRB?\s*[0-9]{4,6}\s*[A-Z]?\b',
    '[HRB-REDACTED]',
    'Filtert Handelsregisternummer',
    false,
    true,
    85,
    'system'
) ON CONFLICT (id) DO NOTHING;

-- Filter ID 25: Interne Projekt-IDs
INSERT INTO content_filters (id, filter_type, pattern, replacement, description, case_sensitive, enabled, priority, created_by)
VALUES (
    25,
    'regex',
    '\b(?:PROJ|PRJ|PROJECT)-[0-9]{4,6}\b',
    '[PROJEKT-REDACTED]',
    'Filtert interne Projekt-IDs',
    false,
    true,
    75,
    'system'
) ON CONFLICT (id) DO NOTHING;

-- Filter ID 26: Vertrags-IDs
INSERT INTO content_filters (id, filter_type, pattern, replacement, description, case_sensitive, enabled, priority, created_by)
VALUES (
    26,
    'regex',
    '\b(?:Vertrags-?Nr\.?|Vertragsnummer|Contract-?ID):\s*[A-Z0-9]{6,15}\b',
    '[VERTRAG-REDACTED]',
    'Filtert Vertragsnummern',
    false,
    true,
    80,
    'system'
) ON CONFLICT (id) DO NOTHING;

-- =============================================================================
-- KATEGORIE 5: CREDENTIALS & SECRETS
-- =============================================================================

-- Filter ID 27: API Keys (Generic)
INSERT INTO content_filters (id, filter_type, pattern, replacement, description, case_sensitive, enabled, priority, created_by)
VALUES (
    27,
    'regex',
    '\b(?:api[_-]?key|apikey|api[_-]?secret)[\s:=]+[''"]?[A-Za-z0-9_\-]{20,}[''"]?\b',
    '[API-KEY-REDACTED]',
    'Filtert API Keys und Secrets',
    false,
    true,
    100,
    'system'
) ON CONFLICT (id) DO NOTHING;

-- Filter ID 28: Passwörter
INSERT INTO content_filters (id, filter_type, pattern, replacement, description, case_sensitive, enabled, priority, created_by)
VALUES (
    28,
    'regex',
    '\b(?:password|passwort|pwd)[\s:=]+[''"]?[^\s]{6,}[''"]?\b',
    '[PASSWORD-REDACTED]',
    'Filtert Passwörter aus Text',
    false,
    true,
    100,
    'system'
) ON CONFLICT (id) DO NOTHING;

-- Filter ID 29: Private Keys (PEM Format)
INSERT INTO content_filters (id, filter_type, pattern, replacement, description, case_sensitive, enabled, priority, created_by)
VALUES (
    29,
    'regex',
    '-----BEGIN\s+(?:RSA\s+)?PRIVATE\s+KEY-----[\s\S]+?-----END\s+(?:RSA\s+)?PRIVATE\s+KEY-----',
    '[PRIVATE-KEY-REDACTED]',
    'Filtert Private Keys im PEM Format',
    false,
    true,
    100,
    'system'
) ON CONFLICT (id) DO NOTHING;

-- Filter ID 30: AWS Access Keys
INSERT INTO content_filters (id, filter_type, pattern, replacement, description, case_sensitive, enabled, priority, created_by)
VALUES (
    30,
    'regex',
    '\b(?:AKIA|A3T|AGPA|AIDA|AROA|AIPA|ANPA|ANVA|ASIA)[A-Z0-9]{16}\b',
    '[AWS-KEY-REDACTED]',
    'Filtert AWS Access Keys',
    false,
    true,
    100,
    'system'
) ON CONFLICT (id) DO NOTHING;

-- Filter ID 31: JWT Tokens
INSERT INTO content_filters (id, filter_type, pattern, replacement, description, case_sensitive, enabled, priority, created_by)
VALUES (
    31,
    'regex',
    '\beyJ[A-Za-z0-9_\-]*\.eyJ[A-Za-z0-9_\-]*\.[A-Za-z0-9_\-]*\b',
    '[JWT-TOKEN-REDACTED]',
    'Filtert JWT Tokens',
    false,
    true,
    100,
    'system'
) ON CONFLICT (id) DO NOTHING;

-- =============================================================================
-- KATEGORIE 6: MEDIZINISCHE DATEN (DSGVO Art. 9 - besondere Kategorien)
-- =============================================================================

-- Filter ID 32: Krankenversicherungsnummer
INSERT INTO content_filters (id, filter_type, pattern, replacement, description, case_sensitive, enabled, priority, created_by)
VALUES (
    32,
    'regex',
    '\b[A-Z][0-9]{9}\b',
    '[KV-NUMMER-REDACTED]',
    'Filtert Krankenversicherungsnummer',
    false,
    true,
    95,
    'system'
) ON CONFLICT (id) DO NOTHING;

-- Filter ID 33: Patientennummer
INSERT INTO content_filters (id, filter_type, pattern, replacement, description, case_sensitive, enabled, priority, created_by)
VALUES (
    33,
    'regex',
    '\b(?:Patienten-?Nr\.?|Patientennummer):\s*[0-9]{6,10}\b',
    '[PATIENT-REDACTED]',
    'Filtert Patientennummern',
    false,
    true,
    95,
    'system'
) ON CONFLICT (id) DO NOTHING;

-- =============================================================================
-- KATEGORIE 7: IP-ADRESSEN & NETZWERK
-- =============================================================================

-- Filter ID 34: IPv4 Adressen
INSERT INTO content_filters (id, filter_type, pattern, replacement, description, case_sensitive, enabled, priority, created_by)
VALUES (
    34,
    'regex',
    '\b(?:10\.|172\.(?:1[6-9]|2[0-9]|3[01])\.|192\.168\.)[0-9]{1,3}\.[0-9]{1,3}\b',
    '[IP-REDACTED]',
    'Filtert private IPv4 Adressen',
    false,
    true,
    65,
    'system'
) ON CONFLICT (id) DO NOTHING;

-- Filter ID 35: MAC Adressen
INSERT INTO content_filters (id, filter_type, pattern, replacement, description, case_sensitive, enabled, priority, created_by)
VALUES (
    35,
    'regex',
    '\b([0-9A-Fa-f]{2}[:-]){5}([0-9A-Fa-f]{2})\b',
    '[MAC-REDACTED]',
    'Filtert MAC Adressen',
    false,
    true,
    70,
    'system'
) ON CONFLICT (id) DO NOTHING;

-- =============================================================================
-- KATEGORIE 8: KOMMUNIKATIONSDATEN
-- =============================================================================

-- Filter ID 36: Mobile Telefonnummer (Deutschland)
INSERT INTO content_filters (id, filter_type, pattern, replacement, description, case_sensitive, enabled, priority, created_by)
VALUES (
    36,
    'regex',
    '\b(?:\+49|0049|0)\s*1[5-7][0-9]\s*[0-9]{7,8}\b',
    '[MOBILNUMMER-REDACTED]',
    'Filtert deutsche Mobilfunknummern',
    false,
    true,
    85,
    'system'
) ON CONFLICT (id) DO NOTHING;

-- Filter ID 37: Faxnummer
INSERT INTO content_filters (id, filter_type, pattern, replacement, description, case_sensitive, enabled, priority, created_by)
VALUES (
    37,
    'regex',
    '\b(?:Fax|Telefax):\s*(?:\+49|0)\s*[0-9]{2,5}\s*[0-9]{4,10}\b',
    '[FAX-REDACTED]',
    'Filtert Faxnummern',
    false,
    true,
    75,
    'system'
) ON CONFLICT (id) DO NOTHING;

-- =============================================================================
-- KATEGORIE 9: INTERNATIONALE DATEN
-- =============================================================================

-- Filter ID 38: US Social Security Number
INSERT INTO content_filters (id, filter_type, pattern, replacement, description, case_sensitive, enabled, priority, created_by)
VALUES (
    38,
    'regex',
    '\b[0-9]{3}-[0-9]{2}-[0-9]{4}\b',
    '[SSN-REDACTED]',
    'Filtert US Social Security Numbers',
    false,
    true,
    95,
    'system'
) ON CONFLICT (id) DO NOTHING;

-- Filter ID 39: UK National Insurance Number
INSERT INTO content_filters (id, filter_type, pattern, replacement, description, case_sensitive, enabled, priority, created_by)
VALUES (
    39,
    'regex',
    '\b[A-CEGHJ-PR-TW-Z]{1}[A-CEGHJ-NPR-TW-Z]{1}[0-9]{6}[A-D]{1}\b',
    '[NINO-REDACTED]',
    'Filtert UK National Insurance Number',
    false,
    true,
    95,
    'system'
) ON CONFLICT (id) DO NOTHING;

-- Filter ID 40: Schweizer AHV-Nummer
INSERT INTO content_filters (id, filter_type, pattern, replacement, description, case_sensitive, enabled, priority, created_by)
VALUES (
    40,
    'regex',
    '\b756\.[0-9]{4}\.[0-9]{4}\.[0-9]{2}\b',
    '[AHV-REDACTED]',
    'Filtert Schweizer AHV-Nummer',
    false,
    true,
    95,
    'system'
) ON CONFLICT (id) DO NOTHING;

-- =============================================================================
-- Update Sequenz
-- =============================================================================
-- Stelle sicher dass die nächste ID nach den eingefügten Filtern kommt
SELECT setval('content_filters_id_seq', (SELECT MAX(id) FROM content_filters) + 1);

COMMIT;

-- =============================================================================
-- Verification
-- =============================================================================
SELECT COUNT(*) as total_filters FROM content_filters;
SELECT filter_type, COUNT(*) as count FROM content_filters GROUP BY filter_type ORDER BY count DESC;

-- Zeige alle neuen Filter
SELECT id, filter_type, description, enabled 
FROM content_filters 
WHERE id >= 7 
ORDER BY id;
