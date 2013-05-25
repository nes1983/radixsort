package sort

import (
	"math/rand"
	"sort"
	"testing"
)

type data []uint64

func (d data) Len() int           { return len(d) }
func (d data) Swap(i, j int)      { d[i], d[j] = d[j], d[i] }
func (d data) Key(pos int) uint64 { return d[pos] }
func (d data) Less(i, j int) bool { return d[i] < d[j] }

func TestSortLarge_Random(t *testing.T) {
	// Test kind of stolen from Go source: sort/sort_test.go
	n := 100000000
	if testing.Short() {
		n /= 100
	}
	data := make(data, n)
	for i := 0; i < len(data); i++ {
		data[i] = uint64(rand.Int63n(1000))
	}

	if sort.IsSorted(data) {
		t.Fatalf("Terrible rand.rand.")
	}

	Sort(data)

	if !sort.IsSorted(data) {
		t.Error("Sort didn't sort 1M ints.")
	}
}

func TestLittleSort(t *testing.T) {
	data := data{100, 4, 1204, 4, 88, 1344, 1, 4942, 39, 23}
	Sort(data)
	if !sort.IsSorted(data) {
		t.Error("Wanted it sorted, but got ", data)
	}
}
