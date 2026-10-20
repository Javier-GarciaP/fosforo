// 04 - Menu lider <Space> (which-key) en el borde inferior
// Orden de capas: en OpenPencil el PRIMER hijo queda encima -> popup, atenuado, app.
const T = THEMES.dark;
const r = screen('04 Menu lider (Space)', 1, 1, T);
const lines = code(T);
const p = popup(r, T, 0, 26.94, 120, 7.06, 'which-key');
app(r, T, { mode: 'NORMAL', left: 'internal/keymap/trie.go [+]', right: 'go  utf-8  LF   12:35  54%', treeSel: 'trie.go',
  editor: { curLine: 12, lines }, cmd: [[' <Space> ', T.mauve, { fontWeight: 700 }], ['esperando tecla...', T.dim]] });
line(p, [[' <Space> ', T.mauve, { fontWeight: 700 }], ['lider', T.dim]]);
hsep(p, T);
const cols = I(p, { type: 'frame', layout: 'horizontal', gap: 0, width: 'fill_container', height: 'fill_container', fill: [] });
function col(items) {
  const c = I(cols, { type: 'frame', layout: 'vertical', gap: 0, width: 40 * CW, height: 'fill_container', fill: [] });
  line(c, [[' ', T.fg]]);
  items.forEach((it) => line(c, [['  ' + it[0].padEnd(3), T.mauve, { fontWeight: 700 }], ['→ ', T.dim], [it[1], it[2] ? T.blue : T.fg]]));
}
col([['f', '+archivos', 1], ['e', 'explorador (alternar)'], ['b', '+buffers', 1], ['s', '+buscar', 1], ['h', 'ayuda']]);
col([['m', '+markdown', 1], ['w', 'guardar'], ['q', 'cerrar buffer'], ['Q', 'salir'], ['t', 'tema claro / oscuro']]);
col([['Space', 'paleta de comandos'], ['/', 'buscar en el archivo'], [':', 'linea de comandos'], ['?', 'lista de atajos']]);
