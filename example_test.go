package binning_test

import (
	"fmt"
	"log"

	"github.com/martijnvermaat/binning"
)

// This example shows the bin for some interval.
func ExampleBinning_Assign() {
	// Use the standard UCSC binning scheme.
	b := binning.StandardBinning()

	bin, error := b.Assign(74012, 173034)
	if error != nil {
		log.Fatal("Binning.Assign error:", error)
	}

	fmt.Println(bin)
	// Output: 73
}

// This example shows the bins for all intervals overlapping some interval by
// at least one position.
func ExampleBinning_Overlapping() {
	// Use the standard UCSC binning scheme.
	b := binning.StandardBinning()

	bins, error := b.Overlapping(73192, 78018)
	if error != nil {
		log.Fatal("Binning.Overlapping:", error)
	}

	fmt.Println(bins)
	// Output: [585 73 9 1 0]
}
