#!/usr/bin/env python3
"""Genera docs/design/src/icons.js: iconos de archivo/carpeta como trazados vectoriales.

Los iconos son glifos de Nerd Fonts (familias Seti y Material Design). En el editor real se imprimen
como caracteres con una Nerd Font instalada; aqui se convierten a trazados SVG para poder dibujarlos
en OpenPencil sin instalar ninguna fuente.

Uso:
  python3 docs/design/tools/make-icons.py <SymbolsNerdFont-Regular.ttf> <glyphnames.json>

Descargas (version 3.x de Nerd Fonts):
  https://github.com/ryanoasis/nerd-fonts/releases/latest/download/NerdFontsSymbolsOnly.tar.xz
  https://raw.githubusercontent.com/ryanoasis/nerd-fonts/master/glyphnames.json
Requiere: pip install fonttools
"""
import json, os, sys
from fontTools.ttLib import TTFont
from fontTools.pens.svgPathPen import SVGPathPen
from fontTools.pens.transformPen import TransformPen
from fontTools.pens.boundsPen import BoundsPen

BOX = 13.0  # lado maximo del icono en px (la celda de terminal mide 9x18, la fuente 15px)

# clave: (glifo, color del tema, [nombres exactos], [extensiones])
SET = {
    'folder':      ('md-folder_outline',        'dim',    [], []),
    'folder_open': ('md-folder_open_outline',   'blue',   [], []),
    'default':     ('md-file_outline',          'dim',    [], []),
    'go':          ('seti-go',                  'teal',   ['go.mod', 'go.sum', 'go.work'], ['go']),
    'rust':        ('seti-rust',                'peach',  ['Cargo.toml', 'Cargo.lock'], ['rs']),
    'python':      ('seti-python',              'yellow', [], ['py', 'pyi']),
    'lua':         ('seti-lua',                 'blue',   [], ['lua']),
    'js':          ('md-language_javascript',   'yellow', [], ['js', 'mjs', 'cjs']),
    'ts':          ('md-language_typescript',   'blue',   [], ['ts', 'mts', 'cts']),
    'react':       ('seti-react',               'teal',   [], ['jsx', 'tsx']),
    'java':        ('seti-java',                'peach',  [], ['java']),
    'c':           ('seti-c',                   'blue',   [], ['c', 'h']),
    'cpp':         ('seti-cpp',                 'mauve',  [], ['cpp', 'cc', 'cxx', 'hpp', 'hh']),
    'ruby':        ('seti-ruby',                'red',    ['Gemfile'], ['rb', 'gemspec']),
    'php':         ('seti-php',                 'mauve',  [], ['php']),
    'kotlin':      ('seti-kotlin',              'mauve',  [], ['kt', 'kts']),
    'swift':       ('seti-swift',               'peach',  [], ['swift']),
    'html':        ('seti-html',                'peach',  [], ['html', 'htm']),
    'css':         ('seti-css',                 'mauve',  [], ['css', 'less']),
    'sass':        ('seti-sass',                'red',    [], ['scss', 'sass']),
    'json':        ('seti-json',                'yellow', [], ['json', 'jsonc', 'json5']),
    'yaml':        ('md-format_list_bulleted',  'mauve',  [], ['yml', 'yaml']),
    'config':      ('seti-config',              'teal',   ['.editorconfig', '.env'], ['toml', 'ini', 'conf', 'cfg', 'env']),
    'markdown':    ('md-language_markdown',     'teal',   [], ['md', 'markdown', 'mdx']),
    'text':        ('seti-text',                'dim',    [], ['txt', 'rst', 'log']),
    'shell':       ('md-bash',                  'green',  ['.bashrc', '.zshrc', '.profile'], ['sh', 'bash', 'zsh', 'fish']),
    'docker':      ('seti-docker',              'blue',   ['Dockerfile', 'docker-compose.yml', 'compose.yaml', '.dockerignore'], []),
    'make':        ('md-hammer_wrench',         'peach',  ['Makefile', 'makefile', 'GNUmakefile'], ['mk']),
    'license':     ('md-scale_balance',         'yellow', ['LICENSE', 'LICENSE.md', 'LICENCE', 'COPYING'], []),
    'git':         ('seti-git',                 'peach',  ['.gitignore', '.gitattributes', '.gitmodules'], []),
    'lock':        ('seti-lock',                'dim',    [], ['lock']),
    'npm':         ('seti-npm',                 'red',    ['package.json', 'package-lock.json'], []),
    'db':          ('seti-db',                  'mauve',  [], ['sql', 'db', 'sqlite']),
    'image':       ('md-image',                 'green',  [], ['png', 'jpg', 'jpeg', 'gif', 'webp', 'ico', 'bmp']),
    'svg':         ('seti-svg',                 'peach',  [], ['svg']),
    'xml':         ('md-xml',                   'peach',  [], ['xml', 'xsd', 'plist']),
    'pdf':         ('md-file_pdf_box',          'red',    [], ['pdf']),
    'archive':     ('md-zip_box',               'dim',    [], ['zip', 'tar', 'gz', 'xz', '7z']),
}

def main(font_path, names_path):
    names = json.load(open(names_path, encoding='utf-8'))
    font = TTFont(font_path); cmap = font.getBestCmap(); gs = font.getGlyphSet()
    icons, by_name, by_ext, color = {}, {}, {}, {}
    for key, (glyph, col, exact, exts) in SET.items():
        cp = int(names[glyph]['code'], 16)
        gname = cmap[cp]
        bp = BoundsPen(gs); gs[gname].draw(bp); x0, y0, x1, y1 = bp.bounds
        s = BOX / max(x1 - x0, y1 - y0)
        pen = SVGPathPen(gs, ntos=lambda v: ('%.2f' % v).rstrip('0').rstrip('.'))
        gs[gname].draw(TransformPen(pen, (s, 0, 0, -s, -x0 * s, y1 * s)))
        icons[key] = {'d': pen.getCommands(), 'w': round((x1 - x0) * s, 2), 'h': round((y1 - y0) * s, 2),
                      'code': 'U+%04X' % cp, 'glyph': glyph}
        color[key] = col
        for n in exact: by_name[n] = key
        for e in exts: by_ext[e] = key
    out = os.path.join(os.path.dirname(__file__), '..', 'src', 'icons.js')
    with open(out, 'w', encoding='utf-8') as f:
        f.write('// GENERADO por tools/make-icons.py: no editar a mano.\n')
        f.write('// Glifos de Nerd Fonts (Seti y Material Design) convertidos a trazados. Ver docs/design/README.md.\n')
        f.write('const ICONS = ' + json.dumps(icons, separators=(',', ':')) + ';\n')
        f.write('const ICON_COLOR = ' + json.dumps(color, separators=(',', ':')) + ';\n')
        f.write('const ICON_BY_NAME = ' + json.dumps(by_name, separators=(',', ':')) + ';\n')
        f.write('const ICON_BY_EXT = ' + json.dumps(by_ext, separators=(',', ':')) + ';\n')
    print(len(icons), 'iconos,', len(by_name), 'nombres exactos,', len(by_ext), 'extensiones ->', os.path.normpath(out),
          '(%d bytes)' % os.path.getsize(out))

if __name__ == '__main__':
    if len(sys.argv) != 3: sys.exit(__doc__)
    main(sys.argv[1], sys.argv[2])
