// 08 - Markdown dividido: fuente | vista previa sincronizada
const T = THEMES.dark;
const r = screen('08 Markdown dividido', 2, 2, T);
const a = appFrame(r, T);
const body = I(a, { type: 'frame', name: 'cuerpo', layout: 'horizontal', gap: 0, width: 'fill_container', height: 34 * CH, fill: [] });
const BT = String.fromCharCode(96); // backtick: el preprocesador del DSL no admite el caracter literal
const h = T.blue, d = T.dim, f = T.fg, pe = T.peach, g = T.green;
const src = [
  [['# ', T.mauve], ['ted — editor de código en la terminal', h, { fontWeight: 700 }]], [],
  [['Un editor ligero con la fluidez de ', f], ['**Neovim**', pe], [' y la', f]], [['comodidad de VS Code, hecho en Go.', f]], [],
  [['## ', T.mauve], ['Características', h, { fontWeight: 700 }]], [],
  [['- ', pe], ['Teclado modal con tecla líder ', f], [BT + 'Space' + BT, g]], [['- ', pe], ['Explorador de archivos y paleta de comandos', f]],
  [['- [x] ', pe], ['Vista previa de Markdown', f]], [['- [ ] ', pe], ['Búsqueda en proyecto', f]], [],
  [['## ', T.mauve], ['Instalación', h, { fontWeight: 700 }]], [],
  [[BT + BT + BT + 'sh', d]], [['go install github.com/usuario/ted@latest', g]], [['ted README.md', g]], [[BT + BT + BT, d]], [],
  [['> ', T.blue], ['Nota: los .md se abren en vista previa.', d]], [],
  [['| Tecla | Acción |', f]], [['|-------|--------|', d]], [['| Space f f | Buscar archivo |', f]],
];
editor(body, T, { title: 'README.md', lines: src, curLine: 10, name: 'fuente' });
vsep(body, T);
mdPreview(body, T, { rule: 40 });
statusline(a, T, 'NORMAL', 'README.md [+]  | vista previa sincronizada', 'markdown  utf-8  LF   10:7  35%');
line(a, [[' ', T.fg]]);
