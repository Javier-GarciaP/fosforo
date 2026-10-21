// 01 - Layout principal, modo NORMAL, tema oscuro
const T = THEMES.dark;
const r = screen('01 Layout principal (Normal)', 1, 0, T);
const lines = code(T);
lines[11] = [['        next, ok := cur.children', T.fg], ['[', T.bg, { bg: T.fg }], ['r]', T.fg]];
app(r, T, { mode: 'NORMAL', left: 'internal/keymap/trie.go [+]', right: 'go  utf-8  LF   12:35  54%', treeSel: 'trie.go',
  editor: { curLine: 12, lines },
  cmd: [[' <Space> ', T.mauve, { fontWeight: 700 }], ['f archivos  e explorador  m markdown  b buffers  s buscar', T.dim]] });
