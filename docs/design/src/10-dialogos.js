// 10 - Dialogo de cambios sin guardar y mensaje de error en la linea de comandos
// Orden de capas: en OpenPencil el PRIMER hijo queda encima -> popup, atenuado, app.
const T = THEMES.dark;
const r = screen('10 Dialogo cambios sin guardar', 1, 3, T);
const p = popup(r, T, 29, 12, 62, 8.06, 'dialogo');
dim(r, T);
app(r, T, { mode: 'NORMAL', left: 'internal/keymap/trie.go [+]', right: 'go  utf-8  LF   12:35  54%', treeSel: 'trie.go', editor: { curLine: 12, lines: code(T) },
  cmd: [['E37: cambios sin guardar (usa :q! para descartarlos)', T.red, { fontWeight: 700 }]] });
line(p, [[' Cambios sin guardar', T.peach, { fontWeight: 700 }]], T.sel);
line(p, [[' ', T.fg]]);
line(p, [['  "internal/keymap/trie.go" tiene cambios sin guardar.', T.fg]]);
line(p, [['  ¿Qué quieres hacer?', T.dim]]);
line(p, [[' ', T.fg]]);
const btns = I(p, { type: 'frame', layout: 'horizontal', gap: 0, width: 'fill_container', height: CH, alignItems: 'center', fill: [] });
span(btns, ['  ', T.fg]); span(btns, [' Guardar (y) ', T.bg, { bg: T.blue, fontWeight: 700 }]); span(btns, ['  ', T.fg]);
span(btns, [' Descartar (n) ', T.fg, { bg: T.sel }]); span(btns, ['  ', T.fg]); span(btns, [' Cancelar (Esc) ', T.fg, { bg: T.sel }]);
line(p, [[' ', T.fg]]);
hsep(p, T);
line(p, [[' ←→ ', T.mauve], ['elegir  ', T.dim], ['⏎ ', T.mauve], ['confirmar', T.dim]], T.sel);
