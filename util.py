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

lines = INPUT.splitlines()

for line in lines:
    chapter_start, chapter_title = line.split(' ', 1)

    template = f"""  - start: '{chapter_start}'\n{' '*4}title: '{chapter_title}'"""

    print(template)
