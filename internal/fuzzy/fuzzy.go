// Package fuzzy ordena y filtra textos con coincidencia difusa al estilo de fzf: los caracteres escritos deben aparecer en
// orden en el texto (subsecuencia), y se puntúa mejor lo más compacto y lo que empieza en un límite de palabra o de ruta.
package fuzzy

import (
	"sort"
	"strings"
	"unicode"
)

// Puntuaciones (inspiradas en fzf v2).
const (
	scoreMatch        = 16
	scoreGapStart     = -3
	scoreGapExtension = -1
	bonusBoundary     = 10 // tras un separador (espacio, «/», «_», «-», «.»...)
	bonusCamel        = 7  // letra mayúscula tras minúscula
	bonusConsecutive  = 8  // carácter pegado al anterior
	bonusFirstChar    = 8  // primer carácter del texto
	bonusBasename     = 10 // la coincidencia cae en el nombre de archivo (tras el último «/»)
)

// Match es una coincidencia: su puntuación y los índices (de runa) del texto que coincidieron con el patrón.
type Match struct {
	Score int
	Pos   []int
}

func isSeparator(r rune) bool {
	return r == '/' || r == '\\' || r == '_' || r == '-' || r == '.' || r == ' ' || r == ':' || r == '(' || r == ')'
}

// bonusAt es la bonificación por coincidir en la posición i del texto.
func bonusAt(text []rune, i int) int {
	if i == 0 {
		return bonusFirstChar
	}
	prev, cur := text[i-1], text[i]
	switch {
	case isSeparator(prev) && !isSeparator(cur):
		return bonusBoundary
	case unicode.IsLower(prev) && unicode.IsUpper(cur):
		return bonusCamel
	}
	return 0
}

// Score busca pattern en text. Es insensible a mayúsculas salvo que el patrón contenga alguna (smart-case).
// Un patrón vacío coincide con todo y puntúa 0.
func Score(pattern, text string) (Match, bool) {
	if pattern == "" {
		return Match{}, true
	}
	pat, txt := []rune(pattern), []rune(text)
	if len(pat) > len(txt) {
		return Match{}, false
	}
	caseSensitive := strings.IndexFunc(pattern, unicode.IsUpper) >= 0
	norm := func(r rune) rune {
		if caseSensitive {
			return r
		}
		return unicode.ToLower(r)
	}
	for i := range pat {
		pat[i] = norm(pat[i])
	}

	// Filtro rápido: ¿es subsecuencia?
	j := 0
	for _, r := range txt {
		if j < len(pat) && norm(r) == pat[j] {
			j++
		}
	}
	if j < len(pat) {
		return Match{}, false
	}

	n, m := len(txt), len(pat)
	baseStart := 0 // inicio del nombre de archivo (tras el último «/»)
	for i, r := range txt {
		if r == '/' || r == '\\' {
			baseStart = i + 1
		}
	}
	const neg = -1 << 30
	// dp[i][j]: mejor puntuación con pat[:j+1] coincidiendo y pat[j] sobre txt[i]; from[i][j]: columna previa usada.
	dp := make([][]int, m)
	from := make([][]int, m)
	for j := range dp {
		dp[j] = make([]int, n)
		from[j] = make([]int, n)
		for i := range dp[j] {
			dp[j][i], from[j][i] = neg, -1
		}
	}
	for j := 0; j < m; j++ {
		x, xk := neg, -1 // mejor puntuación llegando a la columna actual con un hueco desde la coincidencia anterior, y su columna
		for i := 0; i < n; i++ {
			if j > 0 {
				if x > neg {
					x += scoreGapExtension // un carácter más saltado
				}
				if i >= 2 && dp[j-1][i-2] > neg {
					if v := dp[j-1][i-2] + scoreGapStart; v > x {
						x, xk = v, i-2
					}
				}
			}
			if norm(txt[i]) != pat[j] {
				continue
			}
			bonus := bonusAt(txt, i)
			if i >= baseStart { // un texto sin «/» es todo «nombre de archivo»
				bonus += bonusBasename / 2
			}
			if j == 0 {
				dp[j][i] = scoreMatch + bonus
				continue
			}
			cand, candFrom := neg, -1
			if i >= 1 && dp[j-1][i-1] > neg { // pegado al anterior
				cand, candFrom = dp[j-1][i-1]+scoreMatch+bonus+bonusConsecutive, i-1
			}
			if x > neg { // con hueco
				if v := x + scoreMatch + bonus; v > cand {
					cand, candFrom = v, xk
				}
			}
			dp[j][i], from[j][i] = cand, candFrom
		}
	}
	bestScore, end := neg, -1
	for i := 0; i < n; i++ {
		if dp[m-1][i] > bestScore {
			bestScore, end = dp[m-1][i], i
		}
	}
	if end < 0 {
		return Match{}, false
	}
	pos := make([]int, m)
	i := end
	for j := m - 1; j >= 0; j-- {
		pos[j] = i
		i = from[j][i]
	}
	// Un texto corto con la misma coincidencia es mejor: pequeña penalización por longitud.
	score := bestScore - n/8
	if pos[0] >= baseStart {
		score += bonusBasename
	}
	return Match{Score: score, Pos: pos}, true
}

// Result es un elemento que coincide: su índice en la lista original y su coincidencia.
type Result struct {
	Index int
	Match
}

// Filter devuelve los elementos de items que coinciden con pattern, ordenados de mejor a peor (a igualdad, el texto más
// corto y luego el orden original), con como mucho limit resultados (limit <= 0: todos).
func Filter(pattern string, items []string, limit int) []Result {
	out := make([]Result, 0, min(len(items), 1024))
	for i, it := range items {
		if m, ok := Score(pattern, it); ok {
			out = append(out, Result{Index: i, Match: m})
		}
	}
	Sort(out, items)
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out
}

// Sort ordena resultados: mayor puntuación primero; a igualdad, el texto más corto; después, el orden original.
func Sort(rs []Result, items []string) {
	sort.SliceStable(rs, func(a, b int) bool {
		x, y := rs[a], rs[b]
		if x.Score != y.Score {
			return x.Score > y.Score
		}
		if lx, ly := len(items[x.Index]), len(items[y.Index]); lx != ly {
			return lx < ly
		}
		return x.Index < y.Index
	})
}
