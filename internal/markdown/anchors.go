package markdown

import "sort"

// OutLine devuelve la línea del resultado que corresponde a la línea src del fuente. Se interpola entre el ancla anterior y
// la siguiente (por proporción); sin anclas se usa la proporción global del documento.
func (d Doc) OutLine(src int) int {
	if len(d.Lines) == 0 {
		return 0
	}
	src = max(0, min(src, max(0, d.SrcLines-1)))
	if len(d.Anchors) == 0 {
		return proportional(src, d.SrcLines, len(d.Lines))
	}
	i := sort.Search(len(d.Anchors), func(i int) bool { return d.Anchors[i].Src > src }) - 1
	if i < 0 { // antes del primer bloque (líneas en blanco iniciales)
		return 0
	}
	a := d.Anchors[i]
	nextSrc, nextOut := d.SrcLines, len(d.Lines)
	if i+1 < len(d.Anchors) {
		nextSrc, nextOut = d.Anchors[i+1].Src, d.Anchors[i+1].Out
	}
	if nextSrc <= a.Src {
		return a.Out
	}
	out := a.Out + (src-a.Src)*(nextOut-a.Out)/(nextSrc-a.Src)
	return max(0, min(out, len(d.Lines)-1))
}

// SrcLine devuelve la línea del fuente que corresponde a la línea out del resultado (la inversa aproximada de OutLine).
func (d Doc) SrcLine(out int) int {
	if d.SrcLines == 0 {
		return 0
	}
	out = max(0, min(out, max(0, len(d.Lines)-1)))
	if len(d.Anchors) == 0 {
		return proportional(out, len(d.Lines), d.SrcLines)
	}
	i := sort.Search(len(d.Anchors), func(i int) bool { return d.Anchors[i].Out > out }) - 1
	if i < 0 {
		return 0
	}
	a := d.Anchors[i]
	nextSrc, nextOut := d.SrcLines, len(d.Lines)
	if i+1 < len(d.Anchors) {
		nextSrc, nextOut = d.Anchors[i+1].Src, d.Anchors[i+1].Out
	}
	if nextOut <= a.Out {
		return a.Src
	}
	src := a.Src + (out-a.Out)*(nextSrc-a.Src)/(nextOut-a.Out)
	return max(0, min(src, d.SrcLines-1))
}

// HeadingAt devuelve la línea del fuente del encabezado más cercano por encima (o en) la línea out del resultado.
// Si no hay ninguno por encima, devuelve el primer encabezado del documento; ok es false si no hay encabezados.
func (d Doc) HeadingAt(out int) (src int, ok bool) {
	best := -1
	for i, a := range d.Anchors {
		if !a.Heading {
			continue
		}
		if best < 0 {
			best = i // el primer encabezado, por si no hay ninguno por encima
		}
		if a.Out <= out {
			best = i
		}
	}
	if best < 0 {
		return 0, false
	}
	return d.Anchors[best].Src, true
}

func proportional(v, from, to int) int {
	if from <= 1 || to <= 1 {
		return 0
	}
	return max(0, min(v*(to-1)/(from-1), to-1))
}
