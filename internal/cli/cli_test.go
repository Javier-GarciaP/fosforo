package cli_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/Javier-GarciaP/fosforo/internal/cli"
)

func run(t *testing.T, ui func(cli.Options) error, args ...string) (code int, stdout, stderr string) {
	t.Helper()
	var out, errb bytes.Buffer
	code = cli.Run(args, &out, &errb, ui)
	return code, out.String(), errb.String()
}

func neverUI(t *testing.T) func(cli.Options) error {
	return func(cli.Options) error { t.Fatal("no debía abrirse la interfaz"); return nil }
}

func TestRun(t *testing.T) {
	var gotPath, gotIcons string
	record := func(o cli.Options) error { gotPath, gotIcons = o.Path, o.Icons; return nil }

	tests := []struct {
		name       string
		args       []string
		ui         func(cli.Options) error
		wantCode   int
		wantStdout string
		wantStderr string
		wantPath   string
	}{
		{"version", []string{"--version"}, neverUI(t), 0, "fos ", "", ""},
		{"version con un guion", []string{"-version"}, neverUI(t), 0, "fos ", "", ""},
		{"ayuda", []string{"--help"}, neverUI(t), 0, "Uso: fos", "", ""},
		{"ayuda corta", []string{"-h"}, neverUI(t), 0, "Uso: fos", "", ""},
		{"flag desconocido", []string{"--no-existe"}, neverUI(t), 2, "", "no-existe", ""},
		{"demasiadas rutas", []string{"a", "b"}, neverUI(t), 2, "", "demasiados argumentos", ""},
		{"sin ruta", nil, record, 0, "", "", ""},
		{"con ruta", []string{"README.md"}, record, 0, "", "", "README.md"},
		{"error de la interfaz", nil, func(cli.Options) error { return errors.New("boom") }, 1, "", "boom", ""},
		{"interrumpido", nil, func(cli.Options) error { return tea.ErrInterrupted }, 130, "", "", ""},
		{"iconos none", []string{"-icons=none", "x"}, record, 0, "", "", "x"},
		{"iconos inválidos", []string{"-icons=raros"}, neverUI(t), 2, "", "icons", ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotPath, gotIcons = "", ""
			code, stdout, stderr := run(t, tc.ui, tc.args...)
			if code != tc.wantCode {
				t.Errorf("código = %d, quería %d (stderr=%q)", code, tc.wantCode, stderr)
			}
			if !strings.Contains(stdout, tc.wantStdout) {
				t.Errorf("stdout = %q, debía contener %q", stdout, tc.wantStdout)
			}
			if !strings.Contains(stderr, tc.wantStderr) {
				t.Errorf("stderr = %q, debía contener %q", stderr, tc.wantStderr)
			}
			if gotPath != tc.wantPath {
				t.Errorf("ruta = %q, quería %q", gotPath, tc.wantPath)
			}
			if tc.wantCode == 0 && tc.ui != nil && gotPath != "" || tc.name == "sin ruta" {
				want := "nerd"
				if tc.name == "iconos none" {
					want = "none"
				}
				if gotIcons != want {
					t.Errorf("iconos = %q, quería %q", gotIcons, want)
				}
			}
		})
	}
}

func TestVersionIsOneLine(t *testing.T) {
	_, stdout, _ := run(t, neverUI(t), "--version")
	if strings.Count(strings.TrimRight(stdout, "\n"), "\n") != 0 {
		t.Errorf("la versión debe ser una sola línea: %q", stdout)
	}
}

func TestIconsFlagTakesPrecedenceOnlyWhenGiven(t *testing.T) {
	o, err := cli.Parse([]string{"x"})
	if err != nil || o.IconsSet {
		t.Fatalf("sin --icons no debe pisar la configuración: %+v %v", o, err)
	}
	o, err = cli.Parse([]string{"--icons=nerd", "x"})
	if err != nil || !o.IconsSet || o.Icons != "nerd" {
		t.Fatalf("--icons indicado: %+v %v", o, err)
	}
}
