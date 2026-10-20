// 00 - Guia de estilo: paleta, colores de modo y glifos (oscuro y claro)
const T0 = THEMES.dark;
const r = screen('00 Guia de estilo', 0, 0, T0);
const base = I(r, { type: 'frame', name: 'columnas', x: 0, y: 0, width: W, height: H, layout: 'horizontal', gap: 0, fill: [] });
function column(T, titulo) {
  const c = I(base, { type: 'frame', name: titulo, layout: 'vertical', gap: 0, width: 'fill_container', height: 'fill_container', fill: solid(T.bg), padding: [18, 18, 18, 18] });
  line(c, [[titulo, T.fg, { fontWeight: 700 }]]);
  line(c, [[' ', T.fg]]);
  const names = [['bg', 'fondo del editor'], ['panel', 'explorador, tabs, statusline'], ['sel', 'linea actual / seleccion en lista'], ['visual', 'seleccion visual'],
    ['fg', 'texto'], ['dim', 'secundario, numeros de linea'], ['blue', 'carpetas, funciones, NORMAL'], ['green', 'cadenas, INSERT, exito'], ['mauve', 'palabras clave, VISUAL, lider'],
    ['peach', 'numeros, [+], COMMAND'], ['yellow', 'tipos, coincidencias'], ['red', 'errores'], ['teal', 'markdown, PREVIEW'], ['border', 'bordes y separadores'], ['code', 'bloque de codigo md']];
  names.forEach((n) => {
    const row = I(c, { type: 'frame', layout: 'horizontal', gap: 0, height: CH, width: 'fill_container', alignItems: 'center', fill: [] });
    I(row, { type: 'rectangle', width: 45, height: 14, fill: solid(T[n[0]]), stroke: { thickness: 1, fill: solid(T.border) } });
    txt(row, ' ' + n[0].padEnd(7) + T[n[0]] + '  ', T.fg);
    txt(row, n[1], T.dim);
  });
  line(c, [[' ', T.fg]]);
  line(c, [['Modos', T.fg, { fontWeight: 700 }]]);
  const modes = I(c, { type: 'frame', layout: 'horizontal', gap: 9, height: CH, alignItems: 'center', fill: [] });
  Object.keys(MODE_COLOR).forEach((m) => span(modes, [' ' + m + ' ', T.bg, { bg: T[MODE_COLOR[m]], fontWeight: 700 }]));
  line(c, [[' ', T.fg]]);
  line(c, [['Glifos', T.fg, { fontWeight: 700 }]]);
  line(c, [['bordes   ', T.dim], ['caracteres de caja Unicode U+2500-257F', T.border]]);
  const gr = I(c, { type: 'frame', layout: 'horizontal', gap: 0, height: CH, width: 'fill_container', alignItems: 'center', fill: [] });
  txt(gr, 'carpetas ', T.dim); iconNode(gr, T, 'folder'); txt(gr, 'cerrada  ', T.blue); iconNode(gr, T, 'folder_open'); txt(gr, 'abierta', T.blue);
  line(c, [['estado   ', T.dim], ['[+] ', T.peach], ['✓ ', T.green], ['✗ ', T.red], ['☑ ☐ ', T.fg], ['▏ cursor  ~ vacio', T.dim]]);
  line(c, [['md       ', T.dim], ['• lista  ▎ cita  ─ regla', T.teal]]);
  return c;
}
column(THEMES.dark, 'TEMA OSCURO  (paleta Tinta)');
column(THEMES.light, 'TEMA CLARO  (paleta Tinta)');
