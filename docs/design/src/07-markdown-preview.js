// 07 - Archivo .md abierto en vista previa (por defecto)
const T = THEMES.dark;
const r = screen('07 Markdown vista previa', 1, 2, T);
const a = appFrame(r, T);
const body = I(a, { type: 'frame', name: 'cuerpo', layout: 'horizontal', gap: 0, width: 'fill_container', height: 34 * CH, fill: [] });
tree(body, T, 12); vsep(body, T);
mdPreview(body, T);
statusline(a, T, 'PREVIEW', 'README.md  (e: editar   Space m s: dividir)', 'markdown  utf-8  LF   28 lineas  0%');
line(a, [[' ', T.fg]]);
