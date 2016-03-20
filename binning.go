// Package binning implements the interval binning scheme.
//
// These are some utility functions for working with the interval binning
// scheme as used in the UCSC Genome Browser: http://genome.cshlp.org/content/12/6/996.full
//
// This scheme can be used to implement fast overlap-based querying of
// intervals, essentially mimicking an R-tree index: https://en.wikipedia.org/wiki/R-tree
//
// Note that some database systems natively support spatial index methods such
// as R-trees. See for example the PostGIS extension for PostgreSQL: http://postgis.net
//
// Although in principle the method can be used for binning any kind of
// intervals, be aware that the largest position supported by this
// implementation is 2^29 (which covers the longest human chromosome).
//
// All positions and ranges in this package are zero-based and open-ended,
// following standard Go indexing and slicing notation.
package binning

import (
	"errors"
	"fmt"
)

// A Binning implements a specific interval binning scheme.
type Binning struct {
	MaxPosition int
	MaxBin      int

	binOffsets []int
	shiftFirst uint
	shiftNext  uint
}

// The closure created by ranges for the interval start:stop returns the first
// and last bin overlapping the interval for each level, starting with the
// smallest bins.
// Algorithm by Jim Kent: http://genomewiki.ucsc.edu/index.php/Bin_indexing_system
func (b Binning) ranges(start, stop int) (func() (int, int, bool), error) {
	if start < 0 || stop > b.MaxPosition+1 {
		return nil, errors.New(fmt.Sprintf("interval out of range: %d-%d (maximum position is %d)", start, stop, b.MaxPosition))
	}
	if stop <= start {
		stop = start + 1
	}

	startBin := start >> b.shiftFirst
	stopBin := (stop - 1) >> b.shiftFirst
	maxLevel := len(b.binOffsets) - 1
	level := 0

	return func() (int, int, bool) {
		if level > maxLevel {
			return 0, 0, false
		}
		if level > 0 {
			startBin >>= b.shiftNext
			stopBin >>= b.shiftNext
		}
		level++
		return b.binOffsets[level-1] + startBin, b.binOffsets[level-1] + stopBin, true
	}, nil
}

// Assign returns the smallest bin fitting the interval start:stop.
func (b Binning) Assign(start, stop int) (int, error) {
	nextRange, err := b.ranges(start, stop)
	if err != nil {
		return 0, err
	}

	for {
		startBin, stopBin, ok := nextRange()
		if !ok {
			break
		}
		if startBin == stopBin {
			return startBin, nil
		}
	}

	panic("unexpected loop fall-through")
}

// Overlapping returns bins for all intervals overlapping the interval
// start:stop by at least one position.
func (b Binning) Overlapping(start, stop int) ([]int, error) {
	nextRange, err := b.ranges(start, stop)
	if err != nil {
		return nil, err
	}

	bins := []int{}

	for {
		startBin, stopBin, ok := nextRange()
		if !ok {
			break
		}
		tmp := make([]int, stopBin-startBin+1)
		for bin := startBin; bin <= stopBin; bin++ {
			tmp[bin-startBin] = bin
		}
		bins = append(bins, tmp...)
	}

	return bins, nil
}

// Containing returns bins for all intervals completely containing the
// interval start:stop.
func (b Binning) Containing(start, stop int) ([]int, error) {
	maxBin, err := b.Assign(start, stop)
	if err != nil {
		return nil, err
	}

	overlapping, err := b.Overlapping(start, stop)
	if err != nil {
		return nil, err
	}

	bins := overlapping[:0]
	for _, bin := range overlapping {
		if bin <= maxBin {
			bins = append(bins, bin)
		}
	}

	return bins, nil
}

// Contained returns bins for all intervals completely contained by the
// interval start:stop.
func (b Binning) Contained(start, stop int) ([]int, error) {
	minBin, err := b.Assign(start, stop)
	if err != nil {
		return nil, err
	}

	overlapping, err := b.Overlapping(start, stop)
	if err != nil {
		return nil, err
	}

	bins := overlapping[:0]
	for _, bin := range overlapping {
		if bin >= minBin {
			bins = append(bins, bin)
		}
	}

	return bins, nil
}

// Covered returns the interval covered by bin.
func (b Binning) Covered(bin int) (int, int, error) {
	if bin < 0 || bin > b.MaxBin {
		return 0, 0, errors.New(fmt.Sprintf("not a valid bin number: %d (must be >= 0 and <= %d)", bin, b.MaxBin))
	}

	shift := b.shiftFirst
	for _, offset := range b.binOffsets {
		if offset <= bin {
			return (bin - offset) << shift, (bin + 1 - offset) << shift, nil
		}
		shift += b.shiftNext
	}

	panic("unexpected loop fall-through")
}

// NewBinning creates a new binning scheme with maxPosition the maximum
// position that can be binned, binOffsets the first bin number per level,
// shiftFirst how much to shift to get to the smallest bin, and shiftNext how
// much to shift to get to the next larger bin.
func NewBinning(maxPosition int, binOffsets []int, shiftFirst, shiftNext uint) Binning {
	return Binning{
		MaxPosition: maxPosition,
		MaxBin:      binOffsets[0] + (maxPosition >> shiftFirst),
		binOffsets:  binOffsets,
		shiftFirst:  shiftFirst,
		shiftNext:   shiftNext,
	}
}

// StandardBinning returns the standard binning scheme used by the UCSC Genome
// Browser covering positions >= 0 and <= 2^29-1.
// http://genomewiki.ucsc.edu/index.php/Bin_indexing_system
func StandardBinning() Binning {
	return NewBinning(1<<29-1, []int{512 + 64 + 8 + 1, 64 + 8 + 1, 8 + 1, 1, 0}, 17, 3)
}
