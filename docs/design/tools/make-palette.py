#!/usr/bin/env python3
"""Genera docs/design/src/palette.js: la paleta propia en dos temas, derivada en OKLCH.

Idea: todos los acentos comparten luminosidad (L) y croma (C) y solo cambian de tono (H), de modo
que ninguno "grita" mas que otro. Las superficies comparten el tono de la marca con croma bajisimo.
Verifica el contraste (WCAG) y falla si algun par importante queda por debajo del minimo.

Uso:  python3 docs/design/tools/make-palette.py        (escribe src/palette.js)
Para cambiar la identidad ajusta BRAND_HUE y los tonos de ACCENT_HUES.
"""
import json, math, os, sys

BRAND_HUE = 268          # tono de las superficies (azul-violeta de tinta)
ACCENT_HUES = {         # tono de cada acento
    'blue': 248, 'green': 152, 'mauve': 305, 'peach': 52, 'yellow': 92, 'red': 18, 'teal': 196,
}
# Color de marca: verde fosforo (el de los monitores CRT). (L, C, H) por tema.
BRAND = {'dark': (0.86, 0.17, 153), 'light': (0.52, 0.15, 153)}
THEMES = {
    'dark': {
        'surfaces': {   # (L, C)
            'bg': (0.225, 0.024), 'panel': (0.19, 0.024), 'sel': (0.295, 0.042), 'visual': (0.365, 0.055),
            'border': (0.41, 0.036), 'code': (0.165, 0.022),
        },
        'text': {'fg': (0.88, 0.028), 'dim': (0.60, 0.045)},
        'accent': (0.79, 0.115),
    },
    'light': {
        'surfaces': {
            'bg': (0.975, 0.008), 'panel': (0.945, 0.012), 'sel': (0.905, 0.022), 'visual': (0.84, 0.035),
            'border': (0.76, 0.030), 'code': (0.925, 0.012),
        },
        'text': {'fg': (0.30, 0.040), 'dim': (0.52, 0.040)},
        'accent': (0.52, 0.150),
    },
}

def oklch_to_srgb(L, C, H):
    """OKLCH -> sRGB (0..1), reduciendo el croma si cae fuera de gama."""
    def conv(c):
        h = math.radians(H); a, b = c * math.cos(h), c * math.sin(h)
        l_ = L + 0.3963377774 * a + 0.2158037573 * b
        m_ = L - 0.1055613458 * a - 0.0638541728 * b
        s_ = L - 0.0894841775 * a - 1.2914855480 * b
        l, m, s = l_ ** 3, m_ ** 3, s_ ** 3
        return (4.0767416621 * l - 3.3077115913 * m + 0.2309699292 * s,
                -1.2684380046 * l + 2.6097574011 * m - 0.3413193965 * s,
                -0.0041960863 * l - 0.7034186147 * m + 1.7076147010 * s)
    c = C
    while c > 0:
        rgb = conv(c)
        if all(-1e-4 <= v <= 1 + 1e-4 for v in rgb): break
        c -= 0.002
    rgb = conv(max(c, 0))
    return tuple(min(1, max(0, v)) for v in rgb)

def to_hex(rgb_lin):
    def enc(v): return 12.92 * v if v <= 0.0031308 else 1.055 * (v ** (1 / 2.4)) - 0.055
    return '#%02x%02x%02x' % tuple(round(enc(v) * 255) for v in rgb_lin)

def lum(rgb_lin): r, g, b = rgb_lin; return 0.2126 * r + 0.7152 * g + 0.0722 * b
def contrast(a, b):
    la, lb = lum(a), lum(b)
    if la < lb: la, lb = lb, la
    return (la + 0.05) / (lb + 0.05)

out, problems = {}, []
for name, t in THEMES.items():
    lin, hexes = {}, {}
    for k, (L, C) in t['surfaces'].items(): lin[k] = oklch_to_srgb(L, C, BRAND_HUE)
    for k, (L, C) in t['text'].items(): lin[k] = oklch_to_srgb(L, C, BRAND_HUE)
    for k, h in ACCENT_HUES.items(): lin[k] = oklch_to_srgb(*t['accent'][:1], t['accent'][1], h)
    lin['brand'] = oklch_to_srgb(*BRAND[name])
    for k, v in lin.items(): hexes[k] = to_hex(v)
    out[name] = hexes
    bg = lin['bg']
    checks = [('fg', 7.0), ('dim', 3.0), ('brand', 4.5)] + [(k, 4.5) for k in ACCENT_HUES]
    print(f'\n== {name}: contraste sobre bg {hexes["bg"]}')
    for k, minimum in checks:
        r = contrast(lin[k], bg); ok = r >= minimum
        print(f'  {k:7s} {hexes[k]}  {r:5.2f}:1  (min {minimum})  {"ok" if ok else "FALLA"}')
        if not ok: problems.append((name, k, r))
    # texto de la etiqueta de modo: bg sobre cada acento
    for k in ACCENT_HUES:
        r = contrast(bg, lin[k])
        if r < 4.5: problems.append((name, f'etiqueta-{k}', r))

path = os.path.join(os.path.dirname(__file__), '..', 'src', 'palette.js')
with open(path, 'w', encoding='utf-8') as f:
    f.write('// GENERADO por tools/make-palette.py: no editar a mano.\n')
    f.write('// Paleta propia derivada en OKLCH (ver docs/design/README.md).\n')
    f.write('const THEMES = ' + json.dumps(out, indent=2) + ';\n')
print('\nescrito', os.path.normpath(path))
if problems:
    print('PROBLEMAS DE CONTRASTE:', problems); sys.exit(1)
