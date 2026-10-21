package editor_test

import (
	"image/color"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/Javier-GarciaP/fosforo/internal/buffer"
	"github.com/Javier-GarciaP/fosforo/internal/config"
	"github.com/Javier-GarciaP/fosforo/internal/editor"
)

func cfgWith(f func(c *config.Config)) config.Config {
	c := config.Defaults()
	f(&c)
	return c
}

func TestConfigAppliesSettings(t *testing.T) {
	t.Cleanup(func() { buffer.SetTabWidth(4) })
	cfg := cfgWith(func(c *config.Config) {
		c.Theme = "light"
		c.TabWidth = 8
		c.Indent = "spaces"
		c.Mouse = false
		c.ScrollOff = 0
	})
	e := newEd(t, "ab\n", editor.WithConfig(cfg, "", nil))
	if e.m.Light() != true || e.m.MouseEnabled() || buffer.TabWidth != 8 {
		t.Fatalf("light=%v mouse=%v tab=%d", e.m.Light(), e.m.MouseEnabled(), buffer.TabWidth)
	}
	e.typ("i").send(tea.KeyPressMsg{Code: tea.KeyTab})
	if got := e.text(); got != strings.Repeat(" ", 8)+"ab\n" {
		t.Errorf("Tab con indent=spaces inserta 8 espacios, hay %q", got)
	}
	e = newEd(t, "ab\n", editor.WithConfig(config.Defaults(), "", nil)).typ("i").send(tea.KeyPressMsg{Code: tea.KeyTab})
	if got := e.text(); got != "\tab\n" {
		t.Errorf("Tab con indent=tabs inserta un tabulador, hay %q", got)
	}
}

func TestConfigIconsAndProblemsMessage(t *testing.T) {
	cfg, problems := config.Parse([]byte("icons = \"none\"\ntabwidth = \"ocho\"\n"))
	t.Cleanup(func() { buffer.SetTabWidth(4) })
	e := newEd(t, "x\n", editor.WithConfig(cfg, "", problems))
	if !strings.Contains(e.m.Message(), "tabwidth") {
		t.Errorf("debe avisar de tabwidth: %q", e.m.Message())
	}
	if buffer.TabWidth != 4 {
		t.Errorf("tabwidth inválido usa el valor por defecto: %d", buffer.TabWidth)
	}
	if e.m.IconsOn() {
		t.Error("icons = none")
	}
}

func TestConfigRemapAndReject(t *testing.T) {
	cfg := cfgWith(func(c *config.Config) {
		c.Keys = map[string]map[string]string{"normal": {
			"H":       "buffer.prev",      // atajo nuevo
			"x":       "edit.paste_after", // reasigna uno de fábrica
			"q q":     "no.existe",        // comando inexistente: se rechaza
			"space t": "",                 // quitar
			"g":       "file.save",        // choca (es prefijo de gg): se rechaza
		}}
	})
	e := newEd(t, "abc\n", editor.WithConfig(cfg, "", nil))
	msg := e.m.Message()
	if !strings.Contains(msg, "Configuración") {
		t.Fatalf("debe avisar de los rechazos: %q", msg)
	}
	e.typ("x") // ahora pega (registro vacío): no borra
	if e.text() != "abc\n" {
		t.Errorf("x remapeado ya no borra: %q", e.text())
	}
	before := e.m.Light()
	e.space("t")
	if e.m.Light() != before {
		t.Error("Space t quitado no debe alternar el tema")
	}
	e.typ("gg") // el prefijo g sigue funcionando (el remapeo conflictivo se rechazó)
	e.at(0, 0)
}

func TestConfigRemapUsesCommandsFromPalette(t *testing.T) {
	cfg := cfgWith(func(c *config.Config) { c.Keys = map[string]map[string]string{"normal": {"space z": "theme.toggle"}} })
	e := newEd(t, "x\n", editor.WithConfig(cfg, "", nil)).space("z")
	if !e.m.Light() {
		t.Error("Space z debe alternar el tema")
	}
}

func TestThemeAutoFollowsTerminalBackground(t *testing.T) {
	cfg := cfgWith(func(c *config.Config) { c.Theme = "auto" })
	e := newEd(t, "x\n", editor.WithConfig(cfg, "", nil))
	if e.m.Light() {
		t.Error("sin respuesta de la terminal se queda oscuro")
	}
	if e.m.Init() == nil {
		t.Error("theme=auto debe pedir el color de fondo")
	}
	e.send(tea.BackgroundColorMsg{Color: color.RGBA{R: 250, G: 250, B: 250, A: 255}})
	if !e.m.Light() {
		t.Error("fondo claro → tema claro")
	}
	e.send(tea.BackgroundColorMsg{Color: color.RGBA{A: 255}})
	if e.m.Light() {
		t.Error("fondo oscuro → tema oscuro")
	}
	// con un tema fijo se ignora lo que diga la terminal
	e = newEd(t, "x\n", editor.WithConfig(config.Defaults(), "", nil))
	e.send(tea.BackgroundColorMsg{Color: color.RGBA{R: 255, G: 255, B: 255, A: 255}})
	if e.m.Light() || e.m.Init() != nil {
		t.Error("tema fijo: ni se pide ni se aplica el fondo")
	}
}

func TestConfigOpenCreatesTemplateAndReloadsOnSave(t *testing.T) {
	t.Cleanup(func() { buffer.SetTabWidth(4) })
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "config.toml")
	m, err := editor.Open("", editor.WithConfig(config.Defaults(), path, nil))
	if err != nil {
		t.Fatal(err)
	}
	e := &ed{t: t, m: m}
	e.send(tea.WindowSizeMsg{Width: 100, Height: 24})
	e.space("co")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("config.open crea el archivo: %v", err)
	}
	if e.m.Buffer().Path() != path || !strings.Contains(e.text(), "Configuración de fos") {
		t.Fatalf("debe abrir la plantilla: %q msg=%q", e.m.Buffer().Path(), e.m.Message())
	}
	// se edita y se guarda: se recarga sola
	e.typ("ggO").typ(`theme = "light"`).send(tea.KeyPressMsg{Code: tea.KeyEscape}).ex("w")
	if !e.m.Light() || !strings.Contains(e.m.Message(), "recargada") {
		t.Errorf("guardar la configuración la recarga: light=%v msg=%q", e.m.Light(), e.m.Message())
	}
	// recarga manual tras cambiar el archivo desde fuera
	if err := os.WriteFile(path, []byte("theme = \"dark\"\ntabwidth = 2\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	e.space("cr")
	if e.m.Light() || buffer.TabWidth != 2 {
		t.Errorf("config.reload: light=%v tab=%d", e.m.Light(), buffer.TabWidth)
	}
	// un valor inválido al recargar avisa y no rompe
	os.WriteFile(path, []byte("theme = 5\n"), 0o644)
	e.space("cr")
	if !strings.Contains(e.m.Message(), "theme") {
		t.Errorf("aviso de theme: %q", e.m.Message())
	}
}

func TestConfigReloadSwitchesIcons(t *testing.T) {
	path := filepath.Join(t.TempDir(), "c.toml")
	os.WriteFile(path, []byte("icons = \"none\"\n"), 0o644)
	e := newEd(t, "x\n", editor.WithConfig(config.Defaults(), path, nil))
	if !e.m.IconsOn() {
		t.Fatal("empieza con iconos")
	}
	e.space("cr")
	if e.m.IconsOn() {
		t.Error("tras recargar, sin iconos")
	}
}
