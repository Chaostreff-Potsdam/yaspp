# Yet another simple podcast publisher

A really simple podcast publisher (a.k.a. the generator behing [Chaos Radio Potsdam](https://radio.ccc-p.org)).

Generates a static feed and podlove-webplayer list from a given yaml-file.
To adapt for your page edit `templates.py` and `config.py`.

After editing your yaml-file run:

	./yaspp.py [-o output_dir] <CONTENT.YAML>

