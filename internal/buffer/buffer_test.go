package buffer_test

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Javier-GarciaP/fosforo/internal/buffer"
)

func writeFile(t *testing.T, name string, data []byte) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(p, data, 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestRoundTripByteForByte(t *testing.T) {
	for _, tc := range []struct {
		name, content string
		lines         int
		crlf          bool
	}{
		{"vacío", "", 1, false},
		{"solo salto", "\n", 1, false},
		{"una línea sin final", "hola", 1, false},
		{"una línea con final", "hola\n", 1, false},
		{"varias LF", "a\nb\nc\n", 3, false},
		{"varias LF sin final", "a\nb\nc", 3, false},
		{"CRLF", "a\r\nb\r\n", 2, true},
		{"CRLF sin final", "a\r\nb", 2, true},
		{"mixto", "a\r\nb\nc\r\n", 3, false},
		{"líneas vacías", "\n\n\na\n\n", 5, false},
		{"unicode", "héllo\n日本語\n🧑‍💻\n", 3, false},
		{"tabs y espacios finales", "\tx  \n  \t\n", 2, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := writeFile(t, "f.txt", []byte(tc.content))
			b, err := buffer.Open(p)
			if err != nil {
				t.Fatal(err)
			}
			if got := b.Text(); got != tc.content {
				t.Errorf("Text() = %q, quería %q", got, tc.content)
			}
			if b.LineCount() != tc.lines {
				t.Errorf("líneas = %d, quería %d", b.LineCount(), tc.lines)
			}
			if b.CRLF() != tc.crlf {
				t.Errorf("CRLF = %v, quería %v", b.CRLF(), tc.crlf)
			}
			// guardar sin editar deja el archivo idéntico
			if err := b.Save(); err != nil {
				t.Fatal(err)
			}
			got, _ := os.ReadFile(p)
			if !bytes.Equal(got, []byte(tc.content)) {
				t.Errorf("tras guardar: %q, quería %q", got, tc.content)
			}
		})
	}
}

func TestOpenMissingFile(t *testing.T) {
	p := filepath.Join(t.TempDir(), "nuevo.txt")
	b, err := buffer.Open(p)
	if err != nil {
		t.Fatal(err)
	}
	if b.Path() != p || b.LineCount() != 1 || b.Line(0) != "" || b.Modified() {
		t.Fatalf("buffer inicial inesperado: %q %d %q %v", b.Path(), b.LineCount(), b.Line(0), b.Modified())
	}
	if _, err := os.Stat(p); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("abrir no debe crear el archivo")
	}
	if _, _, err := b.Insert(buffer.Pos{}, "hola"); err != nil {
		t.Fatal(err)
	}
	if err := b.Save(); err != nil {
		t.Fatal(err)
	}
	if got, _ := os.ReadFile(p); string(got) != "hola\n" {
		t.Errorf("archivo nuevo = %q, quería %q", got, "hola\n")
	}
}

func TestOpenDirectoryFails(t *testing.T) {
	if _, err := buffer.Open(t.TempDir()); err == nil {
		t.Fatal("abrir un directorio debía fallar")
	}
}

func TestReadOnlyDetection(t *testing.T) {
	big := bytes.Repeat([]byte("a"), buffer.MaxEditableSize+1)
	for _, tc := range []struct {
		name   string
		data   []byte
		wantRO bool
		reason string
	}{
		{"texto normal", []byte("hola\n"), false, ""},
		{"NUL", []byte("ab\x00cd"), true, "NUL"},
		{"UTF-8 inválido", []byte("ab\xffcd"), true, "UTF-8"},
		{"justo en el límite", big[:buffer.MaxEditableSize], false, ""},
		{"mayor de 10 MiB", big, true, "10 MiB"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			b, err := buffer.Open(writeFile(t, "f", tc.data))
			if err != nil {
				t.Fatal(err)
			}
			ro, why := b.ReadOnly()
			if ro != tc.wantRO || !strings.Contains(why, tc.reason) {
				t.Errorf("ReadOnly() = %v, %q; quería %v con %q", ro, why, tc.wantRO, tc.reason)
			}
			if ro {
				if _, _, err := b.Insert(buffer.Pos{}, "x"); !errors.Is(err, buffer.ErrReadOnly) {
					t.Errorf("editar debía devolver ErrReadOnly, devolvió %v", err)
				}
				if b.Modified() || b.Text() != string(tc.data) {
					t.Error("el texto no debe cambiar en un buffer de solo lectura")
				}
				if err := b.Save(); !errors.Is(err, buffer.ErrReadOnly) {
					t.Errorf("guardar debía devolver ErrReadOnly, devolvió %v", err)
				}
			}
		})
	}
}

func TestEditsAndUndoByInversion(t *testing.T) {
	b := buffer.FromString("", "hola\nmundo\n")
	orig := b.Text()

	e1, end, err := b.Insert(buffer.Pos{0, 4}, " a todos\nqueridos")
	if err != nil {
		t.Fatal(err)
	}
	if want := (buffer.Pos{1, 8}); end != want {
		t.Errorf("fin = %+v, quería %+v", end, want)
	}
	if got := b.Text(); got != "hola a todos\nqueridos\nmundo\n" {
		t.Errorf("tras insertar: %q", got)
	}
	e2, err := b.Delete(buffer.Pos{0, 2}, buffer.Pos{2, 2})
	if err != nil {
		t.Fatal(err)
	}
	if got := b.Text(); got != "hondo\n" {
		t.Errorf("tras borrar: %q", got)
	}
	// deshacer en orden inverso devuelve el texto original
	for _, e := range []buffer.Edit{e2, e1} {
		if _, err := b.Apply(e.Invert()); err != nil {
			t.Fatal(err)
		}
	}
	if b.Text() != orig {
		t.Errorf("tras deshacer: %q, quería %q", b.Text(), orig)
	}
}

func TestApplyErrors(t *testing.T) {
	b := buffer.FromString("", "abc\n")
	if _, err := b.Apply(buffer.Edit{Start: buffer.Pos{5, 0}}); !errors.Is(err, buffer.ErrInvalidPos) {
		t.Errorf("línea inexistente: %v", err)
	}
	if _, err := b.Apply(buffer.Edit{Start: buffer.Pos{0, 9}}); !errors.Is(err, buffer.ErrInvalidPos) {
		t.Errorf("columna inexistente: %v", err)
	}
	if _, err := b.Apply(buffer.Edit{Start: buffer.Pos{0, 0}, Deleted: "xyz"}); !errors.Is(err, buffer.ErrMismatch) {
		t.Errorf("texto distinto: %v", err)
	}
	if _, err := b.Apply(buffer.Edit{Start: buffer.Pos{0, 2}, Deleted: "cd"}); !errors.Is(err, buffer.ErrInvalidPos) {
		t.Errorf("borrado más allá del final: %v", err)
	}
	if b.Modified() || b.Text() != "abc\n" {
		t.Error("las ediciones fallidas no deben modificar el buffer")
	}
}

func TestInsertNormalizesCRLF(t *testing.T) {
	b := buffer.FromString("", "")
	if _, _, err := b.Insert(buffer.Pos{}, "a\r\nb"); err != nil {
		t.Fatal(err)
	}
	if b.LineCount() != 2 || b.Line(0) != "a" || b.Line(1) != "b" {
		t.Errorf("líneas: %q %q (%d)", b.Line(0), b.Line(1), b.LineCount())
	}
}

func TestUnicodeEdits(t *testing.T) {
	const family = "👨‍👩‍👧" // un solo grafema formado por 5 puntos de código
	const accent = "é"    // e + acento combinante: un grafema
	b := buffer.FromString("", "a"+family+accent+"日z\n")
	if got := b.LineLen(0); got != 5 {
		t.Fatalf("LineLen = %d, quería 5 grafemas", got)
	}
	// borrar hacia atrás desde después del emoji elimina el emoji completo
	start, _ := b.Prev(buffer.Pos{0, 2})
	if _, err := b.Delete(start, buffer.Pos{0, 2}); err != nil {
		t.Fatal(err)
	}
	if got := b.Line(0); got != "a"+accent+"日z" {
		t.Errorf("tras borrar el emoji: %q", got)
	}
	// borrar el carácter con acento combinante lo elimina entero
	if _, err := b.Delete(buffer.Pos{0, 1}, buffer.Pos{0, 2}); err != nil {
		t.Fatal(err)
	}
	if got := b.Line(0); got != "a日z" {
		t.Errorf("tras borrar e+acento: %q", got)
	}
	// insertar entre grafemas
	if _, _, err := b.Insert(buffer.Pos{0, 2}, "🙂"); err != nil {
		t.Fatal(err)
	}
	if got := b.Line(0); got != "a日🙂z" {
		t.Errorf("tras insertar: %q", got)
	}
}

func TestModifiedFlag(t *testing.T) {
	p := writeFile(t, "f.txt", []byte("uno\n"))
	b, _ := buffer.Open(p)
	if b.Modified() {
		t.Fatal("recién abierto no debe estar modificado")
	}
	if _, _, err := b.Insert(buffer.Pos{0, 3}, "!"); err != nil {
		t.Fatal(err)
	}
	if !b.Modified() {
		t.Fatal("tras editar debe estar modificado")
	}
	if err := b.Save(); err != nil {
		t.Fatal(err)
	}
	if b.Modified() {
		t.Fatal("tras guardar no debe estar modificado")
	}
	if got, _ := os.ReadFile(p); string(got) != "uno!\n" {
		t.Errorf("contenido en disco: %q", got)
	}
}

func TestSaveKeepsPermissionsAndLeavesNoTemp(t *testing.T) {
	p := writeFile(t, "script.sh", []byte("echo hi\n"))
	if err := os.Chmod(p, 0o750); err != nil {
		t.Fatal(err)
	}
	b, _ := buffer.Open(p)
	b.Insert(buffer.Pos{0, 0}, "#!/bin/sh\n")
	if err := b.Save(); err != nil {
		t.Fatal(err)
	}
	info, _ := os.Stat(p)
	if info.Mode().Perm() != 0o750 {
		t.Errorf("permisos = %v, quería 0750", info.Mode().Perm())
	}
	entries, _ := os.ReadDir(filepath.Dir(p))
	if len(entries) != 1 {
		t.Errorf("deben quedar solo el archivo, hay: %v", entries)
	}
}

func TestSaveThroughSymlinkKeepsLink(t *testing.T) {
	dir := t.TempDir()
	real := filepath.Join(dir, "real.txt")
	link := filepath.Join(dir, "link.txt")
	os.WriteFile(real, []byte("a\n"), 0o644)
	if err := os.Symlink(real, link); err != nil {
		t.Skip("enlaces simbólicos no disponibles")
	}
	b, _ := buffer.Open(link)
	b.Insert(buffer.Pos{0, 1}, "b")
	if err := b.Save(); err != nil {
		t.Fatal(err)
	}
	if fi, _ := os.Lstat(link); fi.Mode()&os.ModeSymlink == 0 {
		t.Error("el enlace simbólico fue reemplazado")
	}
	if got, _ := os.ReadFile(real); string(got) != "ab\n" {
		t.Errorf("destino = %q", got)
	}
}

func TestSaveFailureLeavesOriginalIntact(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("como root no se puede quitar el permiso de escritura")
	}
	dir := t.TempDir()
	p := filepath.Join(dir, "f.txt")
	os.WriteFile(p, []byte("original\n"), 0o644)
	b, _ := buffer.Open(p)
	b.Insert(buffer.Pos{0, 0}, "nuevo ")
	if err := os.Chmod(dir, 0o555); err != nil {
		t.Fatal(err)
	}
	defer os.Chmod(dir, 0o755)

	if err := b.Save(); err == nil {
		t.Fatal("guardar en un directorio sin escritura debía fallar")
	}
	if got, _ := os.ReadFile(p); string(got) != "original\n" {
		t.Errorf("el original cambió: %q", got)
	}
	if !b.Modified() {
		t.Error("tras un fallo el buffer debe seguir modificado")
	}
	os.Chmod(dir, 0o755)
	if err := b.Save(); err != nil {
		t.Fatalf("reintentar tras arreglar los permisos: %v", err)
	}
}

func TestSaveWithoutPath(t *testing.T) {
	if err := buffer.New("").Save(); !errors.Is(err, buffer.ErrNoPath) {
		t.Errorf("error = %v, quería ErrNoPath", err)
	}
}

func TestEmptyFileStaysEmpty(t *testing.T) {
	p := writeFile(t, "vacio", nil)
	b, _ := buffer.Open(p)
	b.Insert(buffer.Pos{}, "x")
	b.Delete(buffer.Pos{0, 0}, buffer.Pos{0, 1})
	if b.Text() != "" {
		t.Errorf("Text() = %q, quería vacío", b.Text())
	}
}

func TestRange(t *testing.T) {
	b := buffer.FromString("", "héllo\nmundo\nfin\n")
	for _, tc := range []struct {
		a, b buffer.Pos
		want string
	}{
		{buffer.Pos{0, 1}, buffer.Pos{0, 3}, "él"},
		{buffer.Pos{0, 4}, buffer.Pos{1, 2}, "o\nmu"},
		{buffer.Pos{0, 0}, buffer.Pos{2, 3}, "héllo\nmundo\nfin"},
		{buffer.Pos{1, 2}, buffer.Pos{0, 4}, "o\nmu"}, // orden invertido
	} {
		if got := b.Range(tc.a, tc.b); got != tc.want {
			t.Errorf("Range(%v,%v) = %q, quería %q", tc.a, tc.b, got, tc.want)
		}
	}
}
