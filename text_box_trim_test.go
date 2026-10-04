package htmlbag

import (
	"testing"

	"github.com/boxesandglue/boxesandglue/backend/bag"
	"github.com/boxesandglue/boxesandglue/backend/node"
)

// trimCSS is a 158pt content area with lines of 16pt, of which about 2.4pt
// below the text are leading: ten lines need 160pt, but the tenth line's text
// ends 2.4pt higher.
const trimCSS = `@page { size: 200pt 198pt; margin: 20pt; }
body { margin: 0 } p { margin: 0; font-family: serif; font-size: 10pt; line-height: 16pt }`

// text-box-trim: trim-end lets the line before a page break fit by its text:
// the leading below it may reach past the content area. Without it, the
// tenth line goes to the next page.
func TestTextBoxTrimEndAtAPageBreak(t *testing.T) {
	bodies := []struct {
		name       string
		body       func(style string) string
		none, trim int
	}{
		{"one paragraph", func(st string) string { return `<p style="` + st + `">` + charLines("A", 14) + `</p>` }, 9, 10},
		{"two paragraphs", func(st string) string {
			return `<p style="` + st + `">` + charLines("A", 5) + `</p><p style="` + st + `">` + charLines("B", 9) + `</p>`
		}, 9, 10},
		// Nine lines leave B's last line one widow; with the trim B fits whole.
		{"a block before the break", func(st string) string {
			return `<p>` + charLines("A", 5) + `</p><p style="` + st + `">` + charLines("B", 5) + `</p><p>C</p>`
		}, 8, 10},
	}
	for _, b := range bodies {
		for _, c := range []struct {
			style string
			trims bool
		}{{``, false}, {`text-box-trim: trim-end`, true}, {`text-box-trim: trim-both`, true}, {`text-box-trim: trim-start`, false}} {
			t.Run(b.name+"/"+c.style, func(t *testing.T) {
				want := b.none
				if c.trims {
					want = b.trim
				}
				n := 0
				for _, l := range placedLines(renderHTMLPages(t, trimCSS, b.body(c.style))) {
					if l.page == 1 {
						n++
					}
				}
				if n != want {
					t.Errorf("page 1 holds %d lines, want %d", n, want)
				}
			})
		}
	}
}

// The same in FlowText regions, and the leading is recorded only on the lines
// of a block that asks for it.
func TestTextBoxTrimEndInRegions(t *testing.T) {
	cb, _ := newFlowBuilder(t, `p { font-family: serif; font-size: 10pt; line-height: 16pt }`)
	tr := flow(t, cb, `<p style="text-box-trim: trim-end">`+charLines("A", 14)+`</p>`, wide("158pt"), wide("1000pt"))
	if got := len(boxLines(tr.filled[0])); got != 10 {
		t.Errorf("region 1 holds %d lines, want 10", got)
	}
	if u := tr.filled[0].Used; u <= sp("158pt") || u > sp("160pt") {
		t.Errorf("region 1 Used %s, want the tenth line's box past 158pt", u)
	}
	cb, _ = newFlowBuilder(t, `p { font-family: serif; font-size: 10pt; line-height: 16pt }`)
	te, err := cb.HTMLToText(`<html><body><p>` + charLines("A", 2) + `</p></body></html>`)
	if err != nil {
		t.Fatal(err)
	}
	vl, err := cb.CreateVlist(te, sp("160pt"))
	if err != nil {
		t.Fatal(err)
	}
	var walk func(n node.Node)
	walk = func(n node.Node) {
		for ; n != nil; n = n.Next() {
			switch v := n.(type) {
			case *node.VList:
				walk(v.List)
			case *node.HList:
				if _, ok := v.Attributes[LineTrimEnd]; ok {
					t.Error("a line without text-box-trim records a trimmed leading")
				}
			}
		}
	}
	walk(vl)
}

// trimModel sets each line 13pt + 3pt and records 2.5pt of it as the leading
// below the text, as a line model with its own rule for the text edge may.
type trimModel struct{}

func (trimModel) LineBox(hl *node.HList, _ *node.LinebreakSettings) (bag.ScaledPoint, bag.ScaledPoint) {
	hl.SetAttribute(LineTrimEnd, bag.MustSP("2.5pt"))
	return bag.MustSP("13pt"), bag.MustSP("3pt")
}

func (trimModel) Leading(*node.HList, *node.LinebreakSettings) *node.Glue { return nil }

// A registered line model's LineTrimEnd is taken under text-box-trim:
// trim-end, and left out without it.
func TestTextBoxTrimEndFromALineModel(t *testing.T) {
	register := func(cb *CSSBuilder) {
		if err := cb.RegisterLineModel("trim", func(LineModelStyles) node.LineModel { return trimModel{} }); err != nil {
			t.Fatal(err)
		}
	}
	for _, c := range []struct {
		style string
		lines int
	}{
		{``, 9},
		{`text-box-trim: trim-end`, 10},
	} {
		cb, _ := newLineModelBuilder(t, `p { margin: 0; font-size: 10pt; -bag-leading-model: trim }`, register)
		te, err := cb.HTMLToText(`<html><body><p style="` + c.style + `">` + charLines("A", 14) + `</p></body></html>`)
		if err != nil {
			t.Fatal(err)
		}
		// Ten lines of 16pt need 160pt; the tenth one's text ends 2.5pt higher.
		tr := &testRegions{sizes: []Region{wide("158pt"), wide("1000pt")}}
		if err := cb.FlowText(te, tr); err != nil {
			t.Fatal(err)
		}
		if got := len(boxLines(tr.filled[0])); got != c.lines {
			t.Errorf("%q: region 1 holds %d lines, want %d", c.style, got, c.lines)
		}
	}
}
