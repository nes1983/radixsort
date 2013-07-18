package sort

import "fmt"

type Interface interface {
	Len() int
	Swap(i, j int)
	//	Key(pos, column int) (key uint64, err error) // Change key to allow smaller keys
	Key(pos int) uint64
}

// Insertion sort
func insertionSort(data Interface, a, b int) {
	for i := a + 1; i < b; i++ {
		for j := i; j > a && data.Key(j) < data.Key(j-1); j-- {
			data.Swap(j, j-1)
		}
	}
}

// Insertion sort in byte p only
func insertionSortInPos(data Interface, a, b int, bytePos uint) {
	for i := a + 1; i < b; i++ {
		for j := i; j > a && read(data.Key(j), bytePos) < read(data.Key(j-1), bytePos); j-- {
			data.Swap(j, j-1)
		}
	}
}

func Sort(data Interface) {
	sortRange(data, 0, data.Len()-1, uint(56), uint(0))
}

func sortRange(data Interface, a, b int, boundaryL, boundaryR uint) {
	if b <= a {
		return
	}

	fmt.Println("entry", a, b, boundaryL, boundaryR)
	fmt.Println("\n")
	for i := a; i <= b; i++ {
		fmt.Print(data.Key(i), " ")
	}
	fmt.Println("\n")

	// Steps:
	//    Count inversions. If very many: reverse. Count again. If very few: Selection sort.
	//    If not done:

	// 1. Count inversions in every byte simultaneously.
	// 2. Choose highest byte to sort by.
	// 3. Fill up buckets.
	// For every bucket:
	if false && b-a <= 8 { // Use this once it works.
		insertionSort(data, a, b)
	}

	var inversions int
	var shift uint
	var high bool
	high, boundaryL, boundaryR, inversions = countInversions(data, a, b, boundaryL, boundaryR)
	high = true // XXX remove
	fmt.Println("updated", inversions, high, boundaryL, boundaryR)

	if high {
		shift = boundaryL
	} else {
		shift = boundaryR
	}

	// TODO: Reverse carefully -- easy to break stability.

	if inversions == 0 {
		return
	}
	if false && inversions < 7 { // Use this as soon as it works.
		insertionSort(data, a, b)
		return
	}

	// TODO reverse or not. Then insertion sort or bucket sort.
	// TODO: If count is small, selection sort into place.

	bucketEnds := radixSortInByte(data, a, b, shift)

	fmt.Println("\n")
	for i := a; i <= b; i++ {
		fmt.Print(data.Key(i), " ")
	}
	fmt.Println("\n")

	if boundaryL == boundaryR {
		return
	}
	fmt.Println("old shifts: ", boundaryL, boundaryR, a, b)
	if !high {
		sortRange(data, a, b, boundaryL, boundaryR+8)
	} else {
		last := 0
		for _, v := range bucketEnds {
			sortRange(data, last, v-1, boundaryL-8, boundaryR)
			last = v
		}
	}
}

// Sets perBucket to the bucket sizes.
func computeBucketSizes(data Interface, a, b int, bytePos uint, capacity *[256]int) {
	for i := a; i <= b; i++ {
		capacity[read(data.Key(i), bytePos)]++
	}
}

// Read the bytePos least significant byte (0 is the very least significant).
func read(from uint64, bytePos uint) byte {
	return byte((from >> bytePos) & 0xff)
}

// After this, bucket i starts at writePos[i] and ends at writePos[i] + capacity[i]-1
func computeWritePos(max int, capacity, writePos *[256]int, a int) {
	writePos[0] = a
	for i := 1; true; i++ {
		p := writePos[i-1] + capacity[i-1]
		if p == max {
			break
		}
		writePos[i] = p
	}
}

// shiftDistance is in BITS
func radixSortInByte(data Interface, a, b int, shiftDistance uint) *[256]int {
	fmt.Println("radix", a, b, shiftDistance)
	var capacity [256]int
	var writePos [256]int

	computeBucketSizes(data, a, b, shiftDistance, &capacity)
	computeWritePos(b+1, &capacity, &writePos, a)

	buck := byte(0)
	for capacity[buck] == 0 {
		buck++
	}

	i := a
	for {
		targetBucket := read(data.Key(i), shiftDistance)
		capacity[targetBucket]--
		if targetBucket != buck {
			fmt.Println("swap ", i, writePos[targetBucket], data.Key(i), data.Key(writePos[targetBucket]))
			fmt.Println("masked: ", read(data.Key(i), shiftDistance))
			data.Swap(i, writePos[targetBucket])
		} else {
			i++
			for capacity[buck] == 0 { // Is this the end of the bucket?
				if i == b+1 { // All set
					writePos[targetBucket]++
					return &writePos
				}
				buck++
				i = writePos[buck]
			}
		}
		writePos[targetBucket]++
	}
}

const (
	L_IMPROV = uint64(0xffffffffffffff00)
	R_IMPROV = uint64(0x00ffffffffffffff)
)

// Find the leftmost and rightmost byte with inversions, as well as their inversions.
func countInversions(data Interface, a, b int, boundaryL, boundaryR uint) (
	big bool, lShift, rShift uint, inversions int) {

	lShift, rShift = boundaryR, boundaryL

	lImprovement, rImprovement := L_IMPROV<<lShift, R_IMPROV>>(56-rShift)
	var lInversions, rInversions int // One of these will be returned
	var lastL, lastR byte
	var last uint64
	for i := a; i <= b; i++ {
		cur := data.Key(i)
		fmt.Println("Improve ", last, cur, lImprovement&cur, lImprovement&last, lImprovement)
		for lShift < boundaryL && lImprovement&cur < lImprovement&last {
			lShift += 8
			lInversions = 0
			lastL = 0
			lImprovement = L_IMPROV << lShift
		}
		for rShift > boundaryR && rImprovement&cur < rImprovement&last {
			rShift -= 8
			rInversions = 0
			lastR = 0
			rImprovement = R_IMPROV >> (56 - rShift)
		}

		curL, curR := byte((cur>>lShift)&0xff), byte((cur>>rShift)&0xff)

		if curL < lastL {
			lInversions++
		}
		if curR < lastR {
			rInversions++
		}

		last, lastL, lastR = cur, curL, curR
	}

	if lInversions > rInversions/4 {
		return true, lShift, rShift, lInversions
	} else {
		return false, lShift, rShift, rInversions
	}
}


