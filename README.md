binning: Interval binning for Go
================================

These are some utility functions for working with the interval binning scheme
as used in the
[UCSC Genome Browser](http://genome.cshlp.org/content/12/6/996.full).

This scheme can be used to implement fast overlap-based querying of intervals,
essentially mimicking an [R-tree](https://en.wikipedia.org/wiki/R-tree)
index.

Note that some database systems natively support spatial index methods such as
R-trees. See for example the [PostGIS](http://postgis.net) extension for
PostgreSQL.

Although in principle the method can be used for binning any kind of
intervals, be aware that the largest position supported by this implementation
is `2^29` (which covers the longest human chromosome).

```go
// Use the standard UCSC binning scheme.
b := binning.StandardBinning()

// Get the bin for some interval.
bin, error := b.Assign(74012, 173034)
if error != nil {
	log.Fatal("Binning.Assign error:", error)
}

fmt.Println(bin)
// Output: 73
```
