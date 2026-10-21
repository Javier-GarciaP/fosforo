// Package config carga la configuración del usuario: un archivo TOML opcional sobre valores por defecto embebidos.
// Un valor inválido nunca impide arrancar: se avisa con el nombre de la clave y se usa el valor por defecto.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

// Config son los ajustes del editor.
type Config struct {
	Theme     string // "dark", "light" o "auto"
	Icons     string // "nerd" o "none"
	TabWidth  int    // ancho de un tabulador, 1..16
	Indent    string // "tabs" o "spaces": qué inserta la tecla Tab
	ScrollOff int    // líneas de margen al desplazarse, 0..20
	Mouse     bool   // capturar el ratón

	// Keys son los atajos del usuario por contexto (normal, insert, visual, explorer, preview): secuencia → comando.
	// Un comando vacío quita el atajo.
	Keys map[string]map[string]string
}

// Defaults devuelve los valores por defecto.
func Defaults() Config {
	return Config{Theme: "dark", Icons: "nerd", TabWidth: 4, Indent: "tabs", ScrollOff: 3, Mouse: true, Keys: map[string]map[string]string{}}
}

// Contexts son los contextos de teclas que se pueden remapear.
var Contexts = []string{"normal", "insert", "visual", "explorer", "preview"}

// Problem es un aviso sobre un ajuste: qué clave y por qué no se aplicó.
type Problem struct {
	Key    string
	Reason string
}

func (p Problem) String() string { return p.Key + ": " + p.Reason }

// Path devuelve la ruta del archivo de configuración: $XDG_CONFIG_HOME/fos/config.toml o ~/.config/fos/config.toml.
func Path() string {
	if dir := os.Getenv("XDG_CONFIG_HOME"); dir != "" {
		return filepath.Join(dir, "fos", "config.toml")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "fos", "config.toml")
}

// Load lee path. Si el archivo no existe devuelve los valores por defecto sin avisos. Un error de sintaxis o un valor
// inválido se devuelve como aviso y los valores afectados quedan por defecto; Load solo falla si el archivo no se puede leer.
func Load(path string) (Config, []Problem, error) {
	cfg := Defaults()
	if path == "" {
		return cfg, nil, nil
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return cfg, nil, nil
	}
	if err != nil {
		return cfg, nil, err
	}
	cfg, problems := Parse(data)
	return cfg, problems, nil
}

// Parse interpreta el contenido de un archivo de configuración.
func Parse(data []byte) (Config, []Problem) {
	cfg := Defaults()
	var raw map[string]any
	if err := toml.Unmarshal(data, &raw); err != nil {
		reason := err.Error()
		var de *toml.DecodeError
		if errors.As(err, &de) {
			row, col := de.Position()
			reason = fmt.Sprintf("error de sintaxis en la línea %d, columna %d: %s", row, col, de.Error())
			if i := strings.Index(de.Error(), "\n"); i >= 0 {
				reason = fmt.Sprintf("error de sintaxis en la línea %d, columna %d", row, col)
			}
		}
		return cfg, []Problem{{Key: "config.toml", Reason: reason + " (se usan los valores por defecto)"}}
	}

	var problems []Problem
	bad := func(key, reason string, def any) {
		problems = append(problems, Problem{Key: key, Reason: fmt.Sprintf("%s (se usa %v)", reason, def)})
	}
	str := func(key string, allowed ...string) (string, bool) {
		v, ok := raw[key]
		if !ok {
			return "", false
		}
		delete(raw, key)
		s, isStr := v.(string)
		if isStr {
			for _, a := range allowed {
				if s == a {
					return s, true
				}
			}
		}
		return "", false
	}
	choice := func(key, def string, allowed ...string) string {
		if _, present := raw[key]; !present {
			return def
		}
		if s, ok := str(key, allowed...); ok {
			return s
		}
		bad(key, "debe ser "+quoteList(allowed), def)
		return def
	}
	integer := func(key string, def, lo, hi int) int {
		v, present := raw[key]
		if !present {
			return def
		}
		delete(raw, key)
		n, ok := v.(int64)
		if !ok || n < int64(lo) || n > int64(hi) {
			bad(key, fmt.Sprintf("debe ser un número entero entre %d y %d", lo, hi), def)
			return def
		}
		return int(n)
	}

	cfg.Theme = choice("theme", cfg.Theme, "dark", "light", "auto")
	cfg.Icons = choice("icons", cfg.Icons, "nerd", "none")
	cfg.Indent = choice("indent", cfg.Indent, "tabs", "spaces")
	cfg.TabWidth = integer("tabwidth", cfg.TabWidth, 1, 16)
	cfg.ScrollOff = integer("scrolloff", cfg.ScrollOff, 0, 20)
	if v, present := raw["mouse"]; present {
		delete(raw, "mouse")
		if b, ok := v.(bool); ok {
			cfg.Mouse = b
		} else {
			bad("mouse", "debe ser true o false", cfg.Mouse)
		}
	}

	// Atajos: [keys.normal] o [normal] (también normal."gh" = "...") con secuencia = comando.
	keyTables := map[string]any{}
	if k, ok := raw["keys"]; ok {
		delete(raw, "keys")
		if t, isTable := k.(map[string]any); isTable {
			for ctx, v := range t {
				keyTables[ctx] = v
			}
		} else {
			problems = append(problems, Problem{Key: "keys", Reason: "debe ser una tabla con un contexto por clave (se ignora)"})
		}
	}
	for _, ctx := range Contexts {
		if v, ok := raw[ctx]; ok {
			delete(raw, ctx)
			keyTables[ctx] = v
		}
	}
	ctxs := make([]string, 0, len(keyTables))
	for c := range keyTables {
		ctxs = append(ctxs, c)
	}
	sort.Strings(ctxs)
	for _, ctx := range ctxs {
		if !contains(Contexts, ctx) {
			problems = append(problems, Problem{Key: "keys." + ctx, Reason: "contexto desconocido; los válidos son " + strings.Join(Contexts, ", ") + " (se ignora)"})
			continue
		}
		t, ok := keyTables[ctx].(map[string]any)
		if !ok {
			problems = append(problems, Problem{Key: "keys." + ctx, Reason: "debe ser una tabla de secuencia = comando (se ignora)"})
			continue
		}
		seqs := make([]string, 0, len(t))
		for s := range t {
			seqs = append(seqs, s)
		}
		sort.Strings(seqs)
		for _, seq := range seqs {
			cmd, isStr := t[seq].(string)
			if !isStr {
				problems = append(problems, Problem{Key: "keys." + ctx + "." + seq, Reason: "el comando debe ser un texto (se ignora)"})
				continue
			}
			if cfg.Keys[ctx] == nil {
				cfg.Keys[ctx] = map[string]string{}
			}
			cfg.Keys[ctx][seq] = cmd
		}
	}

	unknown := make([]string, 0, len(raw))
	for k := range raw {
		unknown = append(unknown, k)
	}
	sort.Strings(unknown)
	for _, k := range unknown {
		problems = append(problems, Problem{Key: k, Reason: "ajuste desconocido (se ignora)"})
	}
	return cfg, problems
}

func quoteList(vals []string) string {
	q := make([]string, len(vals))
	for i, v := range vals {
		q[i] = `"` + v + `"`
	}
	if len(q) == 1 {
		return q[0]
	}
	return strings.Join(q[:len(q)-1], ", ") + " o " + q[len(q)-1]
}

func contains(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}

// Template es el contenido inicial de un archivo de configuración nuevo (todo comentado: son los valores por defecto).
const Template = `# Configuración de fos. Todo es opcional: lo que no pongas usa el valor por defecto.
# Después de guardar este archivo se aplica solo; también puedes ejecutar :reload.

# theme = "dark"       # "dark", "light" o "auto" (según el fondo del terminal)
# icons = "nerd"       # "nerd" (necesita una Nerd Font) o "none"
# tabwidth = 4         # ancho de un tabulador (1-16)
# indent = "tabs"      # "tabs" o "spaces": qué inserta la tecla Tab
# scrolloff = 3        # líneas de margen al desplazarse (0-20)
# mouse = true         # capturar el ratón

# Atajos por contexto (normal, insert, visual, explorer, preview): secuencia = "comando".
# Una secuencia de varias teclas se escribe junta ("gh") o con espacios ("space f f"); un comando "" quita el atajo.
# [normal]
# "gh" = "cursor.line_start"
# "gl" = "cursor.line_end"
`
