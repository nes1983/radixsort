package sort

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

const (
	lImprov = uint64(0xffffffffffffff00)
	rImprov = uint64(0x00ffffffffffffff)
)

func Sort(data Interface) {
	sortRange(data, 0, data.Len()-1, uint(0), uint(56))
}

func sortRange(data Interface, a, b int, lShift, rShift uint) {
	if b <= a {
		return
	}
	//	fmt.Println("entry", a, b, lShift, rShift)

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

	var shiftDistance uint
	high, lShift, rShift, inversions := countInversions(data, a, b, lShift, rShift)
	//	fmt.Println("updated", inversions, high, lShift, rShift)

	if high {
		shiftDistance = lShift
	} else {
		shiftDistance = rShift
	}

	// TODO: Reverse carefully -- easy to break stability.

	if inversions == 0 {
		return
	}
	if false && inversions < 7 { // Use this once it works.
		insertionSort(data, a, b)
		return
	}

	// TODO reverse or not. Then insertion sort or bucket sort.
	// TODO: If count is small, selection sort into place.

	perBucket := radixSortInByte(data, a, b, shiftDistance)

	if lShift == rShift {
		return
	}
	//	fmt.Println("old shifts: ", lShift, rShift, a, b)
	if !high {
		sortRange(data, a, b, lShift, rShift+8)
	} else {
		last := 0
		for _, v := range perBucket {
			sortRange(data, last, v-1, lShift-8, rShift)
			last = v
		}
	}
}

// Sets perBucket to the bucket sizes.
func computeBucketSizes(data Interface, a, b int, bytePos uint, perBucket *[256]int) {
	for i := a; i <= b; i++ {
		perBucket[read(data.Key(i), bytePos)]++
	}
}

// Read the bytePos least significant byte (0 is the very least significant).
func read(from uint64, bytePos uint) byte {
	return byte((from >> bytePos) & 0xff)
}

// Assumes that s.perBucket is ok, and then modifies it and sets s.WritePos
// After this, bucket i starts at s.WritePos[i] and ends at s.PerBucket[i]-1
func computeWritePos(max int, perBucket, writePos *[256]int, a int) {
	for i := range writePos {
		writePos[i] = a
	} // Reset writePos

	sum := perBucket[0] + a
	perBucket[0] = sum
	for i := 1; sum-a != max; i++ {
		writePos[i] = sum
		sum += perBucket[i]
		perBucket[i] = sum
	}
}

// shiftDistance is in BITS
func radixSortInByte(data Interface, a, b int, shiftDistance uint) *[256]int {
	//	fmt.Println("radix", a, b, shiftDistance)
	var perBucket [256]int
	var writePos [256]int

	computeBucketSizes(data, a, b, shiftDistance, &perBucket)
	computeWritePos(b-a+1, &perBucket, &writePos, a)

	buck := byte(0)
	for perBucket[buck] == a {
		buck++
	}

	i := a
	for {
		targetBucket := read(data.Key(i), shiftDistance)
		if targetBucket != buck {
			target := writePos[targetBucket]
			data.Swap(i, target)
		} else {
			i++ // Could move over into next bucket.
			for i == perBucket[buck] {
				// Bucket exhausted
				if i == b+1 {
					return &perBucket
				}
				buck++
				i = writePos[buck]
			}
		}
		writePos[targetBucket]++
	}
}

// Find the leftmost and rightmost byte with inversions, as well as their inversions.
func countInversions(data Interface, a, b int, lShift, rShift uint) (
	big bool, newLShift, newRShift uint, inversions int) {

	lImprovement, rImprovement := lImprov<<lShift, rImprov>>(56-rShift)

	mask := uint64(0xffffffffffffffff)
	if lShift >= rShift {
		mask = mask >> rShift << rShift << (56 - lShift) >> (56 - lShift)
	}

	last := data.Key(a) & mask
	var cur uint64
	i := a + 1
	for ; i <= b; i++ {
		cur = data.Key(i) & mask
		if cur != last {
			break
		}
		last = cur
	}
	if cur == last { // All the same
		return true, 0, 0, 0
	}

	k := cur ^ last

	// Duplicate
	for ; lImprovement&k != 0; lImprovement = lImprov << lShift {
		lShift += 8
	}
	for ; rImprovement&k != 0; rImprovement = rImprov >> (56 - rShift) {
		rShift -= 8
	}

	lastL, lastR := byte((cur>>lShift)&0xff), byte((cur>>rShift)&0xff)

	lInversions, rInversions := 0, 0 // One of these will be returned

	for ; i <= b; i++ {
		cur = data.Key(i) & mask
		k = cur ^ last

		// Attempt to shift
		for ; lImprovement&k != 0; lImprovement = lImprov << lShift {
			lShift += 8
			lInversions = 0
			lastL = 0
		}
		for ; rImprovement&k != 0; rImprovement = rImprov >> (56 - rShift) {
			rShift -= 8
			rInversions = 0
			lastR = 0
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
