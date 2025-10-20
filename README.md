# Yet another simple podcast publisher

A really simple podcast publisher (a.k.a. the generator behind [Chaos Radio Potsdam](https://radio.ccc-p.org)).

Generates a static feed and podlove-webplayer list from a given yaml-file.
To adapt for your page edit `templates.py` and `config.py`.

After editing your yaml-file run:

	./yaspp.py [-o output_dir] <CONTENT.YAML>


## Peparing your podcast files

Modern podcasts peer pressure promotes per-episode pictures and chapter marks.
For legacy clients, those should be embedded in the audio file, not only the feed.

How to do this with [ffmpeg](https://ffmpeg.org/) and [mp3chaps](https://pypi.org/project/mp3chaps/):

	export NT_DATE=<yyyy-mm-dd>
	export NT_YEAR=<yyyy>
	export BITRATE=121k  # Up for discussion ;)

	ffmpeg -i <inputfile.mp3> -i cover.jpg -b:a $BITRATE -c:v mjpeg -map 0:0 -map 1:0 -id3v2_version 3 \
		-metadata album="Nerdtalk"  -metadata genre="Podcast" -metadata title="CiR am $NT_DATE" \
		-metadata artist="Chaostreff Potsdam" -metadata publisher="Chaostreff Potsdam" \
		-metadata date=$NT_YEAR \
		-metadata:s:v title="Album cover" -metadata:s:v comment="Cover (front)" <outputfile.mp3>
	mp3chaps -i <outputfile.mp3>  # While having a <outputfile.chapters.txt>

## Workflow Vodoo

Um die Publikation der Episode vorzubereiten, 
1. füllt bitte im Pad die Sections `## Summary`, `## Shownotes`,`## Mukke` und optional `## Kapitel` mit Inhalt. Alle anderen Sections sind für die Publikation nicht relevant. 
Insbesondere die Links zur Musik sollten stimmen ("by Attribution"-Lizenz und so).
1. Triggert einen Admin um die Episode an [die richtige Stelle](https://radio.ccc-p.org/files) zu schieben. (geht hoffentlich bald automatisch)
1. Wenn fertig, dann setzt das Tag `shownotes_complete` oben im Pad und klickt auf [Run Workflow](https://github.com/Chaostreff-Potsdam/yaspp/actions/workflows/create-pr-from-pad.yml) (Rechte im GitHub Repo nötig)
1. Jeder Pull Request [generiert](https://github.com/Chaostreff-Potsdam/yaspp/actions/workflows/docker-build-and-run.yml) ein zip mit Preview der Webseite (zu finden unter Artifacts), das muss auf'm Server an der richtigen Stelle entpackt werden (geht hoffentlich bald automatisch)

### Stable Download Link

For automated deployments, you can always download the latest build from the master branch using this stable URL:

```
https://github.com/Chaostreff-Potsdam/yaspp/releases/download/latest/website.tar.gz
```

This tarball contains `index.html`, `feed.xml`, and `cccp.css` from the latest master build.

Example usage:
```bash
curl -L -o website.tar.gz https://github.com/Chaostreff-Potsdam/yaspp/releases/download/latest/website.tar.gz
tar -xzf website.tar.gz
```

Erzeugt keine Subsections innerhalb der Shownotes-Sections (sieht auf der Webseite komisch aus) und fasst euch kurz.
