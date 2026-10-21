//go:build unix

package buffer

import (
	"os"
	"syscall"
)

// keepOwner intenta conservar propietario y grupo del archivo original; si no hay permiso, se ignora.
func keepOwner(path string, info os.FileInfo) {
	if st, ok := info.Sys().(*syscall.Stat_t); ok {
		_ = os.Chown(path, int(st.Uid), int(st.Gid))
	}
}
