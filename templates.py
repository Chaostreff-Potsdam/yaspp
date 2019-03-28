import string

index_html = string.Template(r"""<!DOCTYPE html>
<html>

<head>

<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0, user-scalable=yes">

<link rel="stylesheet" type="text/css" href="macdown.css"/>	

<script src="https://cdn.podlove.org/web-player/embed.js"></script>

</head>

<body>

<h1 id="toc_0">$podcast_title</h1>

$hello_text

$content

</body>

</html>""")


entry = string.Template(r"""
<h2 id="toc_2">$title</h2>

<p>$summary</p>

<p><div id="$entrydivid" /></p>
""")

