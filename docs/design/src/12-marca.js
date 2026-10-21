// 12 - Marca fosforo: icono, versiones, tamanos, lockups y colores.
// Concepto: un cursor de bloque verde fosforo (monitores CRT) coronado por la cabeza de una cerilla (enciende ideas).
// Geometria en un cuadrado de 512 unidades; la misma del SVG docs/design/brand/fosforo-icon.svg (mantenerlas iguales).
const T = THEMES.dark, L = THEMES.light;
const BW = 1360, BH = 720;
const board = I(null, { type: 'frame', name: '12 Marca fosforo', x: 0, y: 4 * (H + GAP), width: BW, height: BH, layout: 'vertical', gap: 0,
  padding: [28, 32, 28, 32], fill: solid(T.bg) });

const grad = (stops, angle) => [{ type: 'linear_gradient', angle: angle === undefined ? 180 : angle, stops: stops.map((c, i) => ({ offset: i / (stops.length - 1), color: c })) }];
const rad = (c0, c1) => [{ type: 'radial_gradient', cx: 0.5, cy: 0.5, radius: 0.5, stops: [{ offset: 0, color: c0 }, { offset: 1, color: c1 }] }];

// variant: 'dark' | 'light' | 'mono'
function mark(parent, size, variant, name) {
  const k = size / 512;
  const light = variant === 'light', mono = variant === 'mono';
  const tile = I(parent, { type: 'frame', name: name || ('icono ' + variant + ' ' + size), width: size, height: size, layout: 'none', cornerRadius: 128 * k,
    fill: light ? grad(['#fbf8ef', '#e8e2d0'], 135) : grad(['#222a47', '#0d111d'], 135), clipContent: true });
  const R = (x, y, w, h, r, fill) => I(tile, { type: 'rectangle', x: x * k, y: y * k, width: w * k, height: h * k, cornerRadius: r * k, fill });
  const E = (x, y, w, h, fill) => I(tile, { type: 'ellipse', x: x * k, y: y * k, width: w * k, height: h * k, fill });
  // OJO: en OpenPencil el PRIMER hijo queda ENCIMA, asi que se crea de arriba (reflejo) hacia abajo (brillos).
  if (mono) {
    E(190, 88, 132, 160, solid('#e8f0ff'));
    R(222, 264, 68, 176, 34, solid('#e8f0ff'));   // hueco entre cabeza y palo para que no se fundan
    return tile;
  }
  E(222, 112, 28, 44, solid('#ffffff4d'));                                         // reflejo
  E(190, 88, 132, 160, grad(['#ffd98f', '#ff8f4e', '#ff4a33']));                  // cabeza de cerilla
  R(222, 200, 68, 240, 34, grad(light ? ['#2fbf6c', '#067e42'] : ['#aaffc9', '#3fd47f']));  // cursor de bloque
  E(106, 200, 300, 300, rad(light ? '#1fa95b30' : '#69f19a40', '#69f19a00'));     // brillo fosforo (circulo: llega a 0 en el borde)
  E(120, 20, 272, 272, rad(light ? '#ff8a5038' : '#ff8a5070', '#ff8a5000'));      // brillo de la cerilla
  return tile;
}

line(board, [['MARCA', T.fg, { fontWeight: 700 }], ['   fosforo: un cursor verde que enciende ideas', T.dim]]);
line(board, [[' ', T.fg]]);

// --- fila A: icono grande + lockups ---
const A = I(board, { type: 'frame', layout: 'horizontal', gap: 24, width: 'fill_container', height: 276, fill: [] });
mark(A, 256, 'dark', 'logo icono oscuro');
const lockups = I(A, { type: 'frame', layout: 'vertical', gap: 16, width: 'fill_container', height: 'fill_container', fill: [] });
function lockup(name, T2, variant, bg, fg, sub) {
  const p = I(lockups, { type: 'frame', name: name, layout: 'horizontal', gap: 24, width: 'fill_container', height: 130, padding: [0, 32, 0, 32], alignItems: 'center',
    cornerRadius: 14, fill: solid(bg), stroke: { thickness: 1, fill: solid(T.border) } });
  mark(p, 88, variant, name + ' icono');
  const t = I(p, { type: 'frame', layout: 'vertical', gap: 2, fill: [] });
  txt(t, 'fosforo', fg, { fontSize: 56, fontWeight: 700 });
  txt(t, 'editor de código en la terminal', sub, { fontSize: 16 });
  return p;
}
lockup('logo lockup oscuro', T, 'dark', T.panel, T.fg, T.dim);
lockup('logo lockup claro', L, 'light', L.bg, L.fg, L.dim);

line(board, [[' ', T.fg]]);
line(board, [['TAMAÑOS Y VERSIONES', T.fg, { fontWeight: 700 }], ['   legible desde 16 px; la silueta (cursor + cabeza) se mantiene', T.dim]]);
line(board, [[' ', T.fg]]);

// --- fila B: tamanos y versiones ---
const B = I(board, { type: 'frame', layout: 'horizontal', gap: 28, width: 'fill_container', height: 140, alignItems: 'end', fill: [] });
[128, 64, 48, 32, 24, 16].forEach((s) => {
  const c = I(B, { type: 'frame', layout: 'vertical', gap: 8, alignItems: 'center', fill: [] });
  mark(c, s, 'dark');
  txt(c, s + ' px', T.dim, { fontSize: 12 });
});
I(B, { type: 'frame', width: 24, height: 1, fill: [] });
[['light', 'claro'], ['mono', 'monocromo']].forEach((v) => {
  const c = I(B, { type: 'frame', layout: 'vertical', gap: 8, alignItems: 'center', fill: [] });
  mark(c, 96, v[0]);
  txt(c, v[1], T.dim, { fontSize: 12 });
});

line(board, [[' ', T.fg]]);
line(board, [['COLORES DE MARCA', T.fg, { fontWeight: 700 }]]);
line(board, [[' ', T.fg]]);
const C = I(board, { type: 'frame', layout: 'horizontal', gap: 20, width: 'fill_container', height: 78, fill: [] });
[['verde fosforo', T.brand, 'el cursor; acento de marca (claro: ' + L.brand + ')'], ['ember', '#ff8f4e', 'la cabeza de la cerilla (peach de la paleta: ' + T.peach + ')'], ['tinta', '#0d111d', 'fondo del icono'], ['papel', '#fbf8ef', 'fondo del icono claro']].forEach((c) => {
  const w = I(C, { type: 'frame', layout: 'horizontal', gap: 12, alignItems: 'center', fill: [] });
  I(w, { type: 'rectangle', width: 60, height: 60, cornerRadius: 10, fill: solid(c[1]), stroke: { thickness: 1, fill: solid(T.border) } });
  const t = I(w, { type: 'frame', layout: 'vertical', gap: 2, fill: [] });
  txt(t, c[0], T.fg, { fontSize: 14, fontWeight: 700 });
  txt(t, c[1], T.dim, { fontSize: 12 });
});
