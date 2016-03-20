package binning

import "testing"

// Some example intervals with pre-calculated bin numbers.
// http://genomewiki.ucsc.edu/index.php/Bin_indexing_system
var intervalBins = []struct{ start, stop, bin int }{
	{0, 1, 585},
	{1<<29 - 1, 1 << 29, 4680},
	{0, 1 << 29, 0},
	{0, 1 << 17, 585},
	{1, 1 << 17, 585},
	{0, 1<<17 - 1, 585},
	{0, 1<<17 + 1, 73},
	{340, 74012, 585},
	{0, 1 << 18, 73},
	{74012, 173034, 73},
	{423427, 423428, 588},
	{100000, 200000, 73},
	{1000000, 2000000, 9},
	{10000000, 20000000, 1},
	{100000000, 200000000, 0},
	{200000, 1000000, 73},
	{2000000, 10000000, 1},
	{20000000, 100000000, 0},
	{300000000, 300000015, 2873},
	{300000000, 300100015, 359},
	{300000000, 300200015, 359},
	{300000000, 301000015, 44},
	{300000000, 311000015, 5},
	{300000000, 321000015, 5},
	{300000000, 381000015, 0},
	{300000000, 511000015, 0},
	{1200000, 2000000, 74},
}

var invalidIntervals = []struct{ start, stop int }{
	{-23442, -334},
	{-23442, 334},
	{-23442, 0},
	{-1, -1},
	{-1, 0},
	{5656, 1<<29 + 1},
	{-34234, 1<<29 + 3431},
}

var intervalRanges = []struct {
	start, stop int
	ranges      []struct{ start, stop int }
}{
	{0, 0, []struct{ start, stop int }{{585, 585}, {73, 73}, {9, 9}, {1, 1}, {0, 0}}},
	{0, 1, []struct{ start, stop int }{{585, 585}, {73, 73}, {9, 9}, {1, 1}, {0, 0}}},
	{1<<29 - 1, 1<<29 - 1, []struct{ start, stop int }{{4680, 4680}, {584, 584}, {72, 72}, {8, 8}, {0, 0}}},
	{1<<29 - 1, 1 << 29, []struct{ start, stop int }{{4680, 4680}, {584, 584}, {72, 72}, {8, 8}, {0, 0}}},
	{0, 1 << 29, []struct{ start, stop int }{{585, 4680}, {73, 584}, {9, 72}, {1, 8}, {0, 0}}},
	{0, 1<<17 + 1, []struct{ start, stop int }{{585, 586}, {73, 73}, {9, 9}, {1, 1}, {0, 0}}},
	{1200000, 2000000, []struct{ start, stop int }{{594, 600}, {74, 74}, {9, 9}, {1, 1}, {0, 0}}},
	{0, 1 << 18, []struct{ start, stop int }{{585, 586}, {73, 73}, {9, 9}, {1, 1}, {0, 0}}},
	{300000000, 300200015, []struct{ start, stop int }{{2873, 2875}, {359, 359}, {44, 44}, {5, 5}, {0, 0}}},
	{300000000, 301000015, []struct{ start, stop int }{{2873, 2881}, {359, 360}, {44, 44}, {5, 5}, {0, 0}}},
}

var intervalOverlappingBins = []struct {
	start, stop int
	bins        []int
}{
	{0, 1, []int{585, 73, 9, 1, 0}},
	{1<<29 - 1, 1 << 29, []int{4680, 584, 72, 8, 0}},
	{0, 1 << 29, conc(rng(585, 4681), rng(73, 585), rng(9, 73), rng(1, 9), []int{0})},
	{0, 1<<17 + 1, []int{585, 586, 73, 9, 1, 0}},
	{1200000, 2000000, append(rng(594, 601), 74, 9, 1, 0)},
	{0, 1 << 18, []int{585, 586, 73, 9, 1, 0}},
	{300000000, 300200015, []int{2873, 2874, 2875, 359, 44, 5, 0}},
	{300000000, 301000015, append(rng(2873, 2882), 359, 360, 44, 5, 0)},
}

var intervalContainingBins = []struct {
	start, stop int
	bins        []int
}{
	{0, 1, []int{585, 73, 9, 1, 0}},
	{1<<29 - 1, 1 << 29, []int{4680, 584, 72, 8, 0}},
	{0, 1 << 29, []int{0}},
	{0, 1<<17 + 1, []int{73, 9, 1, 0}},
	{1200000, 2000000, []int{74, 9, 1, 0}},
	{0, 1 << 18, []int{73, 9, 1, 0}},
	{300000000, 300200015, []int{359, 44, 5, 0}},
	{300000000, 301000015, []int{44, 5, 0}},
}

var intervalContainedBins = []struct {
	start, stop int
	bins        []int
}{
	{0, 1, []int{585}},
	{1<<29 - 1, 1 << 29, []int{4680}},
	{0, 1 << 29, conc(rng(585, 4681), rng(73, 585), rng(9, 73), rng(1, 9), []int{0})},
	{0, 1<<17 + 1, []int{585, 586, 73}},
	{1200000, 2000000, append(rng(594, 601), 74)},
	{0, 1 << 18, []int{585, 586, 73}},
	{300000000, 300200015, []int{2873, 2874, 2875, 359}},
	{300000000, 301000015, append(rng(2873, 2882), 359, 360, 44)},
}

func TestAssign(t *testing.T) {
	b := StandardBinning()
	for _, v := range intervalBins {
		if bin, error := b.Assign(v.start, v.stop); error != nil {
			t.Errorf("Assign(%d, %d) returned error: %v", v.start, v.stop, error)
		} else if bin != v.bin {
			t.Errorf("Assign(%d, %d) = %d, expected %d", v.start, v.stop, bin, v.bin)
		}
	}
}

func TestAssignInvalid(t *testing.T) {
	b := StandardBinning()
	for _, v := range invalidIntervals {
		if bin, error := b.Assign(v.start, v.stop); error == nil {
			t.Errorf("Assign(%d, %d) = %d, expected error", v.start, v.stop, bin)
		}
	}
}

func TestRanges(t *testing.T) {
	b := StandardBinning()
	for _, v := range intervalRanges {
		r, error := b.ranges(v.start, v.stop)
		if error != nil {
			t.Errorf("ranges(%d, %d) returned error: %v", v.start, v.stop, error)
			continue
		}
		for i, want := range v.ranges {
			if start, stop, ok := r(); !ok {
				t.Errorf("ranges(%d, %d)() call %d returned no value", v.start, v.stop, i+1)
			} else if start != want.start || stop != want.stop {
				t.Errorf("ranges(%d, %d)() call %d = (%d, %d), expected (%d, %d)",
					v.start, v.stop, i+1, start, stop, want.start, want.stop)
			}
		}
		if start, stop, ok := r(); ok {
			t.Errorf("ranges(%d, %d)() call %d = (%d, %d), expected no value",
				v.start, v.stop, len(v.ranges)+1, start, stop)
		}
	}
}

func TestRangesInvalid(t *testing.T) {
	b := StandardBinning()
	for _, v := range invalidIntervals {
		if r, error := b.ranges(v.start, v.stop); error == nil {
			t.Errorf("ranges(%d, %d) = %q, expected error", v.start, v.stop, r)
		}
	}
}

func TestOverlapping(t *testing.T) {
	b := StandardBinning()
	for _, v := range intervalOverlappingBins {
		bins, error := b.Overlapping(v.start, v.stop)
		if error != nil {
			t.Errorf("Overlapping(%d, %d) returned error: %v", v.start, v.stop, error)
			continue
		}
		if len(bins) != len(v.bins) {
			t.Errorf("len(Overlapping(%d, %d)) = %v, expected %v", v.start, v.stop, len(bins), len(v.bins))
			continue
		}
		for i := 0; i < len(bins); i++ {
			if bins[i] != v.bins[i] {
				t.Errorf("Overlapping(%d, %d)[%d] = %v, expected %v", v.start, v.stop, i, bins[i], v.bins[i])
				break
			}
		}
	}
}

func TestContaining(t *testing.T) {
	b := StandardBinning()
	for _, v := range intervalContainingBins {
		bins, error := b.Containing(v.start, v.stop)
		if error != nil {
			t.Errorf("Containing(%d, %d) returned error: %v", v.start, v.stop, error)
			continue
		}
		if len(bins) != len(v.bins) {
			t.Errorf("len(Containing(%d, %d)) = %v, expected %v", v.start, v.stop, len(bins), len(v.bins))
			continue
		}
		for i := 0; i < len(bins); i++ {
			if bins[i] != v.bins[i] {
				t.Errorf("Containing(%d, %d)[%d] = %v, expected %v", v.start, v.stop, i, bins[i], v.bins[i])
				break
			}
		}
	}
}

func TestContained(t *testing.T) {
	b := StandardBinning()
	for _, v := range intervalContainedBins {
		bins, error := b.Contained(v.start, v.stop)
		if error != nil {
			t.Errorf("Contained(%d, %d) returned error: %v", v.start, v.stop, error)
			continue
		}
		if len(bins) != len(v.bins) {
			t.Errorf("len(Contained(%d, %d)) = %v, expected %v", v.start, v.stop, len(bins), len(v.bins))
			continue
		}
		for i := 0; i < len(bins); i++ {
			if bins[i] != v.bins[i] {
				t.Errorf("Contained(%d, %d)[%d] = %v, expected %v", v.start, v.stop, i, bins[i], v.bins[i])
				break
			}
		}
	}
}

func TestAssignCovered(t *testing.T) {
	b := StandardBinning()
	for _, v := range intervalBins {
		bin, error := b.Assign(v.start, v.stop)
		if error != nil {
			t.Errorf("Assign(%d, %d) returned error: %v", v.start, v.stop, error)
			continue
		}
		if start, stop, error := b.Covered(bin); error != nil {
			t.Errorf("Covered(Assign(%d, %d)) returned error: %v", v.start, v.stop, error)
		} else if start > v.start || v.stop > stop {
			t.Errorf("Covered(Assign(%d, %d) = (%d, %d), expected (<=%d, >=%d)",
				v.start, v.stop, start, stop, v.start, v.stop)
		}
	}
}

// Concatenate any number of []int values.
func conc(args ...[]int) (r []int) {
	for _, arg := range args {
		r = append(r, arg...)
	}
	return
}

// Create []int with values from start (incl) to stop (excl).
func rng(start, stop int) []int {
	r := make([]int, stop-start)
	for i := 0; i < stop-start; i++ {
		r[i] = start + i
	}
	return r
}
