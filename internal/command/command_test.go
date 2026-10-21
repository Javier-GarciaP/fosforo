package command_test

import (
	"errors"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/Javier-GarciaP/fosforo/internal/command"
)

type state struct{ log []string }

func cmd(name string) command.Command[*state] {
	return command.Command[*state]{Name: name, Desc: "descripción de " + name, Keys: []string{"x"},
		Run: func(s *state, a command.Args) tea.Cmd { s.log = append(s.log, name); return nil }}
}

func TestRegisterValidation(t *testing.T) {
	r := command.NewRegistry[*state]()
	for _, bad := range []string{"", "guardar", "file", "File.save", "file.", ".save", "file.save.now", "file-save.x", "1file.save"} {
		c := cmd("file.save")
		c.Name = bad
		if err := r.Register(c); err == nil {
			t.Errorf("el nombre %q debía rechazarse", bad)
		}
	}
	noDesc := cmd("file.save")
	noDesc.Desc = ""
	if r.Register(noDesc) == nil {
		t.Error("sin descripción debía rechazarse")
	}
	noRun := cmd("file.save")
	noRun.Run = nil
	if r.Register(noRun) == nil {
		t.Error("sin función debía rechazarse")
	}
	if err := r.Register(cmd("file.save")); err != nil {
		t.Fatal(err)
	}
	if err := r.Register(cmd("file.save")); err == nil || !strings.Contains(err.Error(), "ya está registrado") {
		t.Errorf("duplicado: %v", err)
	}
	if err := r.Register(cmd("cursor.down_2")); err != nil {
		t.Errorf("guion bajo y dígitos son válidos: %v", err)
	}
}

func TestExecRunsAndUnknownDoesNothing(t *testing.T) {
	r := command.NewRegistry[*state]()
	r.MustRegister(cmd("file.save"))
	s := &state{}
	if _, err := r.Exec(s, "file.save", command.Args{}); err != nil || len(s.log) != 1 {
		t.Fatalf("exec: %v %v", err, s.log)
	}
	_, err := r.Exec(s, "file.nada", command.Args{})
	var ue command.UnknownError
	if !errors.As(err, &ue) || ue.Name != "file.nada" || !errors.Is(err, command.ErrUnknown) {
		t.Errorf("error = %v", err)
	}
	if len(s.log) != 1 {
		t.Error("un comando desconocido no debe ejecutar nada")
	}
	if !strings.Contains(err.Error(), "no es un comando del editor: file.nada") {
		t.Errorf("mensaje: %q", err)
	}
}

func TestListIsSortedWithMetadata(t *testing.T) {
	r := command.NewRegistry[*state]()
	for _, n := range []string{"file.save", "cursor.down", "app.quit"} {
		r.MustRegister(cmd(n))
	}
	list := r.List()
	if len(list) != 3 || list[0].Name != "app.quit" || list[1].Name != "cursor.down" || list[2].Name != "file.save" {
		t.Fatalf("lista: %+v", list)
	}
	for _, c := range list {
		if c.Desc == "" || len(c.Keys) == 0 {
			t.Errorf("%s: faltan descripción o atajos", c.Name)
		}
	}
	if c, ok := r.Get("file.save"); !ok || c.Name != "file.save" {
		t.Error("Get")
	}
	if _, ok := r.Get("nada.nada"); ok {
		t.Error("Get de un comando inexistente")
	}
}

func TestArgsN(t *testing.T) {
	for in, want := range map[int]int{0: 1, -3: 1, 1: 1, 7: 7} {
		if got := (command.Args{Count: in}).N(); got != want {
			t.Errorf("N(%d) = %d, quería %d", in, got, want)
		}
	}
}

func TestMustRegisterPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("MustRegister con un nombre inválido debía entrar en pánico")
		}
	}()
	command.NewRegistry[*state]().MustRegister(cmd("mal"))
}

func TestLabelFallsBackToDescription(t *testing.T) {
	c := cmd("file.save")
	if c.Label() != "descripción de file.save" {
		t.Errorf("Label = %q", c.Label())
	}
	c.Short = "guardar"
	if c.Label() != "guardar" {
		t.Errorf("Label = %q", c.Label())
	}
}
