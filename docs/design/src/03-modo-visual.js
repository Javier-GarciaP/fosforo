// 03 - Modo VISUAL: seleccion de lineas 10-12
const T = THEMES.dark;
const r = screen('03 Modo Visual', 0, 1, T);
const bgAll = (spans) => spans.map((s) => [s[0], s[1], Object.assign({}, s[2] || {}, { bg: T.visual })]);
const lines = code(T);
for (const i of [9, 10, 11]) lines[i] = bgAll(lines[i]);
app(r, T, { mode: 'VISUAL', left: 'internal/keymap/trie.go [+]', right: '3 lineas   12:35  54%', treeSel: 8,
  editor: { curLine: 12, lines },
  cmd: [['-- VISUAL --', T.mauve, { fontWeight: 700 }]] });
