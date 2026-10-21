package editor

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Javier-GarciaP/fosforo/internal/buffer"
	"github.com/Javier-GarciaP/fosforo/internal/config"
	"github.com/Javier-GarciaP/fosforo/internal/keymap"
	"github.com/Javier-GarciaP/fosforo/internal/ui"
)

// WithConfig aplica la configuración cargada desde path (puede estar vacío si no hay archivo). Los avisos de problems
// se muestran en la línea de mensajes al arrancar.
func WithConfig(cfg config.Config, path string, problems []config.Problem) Option {
	return func(m *Model) {
		m.configPath = path
		m.applyConfig(cfg, problems)
	}
}

// applyConfig aplica los ajustes al editor y avisa de los que no se pudieron aplicar.
func (m *Model) applyConfig(cfg config.Config, problems []config.Problem) {
	switch cfg.Theme {
	case "light":
		m.light, m.themeAuto = true, false
	case "auto":
		m.themeAuto = true // el color real llega después, con BackgroundColorMsg
	default:
		m.light, m.themeAuto = false, false
	}
	if cfg.Icons == "none" {
		m.icons = ui.IconsNone
	} else {
		m.icons = ui.IconsNerd
	}
	buffer.SetTabWidth(cfg.TabWidth)
	m.spaces = cfg.Indent == "spaces"
	m.scrollOff = cfg.ScrollOff
	m.mouseOn = cfg.Mouse

	km, keyProblems := buildUserKeymap(cfg.Keys)
	m.keys = km
	problems = append(problems, keyProblems...)
	if len(problems) > 0 {
		m.msg, m.msgErr = problemsMessage(problems), true
	}
}

// buildUserKeymap parte de los atajos de fábrica y aplica los del usuario. Un comando inexistente, una secuencia
// que choca con otra o un contexto desconocido se rechazan con un aviso y se conserva el atajo de fábrica.
func buildUserKeymap(user map[string]map[string]string) (*keymap.Keymap, []config.Problem) {
	if len(user) == 0 {
		return defaultKeymap, nil
	}
	km := defaultKeymap.Clone()
	var problems []config.Problem
	for _, ctx := range config.Contexts {
		table := user[ctx]
		seqs := make([]string, 0, len(table))
		for seq := range table {
			seqs = append(seqs, seq)
		}
		sort.Strings(seqs)
		for _, seq := range seqs {
			name := table[seq]
			key := "keys." + ctx + "." + seq
			if name == "" {
				if !km.Unbind(ctx, seq) {
					problems = append(problems, config.Problem{Key: key, Reason: "no había ningún atajo que quitar"})
				}
				continue
			}
			if _, ok := registry.Get(name); !ok {
				problems = append(problems, config.Problem{Key: key, Reason: fmt.Sprintf("el comando %q no existe (se ignora)", name)})
				continue
			}
			trial := km.Clone()
			trial.Unbind(ctx, seq)
			if err := trial.Bind(ctx, seq, name); err != nil {
				problems = append(problems, config.Problem{Key: key, Reason: err.Error() + " (se ignora)"})
				continue
			}
			km = trial
		}
	}
	return km, problems
}

// problemsMessage resume los avisos de configuración en una línea.
func problemsMessage(ps []config.Problem) string {
	msg := "Configuración: " + ps[0].String()
	if len(ps) > 1 {
		msg += fmt.Sprintf(" (y %d avisos más)", len(ps)-1)
	}
	return msg
}

// tabText es lo que inserta la tecla Tab: un tabulador o los espacios hasta la siguiente parada.
func (m *Model) tabText() string {
	if !m.spaces {
		return "\t"
	}
	v := buffer.VisualCol(m.buf.Line(m.cur.Line), m.cur.Col)
	return strings.Repeat(" ", buffer.TabWidth-v%buffer.TabWidth)
}

// isConfigFile indica si path es el archivo de configuración.
func (m *Model) isConfigFile(path string) bool {
	if m.configPath == "" || path == "" {
		return false
	}
	return absPath(path) == absPath(m.configPath)
}

// openConfig abre el archivo de configuración; si no existe lo crea con una plantilla comentada.
func (m *Model) openConfig() {
	if m.configPath == "" {
		m.msg, m.msgErr = "No se pudo determinar la ruta de la configuración", true
		return
	}
	if _, err := os.Stat(m.configPath); os.IsNotExist(err) {
		werr := os.MkdirAll(filepath.Dir(m.configPath), 0o755)
		if werr == nil {
			werr = os.WriteFile(m.configPath, []byte(config.Template), 0o644)
		}
		if werr != nil {
			m.msg, m.msgErr = "No se pudo crear la configuración: "+werr.Error(), true
			return
		}
	}
	m.openPath(m.configPath, m.configPath)
}

// reloadConfig vuelve a leer el archivo de configuración y lo aplica sin reiniciar.
func (m *Model) reloadConfig() {
	if m.configPath == "" {
		m.msg, m.msgErr = "No se pudo determinar la ruta de la configuración", true
		return
	}
	cfg, problems, err := config.Load(m.configPath)
	if err != nil {
		m.msg, m.msgErr = "No se pudo leer la configuración: "+err.Error(), true
		return
	}
	m.msg, m.msgErr = "Configuración recargada", false
	m.applyConfig(cfg, problems)
}

// Light indica si se usa el tema claro.
func (m Model) Light() bool { return m.light }

// IconsOn indica si el explorador muestra iconos.
func (m Model) IconsOn() bool { return m.icons == ui.IconsNerd }
