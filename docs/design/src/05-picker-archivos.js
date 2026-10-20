// 05 - Picker fuzzy de archivos (Space f f) con vista previa
// Orden de capas: en OpenPencil el PRIMER hijo queda encima -> popup, atenuado, app.
const T = THEMES.dark;
const r = screen('05 Picker de archivos', 2, 1, T);
const p = popup(r, T, 10, 4, 100, 26, 'picker');
dim(r, T);
app(r, T, { mode: 'NORMAL', left: 'internal/keymap/trie.go [+]', right: 'go  utf-8  LF   12:35  54%', treeSel: 8, editor: { curLine: 12, lines: code(T) } });
const head = I(p, { type: 'frame', layout: 'horizontal', gap: 0, width: 'fill_container', height: CH, alignItems: 'center', justifyContent: 'space_between', fill: solid(T.sel) });
txt(head, ' Buscar archivos', T.blue, { fontWeight: 700 }); txt(head, '5/42 ', T.dim);
line(p, [[' > ', T.mauve, { fontWeight: 700 }], ['trie', T.fg], ['▏', T.fg]]);
hsep(p, T);
const body = I(p, { type: 'frame', layout: 'horizontal', gap: 0, width: 'fill_container', height: 23 * CH - 1, fill: [] });
const list = I(body, { type: 'frame', layout: 'vertical', gap: 0, width: 40 * CW, height: 'fill_container', fill: [] });
['internal/keymap/trie.go', 'internal/keymap/trie_test.go', 'changes/keymap-trie/design.md', 'changes/keymap-trie/tasks.md', 'changes/keymap-trie/proposal.md'].forEach((f, i) => {
  line(list, [[i === 0 ? ' ▸ ' : '   ', T.mauve]].concat(hl(f, 'trie', i === 0 ? T.fg : T.dim, T.yellow)), i === 0 ? T.sel : null);
});
vsep(body, T);
const prev = I(body, { type: 'frame', layout: 'vertical', gap: 0, width: 'fill_container', height: 'fill_container', fill: [] });
code(T).forEach((l, i) => line(prev, [[String(i + 1).padStart(3, ' ') + ' ', T.dim]].concat(l.length ? l : [[' ', T.fg]])));
hsep(p, T);
line(p, [[' ↑↓ ', T.mauve], ['navegar  ', T.dim], ['⏎ ', T.mauve], ['abrir  ', T.dim], ['Ctrl+n/p ', T.mauve], ['mover  ', T.dim], ['Esc ', T.mauve], ['cerrar', T.dim]], T.sel);
