package htmlbag

import "testing"

// A block too tall for an empty region is placed in it anyway, as on an empty
// page. Below content the caller placed itself, the region is not empty, and
// the block moves on to the next region.
func TestFlowTextOccupiedRegionMovesABlockOn(t *testing.T) {
	body := `<p style="line-height: 30pt">Aq</p>`
	for _, c := range []struct {
		occupied bool
		want     []bool // whether each region holds the block
	}{
		{false, []bool{true}},
		{true, []bool{false, true}},
	} {
		cb, _ := newFlowBuilder(t, "")
		first := wide("20pt")
		first.Occupied = c.occupied
		tr := flow(t, cb, body, first, wide("1000pt"))
		if len(tr.filled) != len(c.want) {
			t.Fatalf("occupied %v: filled %d regions, want %d", c.occupied, len(tr.filled), len(c.want))
		}
		for i, f := range tr.filled {
			if got := f.Used > 0; got != c.want[i] {
				t.Errorf("occupied %v: region %d holds the block: %v, want %v", c.occupied, i+1, got, c.want[i])
			}
		}
	}
}
