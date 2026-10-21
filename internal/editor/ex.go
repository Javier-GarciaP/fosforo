package editor

import (
	"strconv"
	"strings"
	"unicode"

	"github.com/Javier-GarciaP/fosforo/internal/command"
)

// exCommands relaciona los nombres de la línea de comandos «:» con comandos del registro.
var exCommands = map[string]string{
	"w": "file.save", "write": "file.save",
	"q": "app.quit", "quit": "app.quit",
	"wq": "file.save_quit",
	"x":  "file.update_quit", "xit": "file.update_quit",
	"e": "file.open", "edit": "file.open",
}

// parseEx divide una línea de «:» en nombre, «!» y argumento: «w! ruta» → ("w", true, "ruta").
func parseEx(line string) (name string, bang bool, arg string) {
	line = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), ":"))
	i := 0
	for i < len(line) && unicode.IsLetter(rune(line[i])) {
		i++
	}
	name, rest := line[:i], line[i:]
	if strings.HasPrefix(rest, "!") {
		bang, rest = true, rest[1:]
	}
	return name, bang, strings.TrimSpace(rest)
}

// runEx ejecuta una línea de la línea de comandos «:».
func (m *Model) runEx(line string) {
	line = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), ":"))
	if line == "" {
		return
	}
	if n, err := strconv.Atoi(line); err == nil && n >= 0 {
		m.dispatchByName("cursor.goto_line", command.Args{Count: n, HasCount: true})
		return
	}
	name, bang, arg := parseEx(line)
	cmd, ok := exCommands[name]
	if !ok {
		m.fail(command.UnknownError{Name: line})
		return
	}
	m.dispatchByName(cmd, command.Args{Bang: bang, Text: arg})
}

func (m *Model) fail(err error) {
	if ue, ok := err.(command.UnknownError); ok {
		m.msg, m.msgErr = "E492: "+ue.Error(), true
		return
	}
	m.msg, m.msgErr = err.Error(), true
}
