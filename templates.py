import string

index_html = string.Template(r"""<!DOCTYPE html>
<html>

<head>

<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0, user-scalable=yes">

<link rel="stylesheet" type="text/css" href="macdown.css" />	
<link rel="stylesheet" type="text/css" href="yaspp.css" />	

<script src="https://cdn.podlove.org/web-player/embed.js"></script>

</head>

<body>

<h1 id="toc_0">$podcast_title</h1>

<div>
	<div class="introblock">
		$hello_text
	</div>
	<div class="buttonblock">
		$subscribe_button
	</div>
</div>

$content

</body>

</html>""")


entry = string.Template(r"""
<h2 id="toc_2">$title</h2>

<p>$summary</p>

<p><div id="$entrydivid" /></p>
""")

subscribe_button = string.Template(r"""
	<script>window.podcastData={
		"title": "$title",
		"subtitle": "",
		"description": "$description",
		"cover": "$cover",
		"feeds": [{"type":"audio","format":"mp3","url":"$feed_url"}]}
	</script>
	<script class="podlove-subscribe-button" src="https://cdn.podlove.org/subscribe-button/javascripts/app.js" data-language="en" data-size="medium" data-json-data="podcastData" data-color="#469cd1" data-format="cover" data-style="outline"></script><noscript><a href="$feed_url">Subscribe to feed</a></noscript>
""")
