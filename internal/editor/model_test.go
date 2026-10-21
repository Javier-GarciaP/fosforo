package editor_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/exp/teatest/v2"

	"github.com/Javier-GarciaP/fosforo/internal/editor"
)

const wait = 3 * time.Second

func contains(s string) func([]byte) bool {
	return func(b []byte) bool { return bytes.Contains(b, []byte(s)) }
}

func TestQuitKeys(t *testing.T) {
	for _, tc := range []struct {
		name string
		send tea.Msg
	}{
		{"q", tea.KeyPressMsg{Code: 'q', Text: "q"}},
		{"ctrl+c", tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tm := teatest.NewTestModel(t, editor.New(""), teatest.WithInitialTermSize(80, 24))
			teatest.WaitFor(t, tm.Output(), contains("Pulsa q para salir"), teatest.WithDuration(wait))
			tm.Send(tc.send)
			tm.WaitFinished(t, teatest.WithFinalTimeout(wait))
		})
	}
}

func TestOtherKeysDoNotQuit(t *testing.T) {
	tm := teatest.NewTestModel(t, editor.New(""), teatest.WithInitialTermSize(80, 24))
	teatest.WaitFor(t, tm.Output(), contains("Pulsa q para salir"), teatest.WithDuration(wait))
	tm.Send(tea.KeyPressMsg{Code: 'x', Text: "x"})
	tm.Send(tea.KeyPressMsg{Code: 'q', Text: "q"})
	tm.WaitFinished(t, teatest.WithFinalTimeout(wait))
}

func TestShowsPath(t *testing.T) {
	tm := teatest.NewTestModel(t, editor.New("README.md"), teatest.WithInitialTermSize(80, 24))
	teatest.WaitFor(t, tm.Output(), contains("README.md"), teatest.WithDuration(wait))
	_ = tm.Quit()
}

func TestMinimumSize(t *testing.T) {
	tm := teatest.NewTestModel(t, editor.New(""), teatest.WithInitialTermSize(30, 8))
	teatest.WaitFor(t, tm.Output(), contains("demasiado pequeña"), teatest.WithDuration(wait))

	tm.Send(tea.WindowSizeMsg{Width: 80, Height: 24})
	teatest.WaitFor(t, tm.Output(), contains("Pulsa q para salir"), teatest.WithDuration(wait))

	tm.Send(tea.KeyPressMsg{Code: 'q', Text: "q"})
	tm.WaitFinished(t, teatest.WithFinalTimeout(wait))
}

func TestTooSmallBoundaries(t *testing.T) {
	for _, tc := range []struct {
		w, h int
		want bool
	}{
		{0, 0, false}, // tamaño aún desconocido
		{39, 24, true},
		{80, 9, true},
		{40, 10, false},
		{30, 8, true},
		{200, 60, false},
	} {
		m, _ := editor.New("").Update(tea.WindowSizeMsg{Width: tc.w, Height: tc.h})
		if got := m.(editor.Model).TooSmall(); got != tc.want {
			t.Errorf("%dx%d: TooSmall() = %v, quería %v", tc.w, tc.h, got, tc.want)
		}
	}
}

func TestViewUsesAltScreen(t *testing.T) {
	if v := editor.New("").View(); !v.AltScreen || !strings.Contains(v.Content, "Pulsa q") {
		t.Fatalf("vista inesperada: alt=%v contenido=%q", v.AltScreen, v.Content)
	}
}
