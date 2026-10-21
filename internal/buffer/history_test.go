package buffer_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Javier-GarciaP/fosforo/internal/buffer"
)

func TestUndoRedoSingleEdits(t *testing.T) {
	b := buffer.FromString("", "uno\n")
	b.Insert(buffer.Pos{0, 3}, " dos")
	b.Insert(buffer.Pos{0, 7}, " tres")
	if b.Text() != "uno dos tres\n" {
		t.Fatal(b.Text())
	}
	pos, ok := b.Undo()
	if !ok || b.Text() != "uno dos\n" || pos != (buffer.Pos{0, 7}) {
		t.Errorf("primer deshacer: %q %+v %v", b.Text(), pos, ok)
	}
	b.Undo()
	if b.Text() != "uno\n" {
		t.Errorf("segundo deshacer: %q", b.Text())
	}
	if _, ok := b.Undo(); ok {
		t.Error("no hay más que deshacer")
	}
	pos, ok = b.Redo()
	if !ok || b.Text() != "uno dos\n" || pos != (buffer.Pos{0, 3}) {
		t.Errorf("rehacer: %q %+v", b.Text(), pos)
	}
	b.Redo()
	if _, ok := b.Redo(); ok || b.Text() != "uno dos tres\n" {
		t.Errorf("rehacer de más: %q", b.Text())
	}
}

func TestGroupIsOneUndoStep(t *testing.T) {
	b := buffer.FromString("", "x\n")
	b.BeginGroup(buffer.Pos{0, 1})
	b.Insert(buffer.Pos{0, 1}, "h")
	b.Insert(buffer.Pos{0, 2}, "o")
	b.Insert(buffer.Pos{0, 3}, "la")
	b.EndGroup()
	b.Insert(buffer.Pos{0, 0}, "!") // paso aparte
	if b.Text() != "!xhola\n" {
		t.Fatal(b.Text())
	}
	b.Undo()
	if b.Text() != "xhola\n" {
		t.Errorf("deshacer lo suelto: %q", b.Text())
	}
	pos, _ := b.Undo()
	if b.Text() != "x\n" || pos != (buffer.Pos{0, 1}) {
		t.Errorf("deshacer el grupo completo: %q en %+v", b.Text(), pos)
	}
	b.Redo()
	if b.Text() != "xhola\n" {
		t.Errorf("rehacer el grupo completo: %q", b.Text())
	}
}

func TestEmptyGroupCreatesNoStep(t *testing.T) {
	b := buffer.FromString("", "a\n")
	b.BeginGroup(buffer.Pos{})
	b.EndGroup()
	if b.CanUndo() {
		t.Error("un grupo sin ediciones no debe crear un paso")
	}
}

func TestNewEditDiscardsRedo(t *testing.T) {
	b := buffer.FromString("", "a\n")
	b.Insert(buffer.Pos{0, 1}, "b")
	b.Undo()
	if !b.CanRedo() {
		t.Fatal("debería poder rehacerse")
	}
	b.Insert(buffer.Pos{0, 1}, "c")
	if b.CanRedo() {
		t.Error("una edición nueva descarta el futuro")
	}
	if b.Text() != "ac\n" {
		t.Error(b.Text())
	}
}

func TestUndoMultilineEditsAndUnicode(t *testing.T) {
	orig := "a👨‍👩‍👧b\n日本\nfin\n"
	b := buffer.FromString("", orig)
	b.Delete(buffer.Pos{0, 1}, buffer.Pos{2, 1})
	b.Insert(buffer.Pos{0, 1}, "X\nY")
	for b.CanUndo() {
		b.Undo()
	}
	if b.Text() != orig {
		t.Errorf("tras deshacer todo: %q, quería %q", b.Text(), orig)
	}
	for b.CanRedo() {
		b.Redo()
	}
	if b.Text() != "aX\nYin\n" {
		t.Errorf("tras rehacer todo: %q", b.Text())
	}
}

func TestModifiedFollowsSavedVersion(t *testing.T) {
	p := filepath.Join(t.TempDir(), "f.txt")
	os.WriteFile(p, []byte("a\n"), 0o644)
	b, _ := buffer.Open(p)
	if b.Modified() {
		t.Fatal("recién abierto")
	}
	b.Insert(buffer.Pos{0, 1}, "b")
	if !b.Modified() {
		t.Fatal("tras editar")
	}
	b.Undo()
	if b.Modified() {
		t.Error("deshacer hasta el punto guardado deja el buffer limpio")
	}
	b.Redo()
	if !b.Modified() {
		t.Error("rehacer lo vuelve a modificar")
	}
	if err := b.Save(); err != nil {
		t.Fatal(err)
	}
	if b.Modified() {
		t.Fatal("tras guardar")
	}
	b.Undo()
	if !b.Modified() {
		t.Error("deshacer después de guardar modifica respecto al disco")
	}
	b.Redo()
	if b.Modified() {
		t.Error("rehacer hasta lo guardado deja el buffer limpio")
	}
}

func TestSavedPointInDiscardedFutureStaysModified(t *testing.T) {
	p := filepath.Join(t.TempDir(), "f.txt")
	os.WriteFile(p, []byte("a\n"), 0o644)
	b, _ := buffer.Open(p)
	b.Insert(buffer.Pos{0, 1}, "b")
	b.Save()
	b.Undo()
	b.Insert(buffer.Pos{0, 1}, "c") // misma versión numérica que lo guardado, distinto texto
	if !b.Modified() {
		t.Error("el texto difiere de lo guardado: debe estar modificado")
	}
}

func TestSaveInsideGroupKeepsModifiedAccurate(t *testing.T) {
	p := filepath.Join(t.TempDir(), "f.txt")
	os.WriteFile(p, []byte("a\n"), 0o644)
	b, _ := buffer.Open(p)
	b.BeginGroup(buffer.Pos{})
	b.Insert(buffer.Pos{0, 1}, "b")
	b.Save()
	b.Insert(buffer.Pos{0, 2}, "c") // sigue en el mismo grupo
	if !b.Modified() {
		t.Error("tras guardar a mitad de un grupo, la siguiente edición debe marcar modificado")
	}
	b.EndGroup()
	b.Undo()
	if b.Text() != "a\n" {
		t.Errorf("el grupo se deshace entero: %q", b.Text())
	}
}

func TestReadOnlyEditsAreNotRecorded(t *testing.T) {
	p := filepath.Join(t.TempDir(), "bin")
	os.WriteFile(p, []byte("a\x00"), 0o644)
	b, _ := buffer.Open(p)
	b.Insert(buffer.Pos{}, "x")
	if b.CanUndo() || b.Modified() {
		t.Error("una edición rechazada no entra al historial")
	}
}
