// 11 - Iconos por tipo de archivo y modo sin Nerd Font
const T = THEMES.dark;
const r = screen('11 Iconos y explorador', 2, 3, T);
const base = I(r, { type: 'frame', name: 'iconos', x: 0, y: 0, width: W, height: H, layout: 'vertical', gap: 0, padding: [18, 18, 18, 18], fill: [] });
line(base, [['ICONOS POR TIPO DE ARCHIVO', T.fg, { fontWeight: 700 }], ['   requiere una Nerd Font (ver docs/README.md)', T.dim]]);
line(base, [[' ', T.fg]]);
// patrones por icono (inversa de las tablas generadas)
const pat = {};
Object.keys(ICON_BY_NAME).forEach((n) => { (pat[ICON_BY_NAME[n]] = pat[ICON_BY_NAME[n]] || []).push(n); });
Object.keys(ICON_BY_EXT).forEach((e) => { (pat[ICON_BY_EXT[e]] = pat[ICON_BY_EXT[e]] || []).push('.' + e); });
const labels = { folder: 'carpeta cerrada', folder_open: 'carpeta abierta', default: 'otro archivo' };
const keys = Object.keys(ICONS);
const cols = I(base, { type: 'frame', layout: 'horizontal', gap: 0, width: 'fill_container', height: 13 * CH, fill: [] });
for (let c = 0; c < 3; c++) {
  const col = I(cols, { type: 'frame', layout: 'vertical', gap: 0, width: 'fill_container', height: 'fill_container', fill: [] });
  keys.slice(c * 13, c * 13 + 13).forEach((k) => {
    const row = I(col, { type: 'frame', layout: 'horizontal', gap: 0, height: CH, width: 'fill_container', alignItems: 'center', fill: [] });
    iconNode(row, T, k);
    const label = labels[k] || k;
    txt(row, label.padEnd(9).slice(0, 15).padEnd(labels[k] ? 16 : 9) + ' ', T.fg);
    let p = (pat[k] || []).join(' ');
    if (p.length > 22) p = p.slice(0, 21) + '…';
    txt(row, p, T.dim);
  });
}
line(base, [[' ', T.fg]]);
line(base, [['SIN NERD FONT', T.fg, { fontWeight: 700 }], ['   icons = "none"', T.dim]]);
line(base, [['Las carpetas llevan "/" al final: abiertas en azul con sus hijos sangrados, cerradas atenuadas.', T.dim]]);
line(base, [[' ', T.fg]]);
[[0, 'fosforo/', 'o'], [1, 'cmd/', 'o'], [2, 'fosforo/', 'o'], [3, 'main.go', 'f'], [1, 'internal/', 'o'], [2, 'buffer/', 'c'], [2, 'keymap/', 'o'], [3, 'trie.go', 'f'], [1, 'docs/', 'c'], [1, 'go.mod', 'f']].forEach((e) => {
  const col = e[2] === 'o' ? T.blue : (e[2] === 'c' ? T.dim : T.fg);
  line(base, [['  '.repeat(e[0]) + e[1], col, e[2] === 'o' && e[0] === 0 ? { fontWeight: 700 } : {}]]);
});
