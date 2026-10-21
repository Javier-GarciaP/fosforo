// Package search busca texto en todos los archivos de un proyecto. Usa ripgrep (rg) si está instalado, respetando los
// archivos de ignorados, y si no, un buscador propio que omite las carpetas ignoradas y los archivos binarios.
// Los resultados se emiten en lotes mientras se encuentran y la búsqueda se puede cancelar con un contexto.
package search

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/Javier-GarciaP/fosforo/internal/buffer"
	"github.com/Javier-GarciaP/fosforo/internal/fs"
)

// DefaultLimit es el número máximo de resultados.
const DefaultLimit = 1000

// Match es una coincidencia: el archivo (relativo a la raíz, con «/»), la línea (desde 1), la columna del primer
// carácter coincidente (en grafemas, desde 0) y el texto de la línea.
type Match struct {
	Path string
	Line int
	Col  int
	Text string
}

// Options configura una búsqueda. La consulta es literal (no una expresión regular) e insensible a mayúsculas
// salvo que contenga alguna mayúscula (smart-case).
type Options struct {
	Root         string
	Query        string
	Limit        int  // 0: DefaultLimit
	ForceBuiltin bool // no usar rg aunque esté instalado
}

// Result resume una búsqueda terminada.
type Result struct {
	Truncated bool   // se alcanzó el límite de resultados
	Backend   string // "rg" o "builtin"
}

// RipgrepPath devuelve la ruta de rg, o "" si no está instalado.
func RipgrepPath() string {
	p, err := exec.LookPath("rg")
	if err != nil {
		return ""
	}
	return p
}

// Run busca y llama a emit con lotes de resultados a medida que aparecen. Termina al acabar, al alcanzar el límite o
// cuando ctx se cancela (en ese caso devuelve ctx.Err()). Una consulta vacía no busca nada.
func Run(ctx context.Context, o Options, emit func([]Match)) (Result, error) {
	if o.Limit <= 0 {
		o.Limit = DefaultLimit
	}
	if o.Query == "" {
		return Result{Backend: "builtin"}, nil
	}
	if rg := RipgrepPath(); rg != "" && !o.ForceBuiltin {
		res, err := runRipgrep(ctx, rg, o, emit)
		if err == nil || ctx.Err() != nil {
			return res, err
		}
		// si rg falla por algún motivo, se recurre al buscador propio
	}
	return runBuiltin(ctx, o, emit)
}

// batcher agrupa resultados y los emite cada cierto número o tiempo.
type batcher struct {
	emit  func([]Match)
	buf   []Match
	last  time.Time
	total int
	limit int
}

func (b *batcher) add(m Match) (full bool) {
	b.buf = append(b.buf, m)
	b.total++
	if len(b.buf) >= 50 || time.Since(b.last) > 30*time.Millisecond {
		b.flush()
	}
	return b.total >= b.limit
}

func (b *batcher) flush() {
	if len(b.buf) > 0 {
		b.emit(b.buf)
		b.buf = nil
	}
	b.last = time.Now()
}

// ---- ripgrep ----

type rgLine struct {
	Type string `json:"type"`
	Data struct {
		Path struct {
			Text string `json:"text"`
		} `json:"path"`
		Lines struct {
			Text string `json:"text"`
		} `json:"lines"`
		LineNumber int `json:"line_number"`
		Submatches []struct {
			Start int `json:"start"`
		} `json:"submatches"`
	} `json:"data"`
}

func runRipgrep(ctx context.Context, rg string, o Options, emit func([]Match)) (Result, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	cmd := exec.CommandContext(ctx, rg, "--json", "--smart-case", "--fixed-strings", "--sort", "path",
		"--glob", "!.git", "--glob", "!node_modules", "-e", o.Query, ".")
	cmd.Dir = o.Root
	out, err := cmd.StdoutPipe()
	if err != nil {
		return Result{Backend: "rg"}, err
	}
	if err := cmd.Start(); err != nil {
		return Result{Backend: "rg"}, err
	}
	res, perr := parseRipgrep(out, o.Limit, emit, cancel)
	res.Backend = "rg"
	werr := cmd.Wait()
	if ctx.Err() != nil && !res.Truncated { // cancelada desde fuera
		if parent := ctx.Err(); parent != nil {
			return res, parent
		}
	}
	if perr != nil {
		return res, perr
	}
	var ee *exec.ExitError
	if werr != nil && errors.As(werr, &ee) && ee.ExitCode() == 1 { // 1 = sin coincidencias
		return res, nil
	}
	if res.Truncated { // matamos a rg al llegar al límite: no es un error
		return res, nil
	}
	return res, werr
}

// parseRipgrep lee la salida --json de rg. Con el límite alcanzado llama a stop (para detener rg) y deja de leer.
func parseRipgrep(r io.Reader, limit int, emit func([]Match), stop func()) (Result, error) {
	b := &batcher{emit: emit, limit: limit, last: time.Now()}
	var res Result
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 64*1024), 8<<20)
	for sc.Scan() {
		var l rgLine
		if err := json.Unmarshal(sc.Bytes(), &l); err != nil || l.Type != "match" || l.Data.Path.Text == "" {
			continue // «begin», «end», «summary» o rutas que no son UTF-8
		}
		text := strings.TrimRight(l.Data.Lines.Text, "\r\n")
		col := 0
		if len(l.Data.Submatches) > 0 {
			if s := l.Data.Submatches[0].Start; s >= 0 && s <= len(l.Data.Lines.Text) {
				col = buffer.Count(l.Data.Lines.Text[:s])
			}
		}
		path := filepath.ToSlash(strings.TrimPrefix(l.Data.Path.Text, "./"))
		if b.add(Match{Path: path, Line: l.Data.LineNumber, Col: col, Text: text}) {
			res.Truncated = true
			stop()
			break
		}
	}
	b.flush()
	return res, sc.Err()
}

// ---- buscador propio ----

func runBuiltin(ctx context.Context, o Options, emit func([]Match)) (Result, error) {
	res := Result{Backend: "builtin"}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	files := make(chan []string, 4)
	go fs.Walk(ctx, o.Root, nil, 200, files)

	caseSensitive := strings.IndexFunc(o.Query, unicode.IsUpper) >= 0
	q := o.Query
	if !caseSensitive {
		q = strings.ToLower(q)
	}
	b := &batcher{emit: emit, limit: o.Limit, last: time.Now()}
	defer b.flush()
	for batch := range files {
		for _, rel := range batch {
			if ctx.Err() != nil {
				return res, ctx.Err()
			}
			full, err := searchFile(ctx, filepath.Join(o.Root, filepath.FromSlash(rel)), rel, q, caseSensitive, b)
			if err != nil && ctx.Err() != nil {
				return res, ctx.Err()
			}
			if full {
				res.Truncated = true
				cancel()
				for range files { // se vacía el canal para que el recorrido termine
				}
				return res, nil
			}
		}
	}
	return res, ctx.Err()
}

const maxBuiltinFileBytes = 5 << 20

func searchFile(ctx context.Context, path, rel, q string, caseSensitive bool, b *batcher) (full bool, err error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()
	if info, err := f.Stat(); err != nil || !info.Mode().IsRegular() || info.Size() > maxBuiltinFileBytes {
		return false, err
	}
	head := make([]byte, 8192)
	n, _ := f.Read(head)
	if bytes.IndexByte(head[:n], 0) >= 0 { // binario
		return false, nil
	}
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return false, err
	}
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 64*1024), 1<<20)
	for line := 1; sc.Scan(); line++ {
		if line%2000 == 0 && ctx.Err() != nil {
			return false, ctx.Err()
		}
		text := strings.TrimRight(sc.Text(), "\r")
		hay := text
		if !caseSensitive {
			hay = strings.ToLower(text)
		}
		i := strings.Index(hay, q)
		if i < 0 {
			continue
		}
		if len(hay) != len(text) { // minúsculas con distinta longitud en bytes: se recalcula sobre el original
			if i = indexFold(text, q); i < 0 {
				continue
			}
		}
		if b.add(Match{Path: rel, Line: line, Col: buffer.Count(text[:i]), Text: text}) {
			return true, nil
		}
	}
	return false, nil
}

// indexFold busca q (ya en minúsculas) en s sin distinguir mayúsculas y devuelve el desplazamiento en bytes de s.
func indexFold(s, q string) int {
	for i := range s {
		if len(s)-i < 1 {
			break
		}
		j, k := i, 0
		for j < len(s) && k < len(q) {
			rs, size := decodeRune(s[j:])
			rq, qsize := decodeRune(q[k:])
			if unicode.ToLower(rs) != rq {
				break
			}
			j, k = j+size, k+qsize
		}
		if k == len(q) {
			return i
		}
	}
	return -1
}

func decodeRune(s string) (rune, int) { return utf8.DecodeRuneInString(s) }
