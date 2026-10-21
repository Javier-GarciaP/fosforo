package fs

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Errores tipados de las operaciones: se comprueban con errors.Is.
var (
	ErrExists      = errors.New("ya existe")
	ErrInvalidName = errors.New("nombre no válido")
	ErrNotFound    = errors.New("no existe")
	ErrNotDir      = errors.New("no es un directorio")
)

// Error es un fallo de una operación, con la ruta afectada. Unwrap devuelve el error original.
type Error struct {
	Op   string
	Path string
	Err  error
}

func (e *Error) Error() string {
	return fmt.Sprintf("no se pudo %s %s: %v", e.Op, e.Path, simplify(e.Err))
}
func (e *Error) Unwrap() error { return e.Err }

// simplify quita la ruta repetida que añade *os.PathError.
func simplify(err error) error {
	var pe *os.PathError
	if errors.As(err, &pe) {
		return pe.Err
	}
	return err
}

// validName comprueba que name es una ruta relativa que no sale de su carpeta.
func validName(name string) (clean string, isDir bool, err error) {
	isDir = strings.HasSuffix(name, "/")
	clean = filepath.Clean(strings.TrimSpace(name))
	if strings.TrimSpace(name) == "" || clean == "." || clean == ".." || filepath.IsAbs(clean) ||
		strings.HasPrefix(clean, ".."+string(filepath.Separator)) || strings.ContainsRune(name, 0) {
		return "", false, ErrInvalidName
	}
	return clean, isDir, nil
}

// Create crea un archivo o, si name termina en «/», un directorio dentro de dir, con las carpetas intermedias
// que hagan falta. Devuelve la ruta creada. Falla con ErrExists si ya existe y con ErrInvalidName si name sale de dir.
func Create(dir, name string) (string, error) {
	clean, isDir, err := validName(name)
	if err != nil {
		return "", &Error{Op: "crear", Path: name, Err: err}
	}
	target := filepath.Join(dir, clean)
	if _, err := os.Lstat(target); err == nil {
		return "", &Error{Op: "crear", Path: target, Err: ErrExists}
	}
	if isDir {
		if err := os.MkdirAll(target, 0o755); err != nil {
			return "", &Error{Op: "crear", Path: target, Err: err}
		}
		return target, nil
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return "", &Error{Op: "crear", Path: target, Err: err}
	}
	f, err := os.OpenFile(target, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		return "", &Error{Op: "crear", Path: target, Err: err}
	}
	return target, f.Close()
}

// Rename renombra (o mueve) path a newName, interpretado dentro de la carpeta que contiene a path.
// Devuelve la nueva ruta. Falla con ErrExists si el destino ya existe.
func Rename(path, newName string) (string, error) {
	clean, _, err := validName(newName)
	if err != nil {
		return "", &Error{Op: "renombrar", Path: newName, Err: err}
	}
	target := filepath.Join(filepath.Dir(path), clean)
	if target == path {
		return path, nil
	}
	if _, err := os.Lstat(path); err != nil {
		return "", &Error{Op: "renombrar", Path: path, Err: ErrNotFound}
	}
	if _, err := os.Lstat(target); err == nil {
		return "", &Error{Op: "renombrar", Path: target, Err: ErrExists}
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return "", &Error{Op: "renombrar", Path: target, Err: err}
	}
	if err := os.Rename(path, target); err != nil {
		return "", &Error{Op: "renombrar", Path: path, Err: err}
	}
	return target, nil
}

// Remove borra un archivo o un directorio con todo su contenido. El borrado es definitivo: la interfaz exige confirmación.
func Remove(path string) error {
	info, err := os.Lstat(path)
	if err != nil {
		return &Error{Op: "borrar", Path: path, Err: ErrNotFound}
	}
	if info.IsDir() {
		err = os.RemoveAll(path)
	} else {
		err = os.Remove(path)
	}
	if err != nil {
		return &Error{Op: "borrar", Path: path, Err: err}
	}
	return nil
}
