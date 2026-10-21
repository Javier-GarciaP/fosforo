package buffer

import (
	"fmt"
	"os"
	"path/filepath"
)

// Save guarda el buffer en su archivo. Ver SaveAs.
func (b *Buffer) Save() error {
	if b.path == "" {
		return ErrNoPath
	}
	return b.SaveAs(b.path)
}

// SaveAs guarda el buffer en path de forma atómica: escribe un archivo temporal en el mismo directorio y lo renombra
// sobre el destino, de modo que un fallo nunca deja el archivo original a medias. Conserva los permisos existentes.
// Si falla, el buffer sigue marcado como modificado.
func (b *Buffer) SaveAs(path string) error {
	if b.readOnly {
		return ErrReadOnly
	}
	data := []byte(b.Text())
	if err := writeFileAtomic(path, data); err != nil {
		return err
	}
	b.path, b.modified, b.origEmpty = path, false, len(data) == 0
	return nil
}

func writeFileAtomic(path string, data []byte) (err error) {
	target := path
	if resolved, e := filepath.EvalSymlinks(path); e == nil {
		target = resolved // se escribe sobre el destino del enlace, no se reemplaza el enlace
	}
	mode := os.FileMode(0o644)
	info, statErr := os.Stat(target)
	if statErr == nil {
		mode = info.Mode().Perm()
	}

	tmp, err := os.CreateTemp(filepath.Dir(target), "."+filepath.Base(target)+".fos-*")
	if err != nil {
		return fmt.Errorf("no se puede crear el archivo temporal: %w", err)
	}
	defer func() {
		if err != nil {
			tmp.Close()
			os.Remove(tmp.Name())
		}
	}()
	if _, err = tmp.Write(data); err != nil {
		return err
	}
	if err = tmp.Sync(); err != nil {
		return err
	}
	if err = tmp.Chmod(mode); err != nil {
		return err
	}
	if statErr == nil {
		keepOwner(tmp.Name(), info)
	}
	if err = tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), target)
}
