# Small utility to convert the chapter marks into YAML markup

import re

# TODO: use a chapter file instead
INPUT = """00:00:00 Intro mit Hitze, Schulanfang und SUP
00:05:24 Musik: Black Bones - Captain Blood (CC-BY-NC-SA 3.0 US) <https://freemusicarchive.org/music/Black_Bones/Pirates_of_the_Coast/08_-_Black_Bones_-_Captain_Blood>
00:09:32 Update CE-Zeichen
00:15:14 Update SpaceSchalter
00:21:19 Was tun, bevor bzw. nachdem das eigene Smartphone geklaut wird?
00:43:13 Open Health HACKademy
00:53:53 Outro
00:54:58 Musik: Roman Meleshin - Delusion (CC-BY 3.0) <https://meleshin.bandcamp.com/track/delusion-electronic-creative-commons-by-license>"""

YAML_INPUT = """
chapters:
  - start: '00:00:00.000'
    title: 'Intro'
  - start: '00:00:49.000'
    title: 'Musik: Professor Kliq - Bust This Bust That'
    href: 'https://freemusicarchive.org/music/Professor_Kliq/Bust_This_Bust_That/Bust_This_Bust_That'
  - start: '00:05:03.000'
    title: 'Aktuelles vom Chaostreff'
  - start: '00:09:44.000'
    title: 'Corona-Warn-App'
  - start: '00:28:40.700'
    title: 'Musik: Organism - Space Funk'
    href: 'https://freemusicarchive.org/music/Organism/S27-X_II/Organism_-S27-X_II-_19_Space_Funk'
  - start: '00:33:59.000'
    title: 'Kurzmeldungen'
  - start: '00:58:30.000'
    title: 'Erdbeerschnitzel - Walkampfchampagne'
    href: 'https://freemusicarchive.org/music/Erdbeerschnitzel/Pathetik_Party/02-Walkampfchampagne'
"""

def generate_yaml(content):
    lines = content.splitlines()
    for line in lines:
        chapter_start, chapter_title = line.split(' ', 1)

        # Extract hyperlink if present
        href = ''
        if matches := re.search(r' <([^ ]+)>$', chapter_title):
            href = f"\n{' ' * 4}href:  '{ matches.group(1) }'"
            chapter_title = chapter_title[:matches.start()]

        template = f"""  - start: '{chapter_start}'\n{' ' * 4}title: '{chapter_title}'{href}"""

        print(template)


def generate_chapter_file():
    import yaml

    content = yaml.safe_load(YAML_INPUT)

    for chapter in content['chapters']:
        template = f"{chapter['start']} {chapter['title']}"

        if chapter.get('href'):
            template += f" <{chapter['href']}>"

        print(template)

    #read_content_yaml()

if __name__ == '__main__':
    generate_yaml(INPUT)
    #generate_chapter_file()

