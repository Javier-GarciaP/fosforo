#!/usr/bin/env python3
"""Genera el logo de fosforo desde UNA sola definicion de geometria:
  - docs/design/brand/fosforo-logo.svg   (vectorial, unicolor, usa currentColor)
  - docs/design/src/logo.js              (datos para dibujarlo en OpenPencil)

Concepto: < > (simbolo universal de codigo) con una llama sutil dentro: codigo que enciende ideas.
Todo en un cuadrado de 256 unidades. Unicolor y sin fondo.

Uso:  python3 docs/design/tools/make-logo.py
"""
import json, os

STROKE = 24                       # grosor de los chevrones
CHEVRONS = [                      # (x, y, w, h, puntos locales)
    (24, 84, 48, 88, [(48, 0), (0, 44), (48, 88)]),     # <
    (184, 84, 48, 88, [(0, 0), (48, 44), (0, 88)]),     # >
]
# Llama: contorno exterior (sentido horario) + gota interior (sentido antihorario, para que haga hueco)
# Cada subtrazado: punto inicial + lista de curvas cubicas (c1, c2, destino)
FLAME_OUT = ((128, 52), [((136, 82), (162, 102), (162, 142)), ((162, 172), (147, 196), (128, 196)), ((109, 196), (94, 174), (94, 146)),
                         ((94, 124), (105, 111), (114, 98)), ((116, 112), (121, 121), (127, 125)), ((121, 101), (121, 76), (128, 52))])
FLAME_IN = ((128, 140), [((121, 151), (114, 159), (114, 170)), ((114, 178), (120, 184), (128, 184)),
                         ((136, 184), (142, 178), (142, 170)), ((142, 159), (135, 151), (128, 140))])
FLAME_SCALE, FLAME_CENTER = 0.82, (128, 124)     # la llama es sutil: cabe en la altura de los chevrones

def tf(p):
    s, (cx, cy) = FLAME_SCALE, FLAME_CENTER
    return (cx + (p[0] - cx) * s + 0.0, 126 + (p[1] - cy) * s)   # centrada un poco por debajo del eje

def fmt(v): return ('%.2f' % v).rstrip('0').rstrip('.')
def subpath(sp, origin=(0, 0)):
    p0, curves = sp
    f = lambda p: '%s,%s' % tuple(fmt(c - o) for c, o in zip(tf(p), origin))
    return 'M' + f(p0) + ' ' + ' '.join('C' + ' '.join(f(q) for q in c) for c in curves) + ' Z'
def samples(sp):
    pts = []; p0, curves = sp; cur = tf(p0); pts.append(cur)
    for c1, c2, p in curves:
        a, b, c, d = cur, tf(c1), tf(c2), tf(p)
        for i in range(1, 41):
            t = i / 40; u = 1 - t
            pts.append((u**3*a[0] + 3*u*u*t*b[0] + 3*u*t*t*c[0] + t**3*d[0], u**3*a[1] + 3*u*u*t*b[1] + 3*u*t*t*c[1] + t**3*d[1]))
        cur = d
    return pts
pts = samples(FLAME_OUT)
bx0, by0 = min(p[0] for p in pts), min(p[1] for p in pts)
bx1, by1 = max(p[0] for p in pts), max(p[1] for p in pts)
flame_abs = subpath(FLAME_OUT) + ' ' + subpath(FLAME_IN)
flame_loc = subpath(FLAME_OUT, (bx0, by0)) + ' ' + subpath(FLAME_IN, (bx0, by0))

root = os.path.join(os.path.dirname(__file__), '..')
svg = f'''<svg xmlns="http://www.w3.org/2000/svg" width="256" height="256" viewBox="0 0 256 256" role="img" aria-label="fosforo">
  <title>fosforo</title>
  <!-- GENERADO por docs/design/tools/make-logo.py: no editar a mano. Unicolor y sin fondo: toma el color de currentColor. -->
  <g fill="none" stroke="currentColor" stroke-width="{STROKE}" stroke-linecap="round" stroke-linejoin="round">
''' + '\n'.join('    <path d="%s"/>' % ' '.join(('M' if i == 0 else 'L') + '%d %d' % (x + px, y + py) for i, (px, py) in enumerate(pts_)) for x, y, w, h, pts_ in CHEVRONS) + f'''
  </g>
  <path d="{flame_abs}" fill="currentColor" fill-rule="evenodd"/>
</svg>
'''
open(os.path.join(root, 'brand', 'fosforo-logo.svg'), 'w', encoding='utf-8').write(svg)
data = {'size': 256, 'stroke': STROKE,
        'chevrons': [{'x': x, 'y': y, 'w': w, 'h': h, 'pts': pts_} for x, y, w, h, pts_ in CHEVRONS],
        'flame': {'x': round(bx0, 2), 'y': round(by0, 2), 'w': round(bx1 - bx0, 2), 'h': round(by1 - by0, 2), 'd': flame_loc}}
with open(os.path.join(root, 'src', 'logo.js'), 'w', encoding='utf-8') as f:
    f.write('// GENERADO por tools/make-logo.py: no editar a mano. Geometria del logo en un cuadrado de 256 unidades.\n')
    f.write('const LOGO = ' + json.dumps(data, separators=(',', ':')) + ';\n')
print('flama: bbox x=%.1f y=%.1f w=%.1f h=%.1f' % (bx0, by0, bx1 - bx0, by1 - by0), '| chevrones 24..232, y 72..184 (con trazo)')
