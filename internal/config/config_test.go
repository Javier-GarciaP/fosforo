package config_test

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/Javier-GarciaP/fosforo/internal/config"
)

func TestDefaults(t *testing.T) {
	d := config.Defaults()
	if d.Theme != "dark" || d.Icons != "nerd" || d.TabWidth != 4 || d.Indent != "tabs" || d.ScrollOff != 3 || !d.Mouse || d.Keys == nil {
		t.Errorf("%+v", d)
	}
}

func TestPathFollowsXDG(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/tmp/xdg")
	if got := config.Path(); got != "/tmp/xdg/fos/config.toml" {
		t.Errorf("Path = %q", got)
	}
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("HOME", "/home/prueba")
	if got := config.Path(); got != "/home/prueba/.config/fos/config.toml" {
		t.Errorf("Path = %q", got)
	}
}

func TestLoadWithoutFileUsesDefaults(t *testing.T) {
	cfg, problems, err := config.Load(filepath.Join(t.TempDir(), "no-existe.toml"))
	if err != nil || len(problems) != 0 || !reflect.DeepEqual(cfg, config.Defaults()) {
		t.Errorf("sin archivo: %+v %v %v", cfg, problems, err)
	}
	if cfg, _, err := config.Load(""); err != nil || !reflect.DeepEqual(cfg, config.Defaults()) {
		t.Error("ruta vacía")
	}
}

func TestLoadValidFile(t *testing.T) {
	p := filepath.Join(t.TempDir(), "config.toml")
	os.WriteFile(p, []byte(`
theme = "light"
icons = "none"
tabwidth = 8
indent = "spaces"
scrolloff = 5
mouse = false

[normal]
"gh" = "cursor.line_start"
gl = "cursor.line_end"

[keys.explorer]
"x" = "explorer.delete"
`), 0o644)
	cfg, problems, err := config.Load(p)
	if err != nil || len(problems) != 0 {
		t.Fatalf("%v %v", problems, err)
	}
	if cfg.Theme != "light" || cfg.Icons != "none" || cfg.TabWidth != 8 || cfg.Indent != "spaces" || cfg.ScrollOff != 5 || cfg.Mouse {
		t.Errorf("%+v", cfg)
	}
	if cfg.Keys["normal"]["gh"] != "cursor.line_start" || cfg.Keys["normal"]["gl"] != "cursor.line_end" || cfg.Keys["explorer"]["x"] != "explorer.delete" {
		t.Errorf("atajos: %+v", cfg.Keys)
	}
}

func TestLoadUnreadableFileIsAnError(t *testing.T) {
	if _, _, err := config.Load(t.TempDir()); err == nil { // un directorio no se puede leer como archivo
		t.Error("debe devolver error")
	}
}

func TestInvalidValuesFallBackWithKeyNamedInProblem(t *testing.T) {
	for _, tc := range []struct {
		name, toml, key, reason string
		check                   func(config.Config) bool
	}{
		{"tabwidth texto", `tabwidth = "ocho"`, "tabwidth", "entre 1 y 16", func(c config.Config) bool { return c.TabWidth == 4 }},
		{"tabwidth fuera de rango", `tabwidth = 99`, "tabwidth", "entre 1 y 16", func(c config.Config) bool { return c.TabWidth == 4 }},
		{"tabwidth cero", `tabwidth = 0`, "tabwidth", "entre 1 y 16", func(c config.Config) bool { return c.TabWidth == 4 }},
		{"tabwidth decimal", `tabwidth = 4.5`, "tabwidth", "número entero", func(c config.Config) bool { return c.TabWidth == 4 }},
		{"scrolloff negativo", `scrolloff = -1`, "scrolloff", "entre 0 y 20", func(c config.Config) bool { return c.ScrollOff == 3 }},
		{"theme inválido", `theme = "azul"`, "theme", `"dark", "light" o "auto"`, func(c config.Config) bool { return c.Theme == "dark" }},
		{"theme número", `theme = 3`, "theme", "debe ser", func(c config.Config) bool { return c.Theme == "dark" }},
		{"icons inválido", `icons = "emoji"`, "icons", `"nerd" o "none"`, func(c config.Config) bool { return c.Icons == "nerd" }},
		{"indent inválido", `indent = "mixto"`, "indent", `"tabs" o "spaces"`, func(c config.Config) bool { return c.Indent == "tabs" }},
		{"mouse texto", `mouse = "si"`, "mouse", "true o false", func(c config.Config) bool { return c.Mouse }},
		{"clave desconocida", `colorcito = 1`, "colorcito", "desconocido", func(c config.Config) bool { return c.TabWidth == 4 }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg, problems := config.Parse([]byte(tc.toml))
			if len(problems) != 1 || problems[0].Key != tc.key || !strings.Contains(problems[0].Reason, tc.reason) {
				t.Fatalf("avisos = %+v", problems)
			}
			if !tc.check(cfg) {
				t.Errorf("debía quedar el valor por defecto: %+v", cfg)
			}
			if !strings.HasPrefix(problems[0].String(), tc.key+": ") {
				t.Errorf("String = %q", problems[0].String())
			}
		})
	}
}

func TestOneInvalidValueDoesNotAffectTheRest(t *testing.T) {
	cfg, problems := config.Parse([]byte("tabwidth = \"ocho\"\ntheme = \"light\"\nicons = \"none\"\n"))
	if len(problems) != 1 || cfg.Theme != "light" || cfg.Icons != "none" || cfg.TabWidth != 4 {
		t.Errorf("%+v %+v", cfg, problems)
	}
}

func TestSyntaxErrorFallsBackToDefaults(t *testing.T) {
	cfg, problems := config.Parse([]byte("theme = \"light\"\ntabwidth = = 4\n"))
	if !reflect.DeepEqual(cfg, config.Defaults()) {
		t.Errorf("tras un error de sintaxis se usan los valores por defecto: %+v", cfg)
	}
	if len(problems) != 1 || !strings.Contains(problems[0].Reason, "línea 2") {
		t.Errorf("el aviso indica la línea: %+v", problems)
	}
}

func TestKeyTablesValidation(t *testing.T) {
	cfg, problems := config.Parse([]byte(`
[normal]
"gh" = "cursor.line_start"
"gx" = 5

[nada]
"a" = "b"

[keys]
visual = "no es una tabla"

[keys.insert]
"ctrl+x" = ""
`))
	keys := map[string]string{}
	for _, p := range problems {
		keys[p.Key] = p.Reason
	}
	for _, want := range []string{"keys.normal.gx", "keys.visual"} {
		if _, ok := keys[want]; !ok {
			t.Errorf("falta el aviso de %s: %+v", want, problems)
		}
	}
	if _, ok := keys["nada"]; !ok {
		t.Errorf("una tabla desconocida se avisa como ajuste: %+v", problems)
	}
	if cfg.Keys["normal"]["gh"] != "cursor.line_start" {
		t.Error("lo válido se conserva")
	}
	if v, ok := cfg.Keys["insert"]["ctrl+x"]; !ok || v != "" {
		t.Errorf("un comando vacío quita el atajo: %+v", cfg.Keys["insert"])
	}
	// contexto desconocido dentro de keys
	_, p2 := config.Parse([]byte("[keys.raro]\n\"a\" = \"b\"\n"))
	if len(p2) != 1 || p2[0].Key != "keys.raro" || !strings.Contains(p2[0].Reason, "normal, insert") {
		t.Errorf("%+v", p2)
	}
}

func TestTemplateIsValidAndChangesNothing(t *testing.T) {
	cfg, problems := config.Parse([]byte(config.Template))
	if len(problems) != 0 || !reflect.DeepEqual(cfg, config.Defaults()) {
		t.Errorf("la plantilla (todo comentado) es válida y equivale a los valores por defecto: %+v %v", cfg, problems)
	}
	// quitando los comentarios de los ejemplos también es válida
	uncommented := strings.ReplaceAll(strings.ReplaceAll(config.Template, "# theme", "theme"), "# [normal]", "[normal]")
	uncommented = strings.ReplaceAll(uncommented, "# \"gh\"", "\"gh\"")
	if _, problems := config.Parse([]byte(uncommented)); len(problems) != 0 {
		t.Errorf("los ejemplos descomentados son válidos: %+v", problems)
	}
}
