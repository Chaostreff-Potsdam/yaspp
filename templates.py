import string

index_html = string.Template(r"""<!DOCTYPE html>
<html lang="de">

<head>

<title>$podcast_title</title>

<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0, user-scalable=yes">

<link rel="stylesheet" type="text/css" href="cccp.css" />	
<link rel="stylesheet" type="text/css" href="yaspp.css" />	

<script src="https://cdn.podlove.org/web-player/embed.js"></script>

</head>

<body>

<header>

<div class="container introcontainer">
	<h1 id="toc_0">$podcast_title</h1>
	<div class="introblock">
		$hello_text
	</div>
	<div class="buttonblock">
		$subscribe_button
	</div>
</div>

</header>

<div class=container>
<section id="main_content">

$content

</section>
</div>

<h2></h2>

<footer>
	<center>
		<small>
			$footer_text
		</small>
	</center>
</footer>

</body>

</html>""")


entry = string.Template(r"""
<div id="$entrydivid" class="yaspp-entry">
<a href="#$uuid" style="text-decoration: none;"><h2 id="$uuid">$title</h2></a>

<p>$summary</p>

<div id="player_$entrydivid"></div>
<div id="shownotes_$entrydivid" class="yaspp-shownotes">
$long_summary
</div>
</div>
""")

subscribe_button = string.Template(r"""
	<script>window.podcastData={
		"title": "$title",
		"subtitle": "",
		"description": $description,
		"cover": "$cover",
		"feeds": [{"type":"audio","format":"mp3","url":"$feed_url"}]}
	</script>
	<script class="podlove-subscribe-button" src="https://cdn.podlove.org/subscribe-button/javascripts/app.js" data-language="en" data-size="medium" data-json-data="podcastData" data-color="#b5e853" data-format="cover" data-style="filled"></script><noscript><a href="$feed_url">Subscribe to feed</a></noscript>
""")
