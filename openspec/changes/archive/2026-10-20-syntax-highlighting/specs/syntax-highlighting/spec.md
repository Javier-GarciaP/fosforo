# Spec Delta

## Purpose
Define cómo se colorea el código fuente según su lenguaje y el tema activo, sin degradar la capacidad de respuesta del editor.

## ADDED Requirements

### Requirement: Language detection
The system SHALL choose the language from the file name or extension and render files of unknown type as plain text.

#### Scenario: Archivo Go
- **WHEN** se abre `trie.go`
- **THEN** las palabras clave, tipos, funciones, cadenas, números y comentarios se muestran con los colores de la paleta (mauve, amarillo, azul, verde, melocotón y gris respectivamente)

#### Scenario: Tipo desconocido
- **WHEN** se abre un archivo con extensión sin lenguaje conocido
- **THEN** se muestra como texto plano sin errores

### Requirement: Theme-driven colors
Token colors MUST come from the active theme (dark or light palette) and SHALL update immediately when the theme changes.

#### Scenario: Cambiar a tema claro
- **WHEN** el usuario alterna a tema claro con `Space t`
- **THEN** el código se recolorea con la paleta clara sin reabrir el archivo

### Requirement: Non-blocking highlighting
Highlighting SHALL NOT block typing or scrolling: until the tokens for the current version of the text are ready, the editor MUST keep rendering the text, with the previous colors or plain.

#### Scenario: Escribir rápido
- **WHEN** el usuario escribe continuamente en un archivo de varios miles de líneas
- **THEN** cada pulsación se refleja en pantalla sin esperar al resaltado, y los colores se actualizan al terminar la tokenización

### Requirement: Multi-line constructs
Constructs that span several lines (block comments, raw or multi-line strings) SHALL be colored correctly, including after edits that open or close them.

#### Scenario: Abrir un comentario de bloque
- **WHEN** se escribe `/*` antes de varias líneas de código
- **THEN** esas líneas pasan a verse como comentario hasta el cierre `*/`

### Requirement: Composition with selection and cursor line
Syntax colors MUST remain visible under the current-line background and the Visual selection background.

#### Scenario: Selección Visual
- **WHEN** hay texto resaltado por sintaxis dentro de una selección Visual
- **THEN** los colores de los tokens se conservan sobre el fondo de la selección

### Requirement: Safety limits
The system SHALL skip highlighting for files larger than 2 MiB and for lines longer than 2000 characters, showing the text plain without errors.

#### Scenario: Línea gigantesca
- **WHEN** un archivo contiene una línea minificada de 100000 caracteres
- **THEN** esa línea se muestra sin color y el resto del archivo sigue coloreado
