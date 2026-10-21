package editor_test

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/exp/teatest/v2"

	"github.com/Javier-GarciaP/fosforo/internal/buffer"
	"github.com/Javier-GarciaP/fosforo/internal/editor"
)

// goTo coloca el cursor en (línea, columna) usando solo comandos del editor: «:N», «0» y «l».
func (e *ed) goTo(line, col int) *ed {
	e.ex(strconv.Itoa(line + 1)).typ("0")
	for i := 0; i < col; i++ {
		e.typ("l")
	}
	e.t.Helper()
	e.at(line, col)
	return e
}

// screen devuelve lo que se dibuja, sin colores.
func (e *ed) screen() string { return ansi.Strip(e.m.View().Content) }

func (e *ed) mode(want editor.Mode) {
	e.t.Helper()
	if got := e.m.Mode(); got != want {
		e.t.Errorf("modo = %v, quería %v", got, want)
	}
}

func (e *ed) wantText(want string) {
	e.t.Helper()
	if got := e.text(); got != want {
		e.t.Errorf("texto = %q, quería %q", got, want)
	}
}

// ---------- modos ----------

func TestStartsInNormalMode(t *testing.T) {
	e := newEd(t, "abc\n")
	e.mode(editor.ModeNormal)
	if !strings.Contains(e.screen(), "NORMAL") {
		t.Errorf("la statusline debe decir NORMAL:\n%s", e.screen())
	}
}

func TestInsertAndEscape(t *testing.T) {
	e := newEd(t, "abc\n").goTo(0, 2).typ("i")
	e.mode(editor.ModeInsert)
	e.typ("X")
	e.wantText("abXc\n")
	e.send(keyEsc)
	e.mode(editor.ModeNormal) // Esc surte efecto en el mismo ciclo de actualización
	e.at(0, 2)                // y el cursor retrocede una posición
	e.typ("i").send(keyEsc).goTo(0, 0).typ("i").send(keyEsc)
	e.at(0, 0) // en la columna 0 no retrocede
}

func TestEnterInsertKeys(t *testing.T) {
	for _, tc := range []struct {
		name, content, keys string
		line, col           int
		want                string
	}{
		{"i", "  abc\n", "$i", 0, 4, "  abc\n"},
		{"a", "abc\n", "a", 0, 1, "abc\n"},
		{"a al final", "abc\n", "$a", 0, 3, "abc\n"},
		{"I va al primer no blanco", "   abc\n", "$I", 0, 3, "   abc\n"},
		{"A va al final", "abc\n", "A", 0, 3, "abc\n"},
		{"o abre debajo con sangría", "  abc\nxyz\n", "o", 1, 2, "  abc\n  \nxyz\n"},
		{"O abre encima con sangría", "x\n  abc\n", "jO", 1, 2, "x\n  \n  abc\n"},
		{"o en la última línea", "abc", "o", 1, 0, "abc\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e := newEd(t, tc.content).typ(tc.keys)
			e.mode(editor.ModeInsert)
			e.at(tc.line, tc.col)
			if !strings.Contains(e.screen(), "-- INSERT --") {
				t.Error("la línea de comandos debe mostrar -- INSERT --")
			}
			if tc.name == "o en la última línea" {
				if e.m.Buffer().LineCount() != 2 {
					t.Errorf("líneas = %d", e.m.Buffer().LineCount())
				}
				return
			}
			e.wantText(tc.want)
		})
	}
}

func TestUnmappedKeysAreIgnoredInNormal(t *testing.T) {
	e := newEd(t, "abc\n")
	for _, k := range []tea.Msg{
		tea.KeyPressMsg{Code: 'z', Text: "z"}, tea.KeyPressMsg{Code: 'q', Text: "q"}, tea.KeyPressMsg{Code: 'Z', Text: "Z"},
		tea.KeyPressMsg{Code: 'K', Text: "K"}, tea.KeyPressMsg{Code: '!', Text: "!"}, keyEnter, keyTab, keyBackspace,
		tea.KeyPressMsg{Code: 'q', Mod: tea.ModCtrl}, tea.KeyPressMsg{Code: 'ñ', Text: "ñ"},
	} {
		e.send(k)
		if e.m.Buffer().Modified() || e.text() != "abc\n" {
			t.Fatalf("la tecla %v modificó el texto: %q", k, e.text())
		}
		e.mode(editor.ModeNormal)
		e.at(0, 0)
		if e.m.Pending() != "" {
			t.Fatalf("la tecla %v dejó teclas pendientes: %q", k, e.m.Pending())
		}
	}
}

func TestUnmappedKeysAreIgnoredInVisual(t *testing.T) {
	e := newEd(t, "abc\n").typ("v").typ("zq!")
	e.mode(editor.ModeVisual)
	e.wantText("abc\n")
}

func TestCtrlSSavesInEveryMode(t *testing.T) {
	for _, tc := range []struct {
		name string
		keys string
		mode editor.Mode
	}{
		{"normal", "", editor.ModeNormal},
		{"insert", "i", editor.ModeInsert},
		{"visual", "v", editor.ModeVisual},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := filepath.Join(t.TempDir(), "f.txt")
			os.WriteFile(p, []byte("abc\n"), 0o644)
			m, _ := editor.Open(p)
			e := &ed{t: t, m: m}
			e.send(tea.WindowSizeMsg{Width: 80, Height: 24}).typ("ix").send(keyEsc).typ(tc.keys).send(keyCtrlS)
			if got, _ := os.ReadFile(p); string(got) != "xabc\n" {
				t.Errorf("archivo = %q", got)
			}
			e.mode(tc.mode) // guardar no cambia de modo
			if e.m.Buffer().Modified() {
				t.Error("tras guardar no debe estar modificado")
			}
		})
	}
}

func TestCtrlCCancelsLikeEscape(t *testing.T) {
	e := newEd(t, "abc\n").typ("i").send(keyCtrlC)
	e.mode(editor.ModeNormal)
	e.typ("v").send(keyCtrlC)
	e.mode(editor.ModeNormal)
	e.typ("3d").send(keyCtrlC)
	if e.m.Pending() != "" {
		t.Errorf("pendiente = %q", e.m.Pending())
	}
}

// ---------- secuencias, conteos y temporizador ----------

func TestSequenceGG(t *testing.T) {
	e := newEd(t, strings.Repeat("x\n", 30)).typ("G")
	e.at(29, 0)
	e.typ("g")
	rows := strings.Split(e.screen(), "\n")
	if e.m.Pending() != "g" || strings.TrimSpace(rows[len(rows)-1]) != "g" {
		t.Errorf("las teclas pendientes se muestran en la línea de comandos: %q / %q", e.m.Pending(), rows[len(rows)-1])
	}
	e.typ("g")
	e.at(0, 0)
	if e.m.Pending() != "" {
		t.Errorf("tras completar no queda nada pendiente: %q", e.m.Pending())
	}
}

func TestIncompleteSequenceTimesOut(t *testing.T) {
	e := newEd(t, strings.Repeat("x\n", 30)).typ("G").typ("g")
	if e.m.Pending() != "g" {
		t.Fatal("debe quedar «g» pendiente")
	}
	// un temporizador de una secuencia anterior no debe descartar la actual
	e.send(editor.TimeoutMsg{Seq: e.m.Seq() - 1})
	if e.m.Pending() != "g" {
		t.Fatal("un temporizador viejo debe ignorarse")
	}
	e.send(editor.TimeoutMsg{Seq: e.m.Seq()}) // se cumple el segundo sin otra tecla
	if e.m.Pending() != "" {
		t.Fatalf("la secuencia pendiente se descarta: %q", e.m.Pending())
	}
	e.at(29, 0)
	e.wantText(strings.Repeat("x\n", 30))
	e.typ("g")
	e.at(29, 0) // una «g» suelta no completa nada
	e.typ("g")
	e.at(0, 0)
}

func TestTimeoutWithRealRuntime(t *testing.T) {
	m, _ := editor.Open("", editor.WithTimeout(40*time.Millisecond))
	m.Buffer().Insert(buffer.Pos{}, strings.Repeat("x\n", 30))
	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(80, 24))
	tm.Type("G")
	tm.Type("g")
	time.Sleep(300 * time.Millisecond) // vence el temporizador de Bubble Tea
	tm.Type("g")                       // si «g» siguiera pendiente, esto saltaría a la línea 1
	tm.Send(keyEsc)
	_ = tm.Quit()
	fm := tm.FinalModel(t, teatest.WithFinalTimeout(wait)).(editor.Model)
	if fm.Cursor().Line == 0 {
		t.Error("tras el tiempo de espera, la primera «g» debía haberse descartado")
	}
}

func TestCounts(t *testing.T) {
	ten := strings.Repeat("l\n", 40)
	for _, tc := range []struct {
		name, content, keys string
		want                string
		line                int
	}{
		{"3dd", "1\n2\n3\n4\n5\n", "3dd", "4\n5\n", 0},
		{"5j", ten, "5j", ten, 5},
		{"conteo de dos cifras", ten, "12j", ten, 12},
		{"5dd con menos líneas", "1\n2\n3\n", "j5dd", "1\n", 0},
		{"d2w", "a b c d\n", "d2w", "c d\n", 0},
		{"2d3w multiplica", "a b c d e f g h\n", "2d3w", "g h\n", 0},
		{"3x", "abcdef\n", "3x", "def\n", 0},
		{"2dd + u", "1\n2\n3\n", "2ddu", "1\n2\n3\n", 0},
		{"0 sin conteo es una motion", "  abc\n", "$0", "  abc\n", 0},
		{"10 usa el 0 como cifra", ten, "10j", ten, 10},
		{"3G", ten, "3G", ten, 2},
		{"dd con conteo en medio", "1\n2\n3\n4\n", "d2d", "3\n4\n", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e := newEd(t, tc.content).typ(tc.keys)
			e.wantText(tc.want)
			if e.m.Cursor().Line != tc.line {
				t.Errorf("línea = %d, quería %d", e.m.Cursor().Line, tc.line)
			}
		})
	}
}

func TestPendingDisplayAndCancel(t *testing.T) {
	e := newEd(t, strings.Repeat("x\n", 10)).typ("3d")
	if e.m.Pending() != "3d" {
		t.Errorf("pendiente = %q", e.m.Pending())
	}
	e.send(keyEsc)
	if e.m.Pending() != "" {
		t.Fatalf("Esc cancela lo pendiente: %q", e.m.Pending())
	}
	e.typ("j")
	e.at(1, 0) // el conteo se descartó: solo baja una línea
	e.typ("3z").typ("j")
	e.at(2, 0) // una tecla sin función también descarta el conteo
}

// ---------- motions ----------

func TestMotions(t *testing.T) {
	twenty := strings.Repeat("  línea\n", 20)
	for _, tc := range []struct {
		name, content  string
		startL, startC int
		keys           string
		wantL, wantC   int
	}{
		{"l", "abc\n", 0, 0, "l", 0, 1},
		{"l se detiene en el último carácter", "abc\n", 0, 2, "l", 0, 2},
		{"h", "abc\n", 0, 2, "h", 0, 1},
		{"h se detiene en 0", "abc\n", 0, 0, "h", 0, 0},
		{"2l", "abcdef\n", 0, 0, "2l", 0, 2},
		{"j y k", "ab\ncd\nef\n", 0, 1, "jj", 2, 1},
		{"k", "ab\ncd\nef\n", 2, 1, "k", 1, 1},
		{"w", "foo bar  baz\n", 0, 0, "w", 0, 4},
		{"ww", "foo bar  baz\n", 0, 0, "ww", 0, 9},
		{"w en la última palabra", "foo bar  baz\n", 0, 9, "w", 0, 11},
		{"w cruza de línea", "foo\nbar\n", 0, 0, "w", 1, 0},
		{"w con puntuación", "foo.bar\n", 0, 0, "w", 0, 3},
		{"w puntuación y palabra", "foo.bar\n", 0, 0, "ww", 0, 4},
		{"w se detiene en línea vacía", "a\n\nb\n", 0, 0, "w", 1, 0},
		{"w tras línea vacía", "a\n\nb\n", 0, 0, "ww", 2, 0},
		{"3w", "a b c d e\n", 0, 0, "3w", 0, 6},
		{"b", "foo bar\n", 0, 5, "b", 0, 4},
		{"bb", "foo bar\n", 0, 5, "bb", 0, 0},
		{"b cruza de línea", "foo\nbar\n", 1, 0, "b", 0, 0},
		{"b en el inicio", "foo\n", 0, 0, "b", 0, 0},
		{"e", "foo bar\n", 0, 0, "e", 0, 2},
		{"ee", "foo bar\n", 0, 0, "ee", 0, 6},
		{"e cruza de línea", "foo\nbar\n", 0, 2, "e", 1, 2},
		{"0", "  abc\n", 0, 4, "0", 0, 0},
		{"^", "   x yz\n", 0, 6, "^", 0, 3},
		{"$", "   x yz\n", 0, 0, "$", 0, 6},
		{"$ en línea vacía", "\nabc\n", 0, 0, "$", 0, 0},
		{"gg", "a\nb\nc\n", 2, 0, "gg", 0, 0},
		{"gg va al primer no blanco", "  a\nb\n", 1, 0, "gg", 0, 2},
		{"G", "a\nb\nc\n", 0, 0, "G", 2, 0},
		{"12G", twenty, 0, 0, "12G", 11, 2},
		{"G con conteo fuera de rango", "a\nb\nc\n", 0, 0, "99G", 2, 0},
		{"5gg", twenty, 0, 0, "5gg", 4, 2},
		{"Unicode: w", "日本語 テキスト\n", 0, 0, "w", 0, 4},
		{"Unicode: $", "日本語 テキスト\n", 0, 0, "$", 0, 7},
		{"Unicode: l cuenta grafemas", "a👨‍👩‍👧b\n", 0, 0, "ll", 0, 2},
		{"Unicode: e con acentos", "señal ñandú\n", 0, 0, "e", 0, 4},
		{"Ctrl+d", twenty, 0, 0, "\x04", 11, 0},
		{"Ctrl+d y Ctrl+u", twenty, 0, 0, "\x04\x15", 0, 0},
		{"Ctrl+d se detiene al final", "a\nb\n", 0, 0, "\x04", 1, 0},
		{"flechas en Normal", "abc\ndef\n", 0, 0, "\x1b[C\x1b[B", 1, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e := newEd(t, tc.content).goTo(tc.startL, tc.startC)
			for _, k := range tokens(tc.keys) {
				e.send(k)
			}
			e.at(tc.wantL, tc.wantC)
			e.mode(editor.ModeNormal)
		})
	}
}

// tokens convierte una cadena de teclas de prueba en mensajes: caracteres normales, Ctrl+D (\x04), Ctrl+U (\x15)
// y las flechas \x1b[C (derecha) y \x1b[B (abajo).
func tokens(s string) []tea.Msg {
	var out []tea.Msg
	for i := 0; i < len(s); {
		switch {
		case strings.HasPrefix(s[i:], "\x1b[C"):
			out, i = append(out, keyRight), i+3
		case strings.HasPrefix(s[i:], "\x1b[B"):
			out, i = append(out, keyDown), i+3
		case s[i] == 4:
			out, i = append(out, keyCtrlD), i+1
		case s[i] == 21:
			out, i = append(out, keyCtrlU), i+1
		default:
			r := []rune(s[i:])[0]
			out, i = append(out, tea.KeyPressMsg{Code: r, Text: string(r)}), i+len(string(r))
		}
	}
	return out
}

// ---------- operadores ----------

func TestOperators(t *testing.T) {
	for _, tc := range []struct {
		name, content  string
		startL, startC int
		keys           string
		want           string
		wantL, wantC   int
		wantMode       editor.Mode
	}{
		{"dw borra la palabra y su espacio", "foo bar baz\n", 0, 0, "dw", "bar baz\n", 0, 0, editor.ModeNormal},
		{"dw en la última palabra no une líneas", "foo bar\nnext\n", 0, 4, "dw", "foo \nnext\n", 0, 3, editor.ModeNormal},
		{"dw al final del archivo", "foo bar\n", 0, 4, "dw", "foo \n", 0, 3, editor.ModeNormal},
		{"db", "foo bar\n", 0, 4, "db", "bar\n", 0, 0, editor.ModeNormal},
		{"de", "foo bar\n", 0, 0, "de", " bar\n", 0, 0, editor.ModeNormal},
		{"d$", "foo bar\n", 0, 4, "d$", "foo \n", 0, 3, editor.ModeNormal},
		{"D", "foo bar\n", 0, 4, "D", "foo \n", 0, 3, editor.ModeNormal},
		{"d0", "foo bar\n", 0, 4, "d0", "bar\n", 0, 0, editor.ModeNormal},
		{"dd", "a\nb\nc\n", 1, 0, "dd", "a\nc\n", 1, 0, editor.ModeNormal},
		{"dd en la última línea sube", "a\nb\nc\n", 2, 0, "dd", "a\nb\n", 1, 0, editor.ModeNormal},
		{"dd en la única línea la vacía", "abc\n", 0, 0, "dd", "\n", 0, 0, editor.ModeNormal},
		{"dj", "a\nb\nc\nd\n", 1, 0, "dj", "a\nd\n", 1, 0, editor.ModeNormal},
		{"dk", "a\nb\nc\nd\n", 2, 0, "dk", "a\nd\n", 1, 0, editor.ModeNormal},
		{"dG", "a\nb\nc\n", 1, 0, "dG", "a\n", 0, 0, editor.ModeNormal},
		{"dgg", "a\nb\nc\n", 1, 0, "dgg", "c\n", 0, 0, editor.ModeNormal},
		{"x", "abc\n", 0, 0, "x", "bc\n", 0, 0, editor.ModeNormal},
		{"x en el último carácter", "abc\n", 0, 2, "x", "ab\n", 0, 1, editor.ModeNormal},
		{"x en línea vacía", "\nabc\n", 0, 0, "x", "\nabc\n", 0, 0, editor.ModeNormal},
		{"x con Unicode borra el grafema", "a👨‍👩‍👧b\n", 0, 1, "x", "ab\n", 0, 1, editor.ModeNormal},
		{"cw cambia la palabra", "foo bar\n", 0, 0, "cwX\x1b", "X bar\n", 0, 0, editor.ModeNormal},
		{"cw en el último carácter de la palabra", "foo bar\n", 0, 2, "cwX\x1b", "foX bar\n", 0, 2, editor.ModeNormal},
		{"cc vacía la línea y entra en Insert", "  foo\nbar\n", 0, 3, "cc", "\nbar\n", 0, 0, editor.ModeInsert},
		{"cc y escribir", "foo\nbar\n", 0, 0, "ccnuevo\x1b", "nuevo\nbar\n", 0, 4, editor.ModeNormal},
		{"C", "foo bar\n", 0, 4, "CX\x1b", "foo X\n", 0, 4, editor.ModeNormal},
		{"cj cambia dos líneas por una", "a\nb\nc\n", 0, 0, "cjX\x1b", "X\nc\n", 0, 0, editor.ModeNormal},
		{"yy no modifica", "a\nb\n", 0, 0, "yy", "a\nb\n", 0, 0, editor.ModeNormal},
		{"yy p duplica la línea debajo", "a\nb\n", 0, 0, "yyp", "a\na\nb\n", 1, 0, editor.ModeNormal},
		{"yy P duplica encima", "a\nb\n", 1, 0, "yyP", "a\nb\nb\n", 1, 0, editor.ModeNormal},
		{"3yy p", "a\nb\nc\nd\n", 0, 0, "3yyp", "a\na\nb\nc\nb\nc\nd\n", 1, 0, editor.ModeNormal},
		{"yw P", "foo bar\n", 0, 0, "ywP", "foo foo bar\n", 0, 3, editor.ModeNormal},
		{"yw al final y P", "foo bar\n", 0, 0, "yw$P", "foo bafoo r\n", 0, 9, editor.ModeNormal},
		{"yl p", "abc\n", 0, 0, "yl$p", "abca\n", 0, 3, editor.ModeNormal},
		{"dd p mueve la línea", "a\nb\nc\n", 0, 0, "ddp", "b\na\nc\n", 1, 0, editor.ModeNormal},
		{"dd P", "a\nb\nc\n", 1, 0, "ddP", "a\nb\nc\n", 1, 0, editor.ModeNormal},
		{"xp intercambia", "ab\n", 0, 0, "xp", "ba\n", 0, 1, editor.ModeNormal},
		{"3p", "a\n", 0, 0, "yy3p", "a\na\na\na\n", 1, 0, editor.ModeNormal},
		{"p con registro vacío", "a\n", 0, 0, "p", "a\n", 0, 0, editor.ModeNormal},
		{"dd en Unicode", "日本\nテ\n", 0, 0, "dd", "テ\n", 0, 0, editor.ModeNormal},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e := newEd(t, tc.content).goTo(tc.startL, tc.startC)
			for _, k := range tokensEsc(tc.keys) {
				e.send(k)
			}
			e.wantText(tc.want)
			e.at(tc.wantL, tc.wantC)
			e.mode(tc.wantMode)
		})
	}
}

// tokensEsc es como tokens pero interpreta \x1b suelto como la tecla Esc.
func tokensEsc(s string) []tea.Msg {
	var out []tea.Msg
	for _, part := range strings.SplitAfter(s, "\x1b") {
		if strings.HasSuffix(part, "\x1b") {
			out = append(out, tokens(strings.TrimSuffix(part, "\x1b"))...)
			out = append(out, keyEsc)
			continue
		}
		out = append(out, tokens(part)...)
	}
	return out
}

func TestOperatorsOnReadOnlyBufferWarn(t *testing.T) {
	p := filepath.Join(t.TempDir(), "bin")
	os.WriteFile(p, []byte("ab\x00cd\nx\n"), 0o644)
	m, _ := editor.Open(p)
	e := &ed{t: t, m: m}
	e.send(tea.WindowSizeMsg{Width: 80, Height: 24})
	for _, keys := range []string{"dd", "x", "dw", "D", "cc", "ofoo", "p"} {
		e.send(keyEsc).typ(keys)
		if e.m.Buffer().Modified() || e.text() != "ab\x00cd\nx\n" {
			t.Fatalf("%q modificó un buffer de solo lectura: %q", keys, e.text())
		}
	}
	e.send(keyEsc).typ("dd")
	if !strings.Contains(e.m.Message(), "Solo lectura") {
		t.Errorf("mensaje = %q", e.m.Message())
	}
}

func TestOperatorPendingCancelledByUnmappedKey(t *testing.T) {
	e := newEd(t, "abc def\n").typ("dz")
	e.wantText("abc def\n")
	if e.m.Pending() != "" {
		t.Errorf("pendiente = %q", e.m.Pending())
	}
	e.typ("w")
	e.at(0, 4) // «w» vuelve a ser un movimiento normal
	e.wantText("abc def\n")
}

func TestDeleteToFirstLineOnSingleLine(t *testing.T) {
	e := newEd(t, "foo bar\n").typ("dgg")
	e.wantText("\n") // dgg: linewise sobre la única línea
}

// ---------- registro y portapapeles ----------

func TestYankCopiesToClipboard(t *testing.T) {
	var copied []string
	clip := editor.WithClipboard(func(s string) tea.Cmd { copied = append(copied, s); return nil })
	e := newEd(t, "uno dos\ntres\n", clip).typ("yy")
	e.typ("jyw")
	e.typ("ggvly")
	e.typ("ggddjdd") // borrar no copia al portapapeles
	if len(copied) != 3 || copied[0] != "uno dos\n" || copied[1] != "tres" || copied[2] != "un" {
		t.Errorf("portapapeles = %q", copied)
	}
}

func TestDefaultClipboardReturnsCommand(t *testing.T) {
	m, _ := editor.Open("")
	m.Buffer().Insert(buffer.Pos{}, "hola")
	nm, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	var cmd tea.Cmd
	for _, k := range tokens("yy") {
		nm, cmd = nm.Update(k)
	}
	if cmd == nil {
		t.Error("yy con el portapapeles por defecto debe devolver el comando OSC 52")
	}
}

// ---------- visual ----------

func TestVisualLineDelete(t *testing.T) {
	e := newEd(t, "1\n2\n3\n4\n5\n").typ("Vjj")
	e.mode(editor.ModeVisualLine)
	if !strings.Contains(e.screen(), "3 líneas") || !strings.Contains(e.screen(), "VISUAL LINE") {
		t.Errorf("la statusline debe mostrar el tamaño de la selección:\n%s", e.screen())
	}
	e.typ("d")
	e.mode(editor.ModeNormal)
	e.wantText("4\n5\n")
	e.at(0, 0)
}

func TestVisualCharOperations(t *testing.T) {
	for _, tc := range []struct {
		name, keys, want string
		wantC            int
		mode             editor.Mode
	}{
		{"d", "vlld", "def\n", 0, editor.ModeNormal},
		{"d hasta el final", "v$d", "\n", 0, editor.ModeNormal},
		{"x", "vlx", "cdef\n", 0, editor.ModeNormal},
		{"c", "vlcX\x1b", "Xcdef\n", 0, editor.ModeNormal},
		{"y y p", "vlly$p", "abcdefabc\n", 8, editor.ModeNormal},
		{"selección hacia atrás", "$vhhd", "abc\n", 2, editor.ModeNormal},
		{"Esc sale sin cambios", "vll\x1b", "abcdef\n", 2, editor.ModeNormal},
		{"v dos veces sale", "vlv", "abcdef\n", 1, editor.ModeNormal},
		{"v y V cambian de tipo", "vV", "abcdef\n", 0, editor.ModeVisualLine},
		{"V y v cambian de tipo", "Vv", "abcdef\n", 0, editor.ModeVisual},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e := newEd(t, "abcdef\n")
			for _, k := range tokensEsc(tc.keys) {
				e.send(k)
			}
			e.mode(tc.mode)
			e.wantText(tc.want)
		})
	}
}

func TestVisualSwapEnd(t *testing.T) {
	e := newEd(t, "abcdef\n").typ("vll")
	e.at(0, 2)
	e.typ("o")
	e.at(0, 0) // el cursor pasa al otro extremo
	e.typ("hh")
	e.at(0, 0)
	e.typ("d") // la selección sigue siendo de 0 a 2
	e.wantText("def\n")
}

func TestVisualMultilineChar(t *testing.T) {
	e := newEd(t, "abc\ndef\nghi\n").goTo(0, 1).typ("vjd")
	e.wantText("af\nghi\n")
	e.at(0, 1)
}

func TestVisualInfoCharCount(t *testing.T) {
	e := newEd(t, "abcdef\n").typ("vll")
	if !strings.Contains(e.screen(), "3 car.") {
		t.Errorf("debe indicar los caracteres seleccionados:\n%s", e.screen())
	}
}

// ---------- deshacer y rehacer ----------

func TestUndoInsertSessionIsOneStep(t *testing.T) {
	e := newEd(t, "").typ("ihola").send(keyEnter).typ("mundo").send(keyEsc)
	e.wantText("hola\nmundo\n")
	e.typ("u")
	e.wantText("") // todo lo escrito en la sesión desaparece de una vez
	if e.m.Buffer().Modified() {
		t.Error("deshacer todo deja el buffer sin modificar")
	}
	e.send(keyCtrlR)
	e.wantText("hola\nmundo\n")
}

func TestUndoRedoSequence(t *testing.T) {
	e := newEd(t, "").typ("ia").send(keyEsc).typ("ob").send(keyEsc).typ("oc").send(keyEsc)
	e.wantText("a\nb\nc\n")
	e.typ("u")
	e.wantText("a\nb\n")
	e.typ("u")
	e.wantText("a\n")
	e.send(keyCtrlR)
	e.wantText("a\nb\n")
	e.send(keyCtrlR, keyCtrlR)
	e.wantText("a\nb\nc\n")
	if !strings.Contains(e.m.Message(), "más reciente") {
		t.Errorf("mensaje = %q", e.m.Message())
	}
}

func TestUndoRestoresCursor(t *testing.T) {
	e := newEd(t, "a\nb\nc\nd\n").goTo(2, 0).typ("dd")
	e.wantText("a\nb\nd\n")
	e.typ("gg").typ("u")
	e.wantText("a\nb\nc\nd\n")
	e.at(2, 0)
}

func TestUndoCommands(t *testing.T) {
	for _, tc := range []struct {
		name, content, keys string
	}{
		{"dd", "a\nb\nc\n", "ddu"},
		{"3dd", "a\nb\nc\nd\n", "3ddu"},
		{"x", "abc\n", "xu"},
		{"cw", "foo bar\n", "cwxyz\x1bu"},
		{"p", "a\nb\n", "yypu"},
		{"D", "foo bar\n", "wDu"},
		{"o", "a\n", "ofoo\x1bu"},
		{"O", "a\n", "Ofoo\x1bu"},
		{"visual d", "abc\ndef\n", "vjdu"},
		{"visual c", "abc\ndef\n", "vjcZ\x1bu"},
		{"dw", "foo bar\n", "dwu"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e := newEd(t, tc.content)
			for _, k := range tokensEsc(tc.keys) {
				e.send(k)
			}
			e.wantText(tc.content)
			if e.m.Buffer().Modified() {
				t.Error("tras deshacer queda sin modificar")
			}
		})
	}
}

func TestUndoNothingMessage(t *testing.T) {
	e := newEd(t, "abc\n").typ("u")
	if !strings.Contains(e.m.Message(), "más antiguo") {
		t.Errorf("mensaje = %q", e.m.Message())
	}
	e.wantText("abc\n")
}

func TestNewEditAfterUndoDiscardsRedo(t *testing.T) {
	e := newEd(t, "abc\n").typ("x").typ("u").typ("x").send(keyCtrlR)
	e.wantText("bc\n")
	if !strings.Contains(e.m.Message(), "más reciente") {
		t.Errorf("mensaje = %q", e.m.Message())
	}
}

// ---------- búsqueda ----------

const trieSrc = "package keymap\n\ntype Node struct {\n\tchildren map[rune]*Node\n}\n\n// Lookup en el trie\nfunc (n *Node) Lookup() {\n\ttrie := 1\n}\n"

func TestSearchAndRepeat(t *testing.T) {
	e := newEd(t, trieSrc).typ("/trie").send(keyEnter)
	e.mode(editor.ModeNormal)
	e.at(6, 16) // primera coincidencia después del cursor: «trie» en el comentario
	e.typ("n")
	e.at(8, 1) // la siguiente
	e.typ("n")
	e.at(6, 16) // da la vuelta al archivo
	e.typ("N")
	e.at(8, 1)
	e.wantText(trieSrc)
}

func TestSearchBackwardWithN(t *testing.T) {
	e := newEd(t, "ab ab ab\n").goTo(0, 6).typ("/ab").send(keyEnter) // lo siguiente da la vuelta
	e.at(0, 0)
	e.typ("N")
	e.at(0, 6)
}

func TestSearchNotFound(t *testing.T) {
	e := newEd(t, "abc\ndef\n").goTo(1, 1).typ("/zzz").send(keyEnter)
	if got := e.m.Message(); got != "E486: patrón no encontrado: zzz" {
		t.Errorf("mensaje = %q", got)
	}
	e.at(1, 1) // el cursor no se mueve
	e.wantText("abc\ndef\n")
}

func TestSearchInvalidPatternAndNoPrevious(t *testing.T) {
	e := newEd(t, "abc\n").typ("n")
	if !strings.HasPrefix(e.m.Message(), "E35") {
		t.Errorf("n sin búsqueda previa: %q", e.m.Message())
	}
	e.typ("/(").send(keyEnter)
	if !strings.HasPrefix(e.m.Message(), "E383") {
		t.Errorf("patrón inválido: %q", e.m.Message())
	}
	e.at(0, 0)
}

func TestSearchRepeatsLastPatternWithEmptyEnter(t *testing.T) {
	e := newEd(t, "x a x a\n").typ("/a").send(keyEnter)
	e.at(0, 2)
	e.typ("/").send(keyEnter)
	e.at(0, 6)
}

func TestSearchUnicodeColumns(t *testing.T) {
	e := newEd(t, "日本語 日本\n").typ("/本").send(keyEnter)
	e.at(0, 1)
	e.typ("n")
	e.at(0, 5)
}

func TestSearchPromptCanBeCancelled(t *testing.T) {
	e := newEd(t, "abc\nabc\n").typ("/abc").send(keyEsc)
	e.mode(editor.ModeNormal)
	e.at(0, 0)
	if !strings.Contains(e.screen(), "NORMAL") {
		t.Error("tras cancelar vuelve a Normal")
	}
}

// ---------- línea de comandos ----------

func TestPromptShowsTextAndCancels(t *testing.T) {
	e := newEd(t, "abc\n").typ(":wq")
	e.mode(editor.ModeCommand)
	if !strings.Contains(e.screen(), ":wq") || !strings.Contains(e.screen(), "COMMAND") {
		t.Errorf("el prompt debe verse:\n%s", e.screen())
	}
	e.send(keyEsc)
	e.mode(editor.ModeNormal)
	if e.quit || e.m.Buffer().Modified() {
		t.Error("cancelar no ejecuta nada")
	}
	e.typ(":").send(keyBackspace) // Backspace en una línea vacía cancela
	e.mode(editor.ModeNormal)
}

func TestPromptEditing(t *testing.T) {
	e := newEd(t, "abc\n").typ(":wx").send(keyLeft, keyBackspace).typ("q").send(keyEnd)
	if !strings.Contains(e.screen(), ":qx") {
		t.Errorf("edición en la línea de comandos:\n%s", e.screen())
	}
	e.send(keyBackspace, keyHome, keyDelete).typ("w")
	if !strings.Contains(e.screen(), ":w") {
		t.Errorf("%s", e.screen())
	}
	e.send(tea.PasteMsg{Content: "x\ny"})
	if !strings.Contains(e.screen(), ":wx y") {
		t.Errorf("pegar en el prompt:\n%s", e.screen())
	}
}

func TestPromptHistory(t *testing.T) {
	p := filepath.Join(t.TempDir(), "f.txt")
	os.WriteFile(p, []byte("a\n"), 0o644)
	m, _ := editor.Open(p)
	e := &ed{t: t, m: m}
	e.send(tea.WindowSizeMsg{Width: 80, Height: 24}).ex("nada").ex("w")
	e.typ(":").send(keyUp)
	if !strings.Contains(e.screen(), ":w") {
		t.Errorf("la última línea del historial:\n%s", e.screen())
	}
	e.send(keyUp)
	if !strings.Contains(e.screen(), ":nada") {
		t.Errorf("la anterior:\n%s", e.screen())
	}
	e.send(keyDown, keyDown)
	if strings.Contains(e.screen(), ":w") || strings.Contains(e.screen(), ":nada") {
		t.Errorf("volver a la línea vacía:\n%s", e.screen())
	}
}

func TestExWriteQuitFamily(t *testing.T) {
	setup := func(t *testing.T) (*ed, string) {
		p := filepath.Join(t.TempDir(), "f.txt")
		os.WriteFile(p, []byte("a\n"), 0o644)
		m, _ := editor.Open(p)
		e := &ed{t: t, m: m}
		e.send(tea.WindowSizeMsg{Width: 80, Height: 24}).typ("ix").send(keyEsc)
		return e, p
	}
	t.Run(":w guarda y no sale", func(t *testing.T) {
		e, p := setup(t)
		e.ex("w")
		if got, _ := os.ReadFile(p); string(got) != "xa\n" || e.quit {
			t.Errorf("archivo %q quit=%v", got, e.quit)
		}
	})
	t.Run(":wq guarda y sale", func(t *testing.T) {
		e, p := setup(t)
		e.ex("wq")
		if got, _ := os.ReadFile(p); string(got) != "xa\n" || !e.quit {
			t.Errorf("archivo %q quit=%v", got, e.quit)
		}
	})
	t.Run(":x guarda y sale", func(t *testing.T) {
		e, p := setup(t)
		e.ex("x")
		if got, _ := os.ReadFile(p); string(got) != "xa\n" || !e.quit {
			t.Errorf("archivo %q quit=%v", got, e.quit)
		}
	})
	t.Run(":x sin cambios sale sin escribir", func(t *testing.T) {
		e, _ := setup(t)
		e.ex("w").ex("x")
		if !e.quit {
			t.Error("debe salir")
		}
	})
	t.Run(":q con cambios avisa", func(t *testing.T) {
		e, p := setup(t)
		e.ex("q")
		if e.quit || !e.m.DialogOpen() || !strings.Contains(e.m.Message(), "E37") {
			t.Errorf("quit=%v dialog=%v msg=%q", e.quit, e.m.DialogOpen(), e.m.Message())
		}
		if !strings.Contains(e.screen(), "Cambios sin guardar") || !strings.Contains(e.screen(), "E37") {
			t.Errorf("el diálogo y el aviso deben verse:\n%s", e.screen())
		}
		if got, _ := os.ReadFile(p); string(got) != "a\n" {
			t.Error("no debe haber guardado")
		}
	})
	t.Run(":q! descarta", func(t *testing.T) {
		e, p := setup(t)
		e.ex("q!")
		if !e.quit {
			t.Error("debe salir")
		}
		if got, _ := os.ReadFile(p); string(got) != "a\n" {
			t.Error("no debe haber guardado")
		}
	})
	t.Run(":w ruta guarda en otra ruta", func(t *testing.T) {
		e, _ := setup(t)
		other := filepath.Join(t.TempDir(), "otra.txt")
		e.ex("w " + other)
		if got, _ := os.ReadFile(other); string(got) != "xa\n" || e.quit {
			t.Errorf("archivo %q", got)
		}
		if !strings.Contains(e.screen(), "otra.txt") {
			t.Error("la statusline muestra el nuevo nombre")
		}
		if e.m.Buffer().Modified() {
			t.Error("tras guardar no está modificado")
		}
	})
	t.Run(":write y :quit son sinónimos", func(t *testing.T) {
		e, p := setup(t)
		e.ex("write").ex("quit")
		if got, _ := os.ReadFile(p); string(got) != "xa\n" || !e.quit {
			t.Errorf("archivo %q quit=%v", got, e.quit)
		}
	})
	t.Run(":wq sin nombre falla sin salir", func(t *testing.T) {
		e := newEd(t, "").typ("ix").send(keyEsc).ex("wq")
		if e.quit || !strings.HasPrefix(e.m.Message(), "E32") {
			t.Errorf("quit=%v msg=%q", e.quit, e.m.Message())
		}
	})
}

func TestExOpenFile(t *testing.T) {
	dir := t.TempDir()
	a, b := filepath.Join(dir, "a.txt"), filepath.Join(dir, "b.txt")
	os.WriteFile(a, []byte("aaa\n"), 0o644)
	os.WriteFile(b, []byte("bbb\nccc\n"), 0o644)
	m, _ := editor.Open(a)
	e := &ed{t: t, m: m}
	e.send(tea.WindowSizeMsg{Width: 80, Height: 24}).ex("e " + b)
	e.wantText("bbb\nccc\n")
	if !strings.Contains(e.screen(), "b.txt") || e.m.Buffer().Modified() {
		t.Errorf("tras abrir:\n%s", e.screen())
	}
	e.typ("x").ex("e " + a) // con cambios: no abre
	if !strings.HasPrefix(e.m.Message(), "E37") {
		t.Errorf("mensaje = %q", e.m.Message())
	}
	e.wantText("bb\nccc\n")
	e.ex("e! " + a)
	e.wantText("aaa\n")
	e.ex("e " + filepath.Join(dir, "nuevo.txt"))
	e.wantText("") // un archivo que no existe se abre vacío
	e.ex("e " + dir)
	if !strings.HasPrefix(e.m.Message(), "E484") {
		t.Errorf("abrir un directorio: %q", e.m.Message())
	}
}

func TestExGoToLine(t *testing.T) {
	seventeen := strings.Repeat("  x\n", 17)
	e := newEd(t, seventeen).ex("9999")
	e.at(16, 2) // se limita a la última línea
	e.ex("5")
	e.at(4, 2)
	e.ex("0")
	e.at(0, 2)
	e.ex("12")
	e.at(11, 2)
	e.wantText(seventeen)
}

func TestExUnknownCommand(t *testing.T) {
	e := newEd(t, "abc\n").ex("nada")
	if got := e.m.Message(); got != "E492: no es un comando del editor: nada" {
		t.Errorf("mensaje = %q", got)
	}
	e.wantText("abc\n")
	if e.m.Buffer().Modified() || e.quit {
		t.Error("un comando desconocido no cambia nada")
	}
	e.ex("   ")
	e.ex("")
}

func TestExSameResultAsShortcut(t *testing.T) {
	mk := func() (*ed, string) {
		p := filepath.Join(t.TempDir(), "f.txt")
		os.WriteFile(p, []byte("abc\n"), 0o644)
		m, _ := editor.Open(p)
		e := &ed{t: t, m: m}
		e.send(tea.WindowSizeMsg{Width: 80, Height: 24}).typ("ix").send(keyEsc)
		return e, p
	}
	e1, p1 := mk()
	e2, p2 := mk()
	e1.send(keyCtrlS)
	e2.ex("w")
	b1, _ := os.ReadFile(p1)
	b2, _ := os.ReadFile(p2)
	if string(b1) != string(b2) || e1.m.Buffer().Modified() != e2.m.Buffer().Modified() {
		t.Errorf("Ctrl+S y :w deben producir lo mismo: %q vs %q", b1, b2)
	}
	if strings.ReplaceAll(e1.m.Message(), p1, "") != strings.ReplaceAll(e2.m.Message(), p2, "") {
		t.Errorf("mensajes distintos: %q vs %q", e1.m.Message(), e2.m.Message())
	}
}

// ---------- registro de comandos ----------

func TestRegistryMetadata(t *testing.T) {
	list := editor.Registry().List()
	if len(list) < 40 {
		t.Fatalf("pocos comandos registrados: %d", len(list))
	}
	for _, c := range list {
		if c.Desc == "" || len(c.Keys) == 0 {
			t.Errorf("%s: faltan descripción o atajos", c.Name)
		}
	}
	for _, want := range []string{"file.save", "app.quit", "cursor.first_line", "operator.delete", "edit.undo", "mode.visual", "search.next"} {
		if _, ok := editor.Registry().Get(want); !ok {
			t.Errorf("falta el comando %s", want)
		}
	}
}
