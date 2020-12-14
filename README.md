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

	ffmpeg -i <inputfile.mp3> -i cover.jpg -b:a <bitrate>k -c:v mjpeg -map 0:0 -map 1:0 -id3v2_version 3 -metadata:s:v title="Album cover" -metadata:s:v comment="Cover (front) <outputfile.mp3>
	mp3chaps -i <outputfile.mp3>  # While having a <outputfile.chapters.mp3>

