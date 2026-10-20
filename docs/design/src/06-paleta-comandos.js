// 06 - Paleta de comandos (Space Space o :) con atajos a la derecha
// Orden de capas: en OpenPencil el PRIMER hijo queda encima -> popup, atenuado, app.
const T = THEMES.dark;
const r = screen('06 Paleta de comandos', 0, 2, T);
const p = popup(r, T, 25, 3, 70, 7.12, 'paleta');
dim(r, T);
app(r, T, { mode: 'COMMAND', left: 'internal/keymap/trie.go [+]', right: 'go  utf-8  LF   12:35  54%', treeSel: 8, editor: { curLine: 12, lines: code(T) },
  cmd: [[':gua', T.fg], ['▏', T.fg]] });
line(p, [[' : ', T.peach, { fontWeight: 700 }], ['gua', T.fg], ['▏', T.fg]]);
hsep(p, T);
[['Guardar archivo', ':w  Space w'], ['Guardar como...', ':saveas'], ['Guardar todo', ':wa'], ['Guardar y salir', ':wq'], ['Alternar guardado automatico', ':set autosave']].forEach((c, i) => {
  rowLR(p, T, [[i === 0 ? ' ▸ ' : '   ', T.peach]].concat(hl(c[0], 'Gua', i === 0 ? T.fg : T.dim, T.yellow).map((s, j) => s)), c[1], i === 0 ? T.sel : null);
});
hsep(p, T);
line(p, [[' ↑↓ ', T.mauve], ['navegar  ', T.dim], ['⏎ ', T.mauve], ['ejecutar  ', T.dim], ['Esc ', T.mauve], ['cerrar', T.dim]], T.sel);
