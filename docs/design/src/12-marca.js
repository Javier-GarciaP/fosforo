// 12 - Marca fosforo: logo unicolor sin fondo.
// Concepto: < > (simbolo universal de codigo) con una llama sutil dentro: codigo que enciende ideas.
// La geometria viene de src/logo.js (GENERADO por tools/make-logo.py, que tambien genera brand/fosforo-logo.svg).
const T = THEMES.dark, L = THEMES.light;
const BW = 1200, BH = 600;
const board = I(null, { type: 'frame', name: '12 Marca fosforo', x: 0, y: 4 * (H + GAP), width: BW, height: BH, layout: 'vertical', gap: 0,
  padding: [28, 32, 28, 32], fill: solid(T.bg) });

// Escala los numeros de un trazado (las letras de comando se conservan)
function scaleD(d, k) { return d.replace(/-?\d+(\.\d+)?/g, (n) => (parseFloat(n) * k).toFixed(2)); }

// Dibuja el logo en un cuadrado transparente de `size` px, de un solo color.
function logo(parent, size, color, name) {
  const k = size / LOGO.size;
  const box = I(parent, { type: 'frame', name: name || 'logo', width: size, height: size, layout: 'none', fill: [] });
  const stroke = { thickness: LOGO.stroke * k, fill: solid(color) };
  const f = LOGO.flame;
  I(box, { type: 'path', name: 'llama', x: f.x * k, y: f.y * k, width: f.w * k, height: f.h * k, d: scaleD(f.d, k), fill: solid(color) });
  LOGO.chevrons.forEach((c) => I(box, { type: 'path', name: 'chevron', x: c.x * k, y: c.y * k, width: c.w * k, height: c.h * k,
    d: c.pts.map((p, i) => (i ? 'L' : 'M') + (p[0] * k).toFixed(2) + ',' + (p[1] * k).toFixed(2)).join(' '), fill: [], stroke }));
  return box;
}

line(board, [['MARCA', T.fg, { fontWeight: 700 }], ['   fosforo: código que enciende ideas', T.dim]]);
line(board, [[' ', T.fg]]);

// --- fila A: logo grande sobre oscuro y sobre claro ---
const A = I(board, { type: 'frame', layout: 'horizontal', gap: 24, width: 'fill_container', height: 300, fill: [] });
[['sobre oscuro', T.panel, '#ffffff'], ['sobre claro', L.bg, '#000000']].forEach((v) => {
  const p = I(A, { type: 'frame', name: 'logo ' + v[0], layout: 'horizontal', width: 300, height: 300, alignItems: 'center', justifyContent: 'center',
    cornerRadius: 14, fill: solid(v[1]), stroke: { thickness: 1, fill: solid(T.border) } });
  logo(p, 220, v[2], 'logo ' + v[0] + ' 220');
});
const notes = I(A, { type: 'frame', layout: 'vertical', gap: 0, width: 'fill_container', height: 'fill_container', padding: [8, 0, 0, 16], fill: [] });
[['EL LOGO', 1], ['', 0], ['< > es el símbolo universal de código.', 0], ['La llama del centro es el fósforo: lo que enciende.', 0], ['Sutil: cabe en la altura de los chevrones.', 0], ['', 0],
 ['Unicolor y sin fondo: toma el color del texto', 0], ['del contexto (negro, blanco o el que toque).', 0], ['', 0], ['Legible desde 24 px; a 16 px la llama se reduce a un punto.', 0]].forEach((r) => {
  line(notes, [[r[0], r[1] ? T.fg : T.dim, r[1] ? { fontWeight: 700 } : {}]]);
});

line(board, [[' ', T.fg]]);
line(board, [['TAMAÑOS', T.fg, { fontWeight: 700 }]]);
line(board, [[' ', T.fg]]);
const B = I(board, { type: 'frame', layout: 'horizontal', gap: 28, width: 'fill_container', height: 150, alignItems: 'end', fill: [] });
[128, 64, 48, 32, 24, 16].forEach((s) => {
  const c = I(B, { type: 'frame', layout: 'vertical', gap: 8, alignItems: 'center', fill: [] });
  logo(c, s, '#ffffff');
  txt(c, s + ' px', T.dim, { fontSize: 12 });
});
