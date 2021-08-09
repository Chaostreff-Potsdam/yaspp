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

	export NT_DATE=<yyyy-mm-ddd>
	export NT_YEAR=<yyyy>
	export BITRATE=121k  # Up for discussion ;)

	ffmpeg -i <inputfile.mp3> -i cover.jpg -b:a $BITRATE -c:v mjpeg -map 0:0 -map 1:0 -id3v2_version 3 \
		-metadata album="Nerdtalk"  -metadata genre="Podcast" -metadata title="CiR am $NT_DATE" \
		-metadata artist="Chaostreff Potsdam" -metadata publisher="Chaostreff Potsdam" \
		-metadata date=$NT_YEAR \
		-metadata:s:v title="Album cover" -metadata:s:v comment="Cover (front)" <outputfile.mp3>
	mp3chaps -i <outputfile.mp3>  # While having a <outputfile.chapters.txt>

