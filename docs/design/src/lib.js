// Libreria compartida de las maquetas. Se concatena asi (palette.js e icons.js son generados, ver tools/):
//   cat src/palette.js src/icons.js src/lib.js src/NN-pantalla.js > /tmp/NN.js && op design @/tmp/NN.js
// Cuadricula de terminal: 120x36 celdas, celda = 9x18 px (JetBrains Mono 15px => avance de 9px).
const FONT = 'JetBrains Mono', FS = 15, CW = 9, CH = 18, COLS = 120, ROWS = 36, W = COLS * CW, H = ROWS * CH, GAP = 120;
const solid = (c) => [{ type: 'solid', color: c }];
const MODE_COLOR = { NORMAL: 'blue', INSERT: 'green', VISUAL: 'mauve', COMMAND: 'peach', PREVIEW: 'teal' };

function txt(p, s, color, o) {
  return I(p, Object.assign({ type: 'text', content: s, fontFamily: FONT, fontSize: FS, fontWeight: 400,
    fill: solid(color), textGrowth: 'auto', lineHeight: 1.2 }, o || {}));
}
// span = [texto, color, {bg, fontWeight, ...}]
function span(p, s) {
  const o = Object.assign({}, s[2] || {}); const bg = o.bg; delete o.bg;
  if (bg) {
    const f = I(p, { type: 'frame', layout: 'horizontal', gap: 0, height: CH, alignItems: 'center', fill: solid(bg) });
    return txt(f, s[0], s[1], o);
  }
  return txt(p, s[0], s[1], o);
}
function line(p, spans, bg) {
  const r = I(p, { type: 'frame', name: 'linea', layout: 'horizontal', gap: 0, height: CH, width: 'fill_container',
    alignItems: 'center', fill: bg ? solid(bg) : [] });
  spans.forEach((s) => span(r, s));
  return r;
}
function screen(name, col, row, T) {
  return I(null, { type: 'frame', name, x: col * (W + GAP), y: row * (H + GAP), width: W, height: H, layout: 'none',
    fill: solid(T.bg), clipContent: true });
}
function appFrame(root, T) {
  return I(root, { type: 'frame', name: 'app', x: 0, y: 0, width: W, height: H, layout: 'vertical', gap: 0, fill: solid(T.bg) });
}
function dim(root, T) {
  return I(root, { type: 'rectangle', name: 'atenuado', x: 0, y: 0, width: W, height: H, fill: solid(T.bg), opacity: 0.6 });
}
// Ventana flotante con borde (x, y, w, h en CELDAS)
function popup(root, T, x, y, w, h, name) {
  return I(root, { type: 'frame', name: name || 'popup', x: x * CW, y: y * CH, width: w * CW, height: h * CH, layout: 'vertical', gap: 0,
    fill: solid(T.panel), stroke: { thickness: 1, fill: solid(T.border) }, clipContent: true,
    effects: [{ type: 'shadow', offsetX: 0, offsetY: 6, blur: 18, spread: 0, color: '#00000066' }] });
}
function vsep(p, T) {
  return I(p, { type: 'rectangle', name: 'separador', width: 1, height: 'fill_container', fill: solid(T.border) });
}
function hsep(p, T) {
  return I(p, { type: 'rectangle', name: 'separador', width: 'fill_container', height: 1, fill: solid(T.border) });
}

// ---- piezas de la app ----
function statusline(app, T, mode, leftText, rightText) {
  const st = I(app, { type: 'frame', name: 'statusline', layout: 'horizontal', gap: 0, width: 'fill_container', height: CH,
    alignItems: 'center', justifyContent: 'space_between', fill: solid(T.panel) });
  const l = I(st, { type: 'frame', layout: 'horizontal', gap: 0, height: CH, alignItems: 'center', fill: [] });
  span(l, [' ' + mode + ' ', T.bg, { bg: T[MODE_COLOR[mode]], fontWeight: 700 }]);
  txt(l, '  ' + leftText, T.fg);
  const r = I(st, { type: 'frame', layout: 'horizontal', gap: 0, height: CH, alignItems: 'center', fill: [] });
  txt(r, rightText + ' ', T.dim);
  return st;
}
// ---- iconos (glifos de Nerd Fonts convertidos a trazados; ver icons.js) ----
function iconFor(name) {
  if (ICON_BY_NAME[name]) return ICON_BY_NAME[name];
  const dot = name.lastIndexOf('.');
  const ext = dot > 0 ? name.slice(dot + 1).toLowerCase() : '';
  return ICON_BY_EXT[ext] || 'default';
}
// Hueco de 2 celdas con el icono dentro (en el terminal: el glifo + un espacio)
function iconNode(p, T, key, color) {
  const ic = ICONS[key];
  const slot = I(p, { type: 'frame', name: 'icono ' + key, layout: 'horizontal', width: 2 * CW, height: CH, alignItems: 'center', justifyContent: 'start', padding: [0, 0, 0, 2], fill: [] });
  I(slot, { type: 'path', name: key, d: ic.d, width: ic.w, height: ic.h, fill: solid(T[color || ICON_COLOR[key]]) });
  return slot;
}
// Fila del explorador: sangria de 2 celdas por nivel, sin flechas. Carpeta abierta/cerrada = icono distinto.
function treeRow(t, T, depth, kind, name, selected) {
  const row = I(t, { type: 'frame', name: 'fila ' + name, layout: 'horizontal', gap: 0, width: 'fill_container', height: CH, alignItems: 'center',
    padding: [0, 0, 0, CW], fill: selected ? solid(T.sel) : [] });
  if (depth > 0) I(row, { type: 'frame', name: 'sangria', width: depth * 2 * CW, height: CH, fill: [] });
  if (kind === 'root') { iconNode(row, T, 'folder_open', 'fg'); txt(row, name, T.fg, { fontWeight: 700 }); }
  else if (kind === 'dir-open') { iconNode(row, T, 'folder_open'); txt(row, name, T.blue); }
  else if (kind === 'dir-closed') { iconNode(row, T, 'folder'); txt(row, name, T.blue); }
  else { iconNode(row, T, iconFor(name)); txt(row, name, T.fg, selected ? { fontWeight: 700 } : {}); }
  return row;
}
const TREE_ROWS = [
  [0, 'root', 'terminal-IDE'],
  [1, 'dir-open', 'cmd'], [2, 'dir-open', 'ted'], [3, 'file', 'main.go'],
  [1, 'dir-open', 'internal'], [2, 'dir-closed', 'buffer'], [2, 'dir-closed', 'editor'], [2, 'dir-open', 'keymap'],
  [3, 'file', 'trie.go'], [3, 'file', 'trie_test.go'], [2, 'dir-closed', 'ui'],
  [1, 'dir-open', 'docs'], [2, 'dir-closed', 'design'], [2, 'file', 'README.md'],
  [1, 'dir-closed', 'openspec'],
  [1, 'file', '.gitignore'], [1, 'file', 'CONTRIBUTING.md'], [1, 'file', 'go.mod'], [1, 'file', 'LICENSE'], [1, 'file', 'Makefile'],
];
// selected = nombre del archivo seleccionado (p. ej. 'trie.go')
function tree(parent, T, selected, rows) {
  const t = I(parent, { type: 'frame', name: 'explorador', layout: 'vertical', gap: 0, width: 24 * CW, height: 'fill_container', padding: [CH / 2, 0, 0, 0], fill: solid(T.panel) });
  (rows || TREE_ROWS).forEach((r) => treeRow(t, T, r[0], r[1], r[2], r[1] === 'file' && r[2] === selected));
  return t;
}
// Lineas de codigo de ejemplo (trie.go)
function code(T) {
  const k = T.mauve, ty = T.yellow, f = T.fg, d = T.dim, fn = T.blue, s = T.green, b = T.peach;
  return [
    [['package', k], [' keymap', f]],
    [],
    [['type', k], [' Node ', ty], ['struct', k], [' {', f]],
    [['    children ', f], ['map', k], ['[', f], ['rune', ty], [']*Node', f]],
    [['    command  ', f], ['string', ty]],
    [['}', f]],
    [],
    [['// Lookup resuelve una secuencia de teclas.', d]],
    [['func', k], [' (n *Node) ', f], ['Lookup', fn], ['(seq []', f], ['rune', ty], [') (', f], ['string', ty], [', ', f], ['bool', ty], [') {', f]],
    [['    cur := n', f]],
    [['    ', f], ['for', k], [' _, r := ', f], ['range', k], [' seq {', f]],
    [['        next, ok := cur.children[r]', f]],
    [['        ', f], ['if', k], [' !ok { ', f], ['return', k], [' ""', s], [', ', f], ['false', b], [' }', f]],
    [['        cur = next', f]],
    [['    }', f]],
    [['    ', f], ['return', k], [' cur.command, cur.command != ""', f]],
    [['}', f]],
  ];
}
// Panel de edicion: pestaña + lineas numeradas + relleno '~'
function editor(parent, T, o) {
  o = o || {};
  const ed = I(parent, { type: 'frame', name: o.name || 'editor', layout: 'vertical', gap: 0, width: o.width || 'fill_container', height: 'fill_container', fill: [] });
  line(ed, [[' ' + (o.title || 'trie.go') + ' ', T.fg, { fontWeight: 700 }]].concat(o.modified === false ? [] : [[' [+]', T.peach]]), T.panel);
  const lines = o.lines || code(T);
  const total = o.rows || 33;
  for (let i = 0; i < total - 1; i++) {
    if (i < lines.length) {
      const cur = o.curLine === i + 1;
      const n = String(i + 1).padStart(3, ' ') + ' ';
      line(ed, [[n, cur ? T.yellow : T.dim]].concat(lines[i].length ? lines[i] : [[' ', T.fg]]), cur ? T.sel : null);
    } else {
      line(ed, [['  ~', T.dim]]);
    }
  }
  return ed;
}
// Aplicacion completa: explorador | editor, statusline y linea de comandos
function app(root, T, o) {
  const a = appFrame(root, T);
  const body = I(a, { type: 'frame', name: 'cuerpo', layout: 'horizontal', gap: 0, width: 'fill_container', height: 34 * CH, fill: [] });
  if (!o.noTree) { tree(body, T, o.treeSel); vsep(body, T); }
  editor(body, T, o.editor);
  statusline(a, T, o.mode, o.left, o.right);
  line(a, o.cmd || [[' ', T.fg]]);
  return a;
}

// ---- helpers extra (picker, paleta, markdown) ----
function hl(path, q, base, hi) {
  const i = path.toLowerCase().indexOf(q.toLowerCase());
  if (i < 0) return [[path, base]];
  // sin textos vacios: un span '' ocupa ancho en el render
  return [[path.slice(0, i), base], [path.slice(i, i + q.length), hi, { fontWeight: 700 }], [path.slice(i + q.length), base]].filter((x) => x[0] !== '');
}
// fila con texto a la izquierda y atajo a la derecha
function rowLR(p, T, leftSpans, rightText, bg) {
  const r = I(p, { type: 'frame', layout: 'horizontal', gap: 0, width: 'fill_container', height: CH, alignItems: 'center',
    justifyContent: 'space_between', fill: bg ? solid(bg) : [] });
  const l = I(r, { type: 'frame', layout: 'horizontal', gap: 0, height: CH, alignItems: 'center', fill: [] });
  leftSpans.forEach((s) => span(l, s));
  txt(r, rightText + ' ', T.dim);
  return r;
}
// Tabla con bordes dibujados (los glifos de caja se renderizan con otra altura en esta herramienta).
// En el editor real se pintan con caracteres de caja; esto es solo la referencia visual.
function tableGrid(p, T, colW, rows) {
  const wrap = I(p, { type: 'frame', layout: 'horizontal', gap: 0, width: 'fill_container', height: rows.length * CH + 1, fill: [] });
  I(wrap, { type: 'frame', width: 2 * CW, height: CH, fill: [] });
  const tbl = I(wrap, { type: 'frame', layout: 'vertical', gap: 0, fill: [], stroke: { thickness: { top: 1, left: 1 }, fill: solid(T.border) } });
  rows.forEach((row, ri) => {
    const r = I(tbl, { type: 'frame', layout: 'horizontal', gap: 0, height: CH, fill: [] });
    row.forEach((cell, ci) => {
      const c = I(r, { type: 'frame', layout: 'horizontal', gap: 0, width: colW[ci] * CW, height: CH, alignItems: 'center', fill: [],
        stroke: { thickness: { right: 1, bottom: 1 }, fill: solid(T.border) } });
      txt(c, ' ' + cell, ri === 0 ? T.fg : (ci === 0 ? T.mauve : T.fg), ri === 0 ? { fontWeight: 700 } : {});
    });
  });
  return wrap;
}
// Vista previa de Markdown (render estilo GitHub sobre la cuadricula de terminal)
function mdPreview(parent, T, o) {
  o = o || {};
  const pane = I(parent, { type: 'frame', name: 'preview markdown', layout: 'vertical', gap: 0, width: 'fill_container', height: 'fill_container', fill: [] });
  line(pane, [[' README.md ', T.fg, { fontWeight: 700 }], [' preview', T.teal]], T.panel);
  const P = '  ', rule = '─'.repeat(o.rule || 54);
  const bl = () => line(pane, [[' ', T.fg]]);
  line(pane, [[P + 'ted — editor de código en la terminal', T.fg, { fontWeight: 700 }]]);
  line(pane, [[P + rule, T.border]]); bl();
  line(pane, [[P + 'Un editor ligero con la fluidez de ', T.fg], ['Neovim', T.blue], [' y la', T.fg]]);
  line(pane, [[P + 'comodidad de VS Code, hecho en Go.', T.fg]]); bl();
  line(pane, [[P + 'Características', T.fg, { fontWeight: 700 }]]);
  line(pane, [[P + rule, T.border]]); bl();
  line(pane, [[P + '• ', T.peach], ['Teclado modal con tecla líder ', T.fg], ['Space', T.mauve, { bg: T.sel }]]);
  line(pane, [[P + '• ', T.peach], ['Explorador de archivos y paleta de comandos', T.fg]]);
  line(pane, [[P + '☑ ', T.green], ['Vista previa de Markdown', T.fg]]);
  line(pane, [[P + '☐ ', T.dim], ['Búsqueda en proyecto', T.fg]]); bl();
  line(pane, [[P + 'Instalación', T.fg, { fontWeight: 700 }]]);
  line(pane, [[P + rule, T.border]]); bl();
  line(pane, [[P + ' ', T.fg, { bg: T.code }], ['go install github.com/usuario/ted@latest' + ' '.repeat(12), T.green, { bg: T.code }]]);
  line(pane, [[P + ' ', T.fg, { bg: T.code }], ['ted README.md' + ' '.repeat(39), T.green, { bg: T.code }]]); bl();
  line(pane, [[P, T.fg], ['▎ ', T.blue], ['Nota: los .md se abren en vista previa.', T.dim]]); bl();
  tableGrid(pane, T, [13, 27], [['Tecla', 'Acción'], ['Space f f', 'Buscar archivo'], ['Space m p', 'Alternar vista previa']]); bl();
  line(pane, [[P + 'Más información en ', T.fg], ['docs/README.md', T.blue]]);
  return pane;
}
