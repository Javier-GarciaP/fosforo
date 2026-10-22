# Spec Delta

## Purpose
Mostrar una pantalla de inicio con la marca FOSFORO y los atajos principales cuando no hay ningún archivo abierto, para orientar al usuario sin cambiar cómo funciona el editor.

## ADDED Requirements

### Requirement: Start screen visibility
The system SHALL show the start screen in the editor area only while the active buffer is empty, has no file name and is not modified. It MUST disappear as soon as any of those conditions stops being true.

#### Scenario: Abrir una carpeta
- **WHEN** se ejecuta `fos .` y el área del editor es lo bastante grande
- **THEN** el editor muestra el logo y la lista de atajos, y el explorador sigue visible a la izquierda

#### Scenario: Abrir un archivo
- **WHEN** el usuario abre un archivo desde el explorador o con `Space f f`
- **THEN** la pantalla de inicio desaparece y se muestra el texto del archivo con sus números de línea

#### Scenario: Empezar a escribir
- **WHEN** el usuario pulsa `i` y escribe un carácter en el buffer vacío
- **THEN** la pantalla de inicio desaparece y se muestra el texto escrito

#### Scenario: Cerrar el último archivo
- **WHEN** el usuario cierra el último buffer abierto y queda un buffer vacío sin nombre
- **THEN** la pantalla de inicio vuelve a mostrarse

#### Scenario: Archivo existente vacío
- **WHEN** el usuario abre un archivo que existe pero está vacío
- **THEN** no se muestra la pantalla de inicio, porque el buffer tiene nombre

### Requirement: Start screen content
The start screen SHALL show the word "FOSFORO" in large block letters using the theme `Brand` color, centered horizontally, and below it a list of actions with their leader shortcut on the right: buscar archivo (`Space f f`), nuevo archivo (`Space f n`), archivos recientes (`Space f r`), buscar texto (`Space s g`), explorador (`Space e`), configuración (`Space c o`), lista de atajos (`Space ?`) and salir (`Space Q`). Every shortcut shown MUST be bound to the action it is listed next to.

#### Scenario: Contenido completo
- **WHEN** la pantalla de inicio se muestra en un área de al menos 60 columnas y 24 filas
- **THEN** aparecen el logo en color `Brand`, las ocho acciones con sus atajos alineados a la derecha, y todo queda centrado en el área del editor

#### Scenario: Cambio de tema
- **WHEN** el usuario pulsa `Space t` con la pantalla de inicio visible
- **THEN** el logo y la lista se redibujan con los colores del otro tema

#### Scenario: Atajo real
- **WHEN** el usuario pulsa `Space f f` con la pantalla de inicio visible
- **THEN** se abre el buscador de archivos, igual que sin pantalla de inicio

### Requirement: Start screen is informational
The start screen MUST NOT change how keys or the mouse behave, and its rows MUST NOT react to mouse clicks. The cursor SHALL NOT be shown over it.

#### Scenario: Clic sobre una fila
- **WHEN** el usuario hace clic sobre una acción de la lista
- **THEN** no ocurre nada

#### Scenario: Teclas normales
- **WHEN** el usuario pulsa `:` con la pantalla de inicio visible
- **THEN** se abre la línea de comandos como siempre

### Requirement: Small windows
The system SHALL adapt the start screen to the available area: if the full screen does not fit it MUST hide the logo and keep only the list, and if the list does not fit either it MUST show the normal empty editor. It MUST NOT draw outside the editor area or break the layout.

#### Scenario: Área baja
- **WHEN** el área del editor tiene menos filas de las que necesitan el logo y la lista juntos, pero sí las de la lista
- **THEN** se muestra solo la lista de atajos, centrada

#### Scenario: Área muy pequeña
- **WHEN** el área del editor no tiene espacio ni para la lista
- **THEN** se muestra el editor vacío normal con sus `~`

#### Scenario: Terminal estrecha con explorador
- **WHEN** el explorador está visible y el área restante es más estrecha que la lista
- **THEN** se muestra el editor vacío normal, sin cortar texto a medias
