// 09 - Layout principal en tema claro
const T = THEMES.light;
const r = screen('09 Tema claro (Normal)', 0, 3, T);
const lines = code(T);
lines[11] = [['        next, ok := cur.children', T.fg], ['[', T.bg, { bg: T.fg }], ['r]', T.fg]];
app(r, T, { mode: 'NORMAL', left: 'internal/keymap/trie.go [+]', right: 'go  utf-8  LF   12:35  54%', treeSel: 8,
  editor: { curLine: 12, lines },
  cmd: [[' <Space> ', T.mauve, { fontWeight: 700 }], ['f archivos  e explorador  m markdown  b buffers  s buscar', T.dim]] });
