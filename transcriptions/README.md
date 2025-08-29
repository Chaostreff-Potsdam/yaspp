# Transkriptionen

Dieses Verzeichnis enthält die automatisierten Transkriptionen aller Podcast-Episoden von "Chaos im Radio" seit Juli 2022.

## Wichtige Hinweise

⚠️ **Diese Transkriptionen wurden automatisiert erstellt und sind nicht korrekturgelesen!**

- Die Transkriptionen wurden mittels automatischer Spracherkennung generiert (siehe `transcribe-folder.sh`)
- Es können Fehler bei der Erkennung von Namen, Fachbegrffen oder undeutlich gesprochenen Wörtern auftreten
- Die Transkriptionen dienen als Hilfsmittel und repräsentieren das gesprochene Wort im Podcast
- Zeitstempel können geringfügig ungenau sein

## Verfügbare Skripte

### `transcribe-folder.sh`
Skript zur automatischen Transkription neuer Audio-Dateien mit whisper.cpp

### `fix-mistakes.sh`
Ein Bash-Skript zur automatischen Korrektur häufiger Spracherkennungsfehler:
- Korrigiert falsch erkannte Namen der Moderator:innen (z.B. "Syrux" → "Cyroxx")
- Behebt typische Erkennungsfehler bei häufig verwendeten Wörtern

## Dateiformat

Die Transkriptionen liegen im WebVTT-Format (`.vtt`) vor, welches sowohl Zeitstempel als auch den transkribierten Text enthält. Dieses Format ist mit den meisten Video- und Audio-Playern kompatibel und ermöglicht die Anzeige von Untertiteln.

## Mitwirkende

Die Stimmen in den Transkriptionen gehören zu den regelmäßigen Moderator:innen:
- Cyroxx
- Hannes
- Gini
- Knurps
- Ajuvo

---

*Diese README wurde erstellt, um Transparenz über die Entstehung und Qualität der automatisierten Transkriptionen zu schaffen.*
