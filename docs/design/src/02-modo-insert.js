// 02 - Modo INSERT: cursor de barra, statusline verde, "-- INSERT --"
const T = THEMES.dark;
const r = screen('02 Modo Insert', 2, 0, T);
const lines = code(T);
lines[11] = [['        next, ok := cur.children[r]', T.fg], ['▏', T.fg]];
lines.splice(12, 0, [['        ', T.fg], ['// comprobar que existe el nodo', T.dim], ['▏', T.fg, { bg: null }]].slice(0, 2));
app(r, T, { mode: 'INSERT', left: 'internal/keymap/trie.go [+]', right: 'go  utf-8  LF   13:40  54%', treeSel: 'trie.go',
  editor: { curLine: 13, lines },
  cmd: [['-- INSERT --', T.green, { fontWeight: 700 }]] });
