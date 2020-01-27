#!/usr/bin/env python

import json
import os
import string

import podgen
import yaml

import config
import templates


#-------------------------------------------------------------------------------

import html.parser

class HTMLStripParser(html.parser.HTMLParser):
    def __init__(self):
        super(HTMLStripParser, self).__init__()
        self.result = []

    def handle_data(self, data):
        self.result.append(data)

    def get_text(self):
        return "".join(self.result)

def strip_html_tags(html):
	p = HTMLStripParser()
	p.feed(html)
	return p.get_text()

#-------------------------------------------------------------------------------

def read_content_yaml(filename):
	for entry in yaml.load_all(open(filename), Loader=yaml.Loader):
		for audio in entry["audio"]:
			audio["url"] = string.Template(audio["url"]).substitute(
					media_base_url=config.media_base_url)
		yield entry


def generate_subscribe_button(content):
	return templates.subscribe_button.substitute(
			title=config.podcast_title,
			description=repr(config.hello_text),
			cover=config.small_cover_image,
			feed_url=config.feed_url
		)


def generate_html_entry(entryid, entry):
	entrydivid = "yassp_entry_%d" % entryid
	clean_entry = entry.copy()
	try:
		clean_entry.pop("uuid")
	except KeyError:
		pass

	podlove_player = "<script>podlovePlayer('#%s', %s);</script>" % \
			(entrydivid, json.dumps(clean_entry))

	return templates.entry.substitute(
			entrydivid=entrydivid,
			title=entry["title"],
			summary=entry["summary"]
		) + podlove_player


def generate_html(content, rev=True):
	op = reversed if True else lambda x: x
	content_list = "\n".join(generate_html_entry(i, e)
								for i, e in enumerate(op(content)))

	return templates.index_html.substitute(
			podcast_title=config.podcast_title,
			hello_text=config.hello_text,
			subscribe_button=generate_subscribe_button(content),
			content=content_list
		)

#-------------------------------------------------------------------------------

def generate_feed(content, audio_idx=0):
	p = podgen.Podcast(
			name=config.podcast_title,
			description=config.hello_text,
			website=config.website,
			image=config.cover_image,
			language=config.language,
			copyright=config.copyright,
			authors=[podgen.Person(config.author_name)],
			category=podgen.Category(config.category),
			explicit=False
		)
	p.episodes = [podgen.Episode(
			id=entry["uuid"],
			title=entry["title"],
			summary=entry["summary"],
			long_summary=("<p>%s</p>\n%s" % (entry["summary"], entry["long_summary"])) if "long_summary" in entry else None,
			publication_date=entry["publicationDate"],
			media=podgen.Media.create_from_server_response(entry["audio"][audio_idx]["url"]),
		) for entry in content]
	return str(p)


#-------------------------------------------------------------------------------

def parseArgs():
	import argparse

	parser = argparse.ArgumentParser(
			description='Generate a static feed and podlove webplayer list.')

	parser.add_argument("yaml_file", type=str, help="The content file.")
	parser.add_argument("-o", "--output-dir", type=str, default=".",
			help="Output directory (default: .)")

	return parser.parse_args()


def main():
	args = parseArgs()
	content = list(read_content_yaml(args.yaml_file))
	with open(os.path.join(args.output_dir, "index.html"), "w") as outfile:
		outfile.writelines(generate_html(content))

	with open(os.path.join(args.output_dir, "feed.xml"), "w") as outfile:
		outfile.writelines(generate_feed(content))


if __name__ == "__main__":
	main()
