# Spezifikation — <Projektname>

> **Template-Hinweis.** Diese Datei ist eine Vorlage. Sie ist
> **technisch verbindlich, aber ohne Lastenheft-Änderung fortschreibbar**
> (siehe Spec-Stratifizierung in
> [`grundlagen/konventionen.md`](https://github.com/pt9912/ai-harness-course/blob/v3.5.1/kurs/de/grundlagen/konventionen.md#spec-stratifizierung)).
> Kopiere sie nach `spec/spezifikation.md`, ersetze `<Platzhalter>` und
> lösche diesen Block.

**Status:** Aktiv. **Letzte Änderung:** YYYY-MM-DD.

**Bezug zum Lastenheft:** Diese Spezifikation präzisiert die in
`spec/lastenheft.md` formulierten Anforderungen (`LH-*`-IDs). Bei
Konflikt gewinnt das Lastenheft.

---

## 1. Algorithmen und Datenflüsse

<!--
Wie wird die funktionale Anforderung *technisch* erfüllt? Pseudocode
oder Sequenzbeschreibung erlaubt; tatsächlicher Code gehört in
src/.

ID-Schema: <PREFIX>-FA-<NN>.<Buchstabe> für Verfeinerungen einzelner
Lastenheft-IDs, z.B. LH-FA-03.a für eine konkrete Algorithmus-Variante.
-->

### LH-FA-01.a — Algorithmus für <…>

**Eingabe:** <…>. **Ausgabe:** <…>. **Schritte:**

1. <…>
2. <…>

**Komplexität:** <O(n log n)>, **Fehlermodi:** <…>.

---

## 2. Datenstrukturen und Schemas

<!--
Konkrete JSON-Schemas, OpenAPI-Snippets, Protokoll-Definitionen.
Fortschreibbar ohne Lastenheft-Änderung, solange die Lastenheft-
Anforderung gewahrt bleibt.
-->

### <Datenstruktur 1>

```json
{
  "field": "type"
}
```

## 3. Defaults und Konstanten

<!-- Werte, die in Code fest sind. Die ADR, die einen Wert festlegt,
deklariert das aufwärts in ihrem Schärft:-Feld (Kurs §Referenz-Richtung) —
kein ADR-Rückzeiger hier. -->

| Name | Wert | Begründung |
|---|---|---|
| `MAX_BATCH_SIZE` | 100 | <…> |

## 4. Fehler-Codes und Logging-Felder

<!-- Verbindliche Codes und Felder. -->

| Code | Bedingung | Aktion |
|---|---|---|
| E001 | <…> | <…> |

## 5. Metriken und Tracing-Felder

<!--
Verbindliche OTel-Felder pro Span (siehe Kurs Modul 15).
-->

| Span | Pflicht-Attribute | Quelle |
|---|---|---|
| `<service>.<operation>` | `<feldname>`, `<feldname>` | <…> |

## 6. Externe Verträge

<!-- Schnittstellen zu Drittsystemen, mit Versionsannahme. -->

| System | Version | Vertrag-Datei |
|---|---|---|
| <…> | <…> | <Pfad> |

## 7. Historie

| Datum | Änderung | ADR |
|---|---|---|
| YYYY-MM-DD | Initial | — |
