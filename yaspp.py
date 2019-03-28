#!/usr/bin/env python

import json
import os
import string

import podgen
import yaml

import config
import templates

def read_content_yaml(filename):
	for entry in yaml.load_all(open(filename), Loader=yaml.Loader):
		for audio in entry["audio"]:
			audio["url"] = string.Template(audio["url"]).substitute(
					media_base_url=config.media_base_url)
		yield entry


def generate_html_entry(entryid, entry):
	entrydivid = "yassp_entry_%d" % entryid

	podlove_player = "<script>podlovePlayer('#%s', %s);</script>" % \
			(entrydivid, json.dumps(entry))

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
			content=content_list
		)


def generate_feed(content):
	return ""


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
