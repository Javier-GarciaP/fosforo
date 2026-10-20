// 12 - Marca: candidatos de nombre e icono (boceto de direccion; se sustituye al elegir)
// Fuera de la cuadricula de terminal: es un tablero de decision, no una pantalla del editor.
const T = THEMES.dark;
const BW = 1360, BH = 640;
// Este motor no ajusta el ancho de los parrafos: se cortan las lineas a mano
function wrap(str, n) {
  const out = []; let cur = '';
  str.split(' ').forEach((w) => { if ((cur + ' ' + w).trim().length > n) { out.push(cur); cur = w; } else { cur = (cur + ' ' + w).trim(); } });
  if (cur) out.push(cur);
  return out;
}
function para(parent, str, color, size, n) {
  wrap(str, n).forEach((l) => {
    const r = I(parent, { type: 'frame', layout: 'horizontal', justifyContent: 'center', alignItems: 'center', width: 'fill_container', height: Math.round(size * 1.5), fill: [] });
    txt(r, l, color, { fontSize: size });
  });
}
const board = I(null, { type: 'frame', name: '12 Marca - candidatos', x: 0, y: 4 * (H + GAP), width: BW, height: BH, layout: 'vertical', gap: 0,
  padding: [28, 32, 28, 32], fill: solid(T.bg) });
line(board, [['NOMBRE E ICONO', T.fg, { fontWeight: 700 }], ['   cinco direcciones. Libres en GitHub, crates.io, npm, PyPI y Homebrew (comprobado).', T.dim]]);
line(board, [[' ', T.fg]]);
const row = I(board, { type: 'frame', layout: 'horizontal', gap: 16, width: 'fill_container', height: 'fill_container', fill: [] });

const grad = (a, b, angle) => [{ type: 'linear_gradient', angle: angle === undefined ? 180 : angle, stops: [{ offset: 0, color: a }, { offset: 1, color: b }] }];
const rad = (c0, c1) => [{ type: 'radial_gradient', cx: 0.5, cy: 0.5, radius: 0.7, stops: [{ offset: 0, color: c0 }, { offset: 1, color: c1 }] }];

// Dibuja la marca en un cuadrado de `size` px (el diseno se piensa en 120 unidades)
function mark(parent, kind, size) {
  const k = size / 120;
  const tile = I(parent, { type: 'frame', name: 'icono ' + kind, width: size, height: size, layout: 'none', cornerRadius: 28 * k,
    fill: grad('#222a47', '#0d111d', 135), stroke: { thickness: 1, fill: solid('#3a4466') }, clipContent: true });
  const R = (x, y, w, h, r, fill, extra) => I(tile, Object.assign({ type: 'rectangle', x: x * k, y: y * k, width: w * k, height: h * k, cornerRadius: r * k, fill }, extra || {}));
  const E = (x, y, w, h, fill) => I(tile, { type: 'ellipse', x: x * k, y: y * k, width: w * k, height: h * k, fill });
  const P = (x, y, w, h, d, fill, stroke) => I(tile, Object.assign({ type: 'path', x: x * k, y: y * k, width: w * k, height: h * k, d }, fill ? { fill } : { fill: [] }, stroke ? { stroke } : {}));
  if (kind === 'fosforo') {
    E(24, 4, 72, 72, rad('#ff8a5055', '#ff8a5000'));            // brillo
    R(53, 44, 14, 62, 7, solid('#e9e4d4'));                      // palo = cursor
    E(44, 14, 32, 38, grad('#ffc47a', '#ff4d33'));               // cabeza de la cerilla
  } else if (kind === 'lumbre') {
    R(34, 54, 52, 50, 11, grad('#ffa45a', '#d63f2c'));            // brasa = cursor de bloque
    P(47, 8, 26, 46, 'M13,46 C3,46 0,33 6,24 C10,18 13,10 13,0 C22,9 26,18 26,30 C26,39 21,46 13,46 Z', grad('#ffe08a', '#ff8a3c'));
    E(80, 30, 8, 8, solid('#ffcf7a'));
  } else if (kind === 'atajo') {
    P(22, 36, 34, 48, 'M2,2 L28,24 L2,46', null, { thickness: 11, fill: solid('#7fc1fe') });   // »
    P(50, 36, 34, 48, 'M2,2 L28,24 L2,46', null, { thickness: 11, fill: solid('#4ad1d3') });
    R(90, 36, 14, 48, 4, solid('#7fd096'));                       // cursor
  } else if (kind === 'buril') {
    P(30, 14, 64, 64, 'M40,0 L64,24 L26,62 L2,38 Z', solid('#a9b6d8'));       // mango en diagonal
    P(26, 52, 34, 40, 'M10,10 L34,34 L2,38 Z', solid('#4ad1d3'));              // punta
    R(26, 98, 68, 5, 2.5, solid('#4a5578'));                                  // surco grabado
  } else if (kind === 'rafaga') {
    R(20, 34, 48, 11, 5.5, solid('#cca8f6'), { opacity: 0.4 });
    R(20, 54, 38, 11, 5.5, solid('#cca8f6'), { opacity: 0.7 });
    R(20, 74, 28, 11, 5.5, solid('#cca8f6'));
    R(78, 28, 22, 64, 5, solid('#d5b85e'));                       // cursor
  }
  return tile;
}

const CANDS = [
  { id: 'fosforo', name: 'fosforo', acc: T.green, idea: 'Cerilla + fósforo de los monitores CRT. El cursor que enciende ideas.',
    ok: ['GitHub', 'crates', 'npm', 'PyPI', 'brew'], note: '7 letras; alias corto posible', rec: 1 },
  { id: 'lumbre', name: 'lumbre', acc: T.peach, idea: 'Brasa y luz. Un bloque de cursor al rojo vivo.',
    ok: ['GitHub', 'crates', 'npm', 'PyPI', 'brew'], note: '6 letras, cálido, fácil de pronunciar' },
  { id: 'atajo', name: 'atajo', acc: T.blue, idea: 'Atajo = shortcut: la promesa del teclado primero.',
    ok: ['GitHub', 'crates', 'npm', 'PyPI', 'brew'], note: 'Ojo: "Atajos" es la app Shortcuts de Apple en español', warn: 1 },
  { id: 'buril', name: 'buril', acc: T.teal, idea: 'Herramienta del grabador: precisión sobre el texto.',
    ok: ['GitHub', 'crates', 'npm', 'PyPI', 'brew'], note: '5 letras, poco común y muy propio' },
  { id: 'rafaga', name: 'rafaga', acc: T.mauve, idea: 'Ráfaga de pulsaciones: velocidad y fluidez.',
    ok: ['GitHub', 'crates', 'npm', 'PyPI', 'brew'], note: '6 letras; sin tildes en el comando' },
];
CANDS.forEach((c) => {
  const col = I(row, { type: 'frame', name: 'candidato ' + c.name, layout: 'vertical', gap: 8, width: 'fill_container', height: 'fill_container', padding: [20, 12, 16, 12],
    alignItems: 'center', cornerRadius: 12, fill: solid(T.panel), stroke: { thickness: 1, fill: solid(c.rec ? c.acc : T.border) } });
  mark(col, c.id, 132);
  txt(col, c.name, T.fg, { fontSize: 30, fontWeight: 700 });
  const tags = I(col, { type: 'frame', layout: 'horizontal', gap: 8, alignItems: 'center', fill: [] });
  mark(tags, c.id, 32);
  txt(tags, '$ ' + c.name + ' README.md', c.acc, { fontSize: 13 });
  para(col, c.idea, T.fg, 13, 27);
  para(col, c.note, c.warn ? T.peach : T.dim, 12, 29);
  para(col, '✓ libre en GitHub, crates, npm, PyPI y brew', T.green, 11, 32);
});
