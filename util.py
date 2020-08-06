# Small utility to convert the chapter marks into YAML markup

# TODO: use a chapter file instead
INPUT = """00:00:00.000 Intro
00:00:54.114 Citrix-Update
00:07:08.971 Bundestagshack:Haftbefehl gegen russischen Hacker
00:09:04.332 Gesetze für mehr angebliche Sicherheit
00:22:06.578 Musik: Virgill - Computers are gay
00:25:32.756 Gesichtsvisiere im Museum
00:27:53.515 Patentendatenschutzgesetz
00:34:23.902 IT-Sicherheit per Gesetz
00:40:01.013 Corona-Warn-App
00:43:55.898 Musik: !Cube - My Pixels are Weapons
00:47:29.475 Jugendmedienschutzstaatsvertrag
00:55:06.622 Abschluss mit Hotfix
00:56:34.209 Musik: !Cube - Glittering Waves"""

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

        template = f"""  - start: '{chapter_start}'\n{' ' * 4}title: '{chapter_title}'"""

        print(template)

def generate_chapter_file():
    import yaml

    content = yaml.safe_load(YAML_INPUT)

    for chapter in content['chapters']:
        template = f"{chapter['start']} {chapter['title']}"

        print(template)

    #read_content_yaml()

if __name__ == '__main__':
    #generate_yaml(INPUT)
    generate_chapter_file()

