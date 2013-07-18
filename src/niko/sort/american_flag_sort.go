package sort

import "fmt"

const (
	SIZE      = 510
	THRESHOLD = 16
	ALPHABET  = 1 << 8
)

type Sorter []segment

type segment struct{ from, end, bytePos uint }

func getByte(data Interface, element, bytePos uint) byte {
	return byte((data.Key(int(element)) >> uint(56-bytePos*8)) & 0xff) // TODO Remove conversion
}

func (z *Sorter) rsorta(data Interface, from, end, bytePos uint) {
	if cap(*z) == 0 {
		*z = make([]segment, 0, SIZE)
	}

	// Piles[i] is the position of the element just *past* bucket i.
	*z = append(*z, segment{from, end, bytePos})

	for len(*z) != 0 {
		cur := (*z)[len(*z)-1]
		from, end, bytePos = cur.from, cur.end, cur.bytePos
		fmt.Println(cur)
		if bytePos > 7 {
			panic("Impossible.")
		}
		*z = (*z)[:len(*z)-1]
		//		if cur.to-cur.from < 31 {
		//			insertionSort(data, int(from), int(to))
		//			continue
		//		}

		// Compute counts.
		var nPiles int
		cmin := byte(255) // tally
		var counts [ALPHABET]uint
		for ak := from; ak < end; ak++ {
			c := getByte(data, ak, bytePos)
			counts[c]++
			if counts[c] == 1 {
				if c > 0 && c < cmin {
					cmin = c
				}
				nPiles++
			}
		}
		fmt.Println(counts)

		var piles [ALPHABET]uint
		// Compute piles and recurse
		piles[0] = from + counts[0]
		if counts[0] != 0 {
			nPiles--
			if counts[0] > 1 {
				*z = append(*z, segment{from, from + counts[0], bytePos + 1})
			}
		}
		for curPile, ak := cmin, from; nPiles > 0; curPile, nPiles = curPile+1, nPiles-1 {
			for counts[curPile] == 0 {
				curPile++
			}
			if counts[curPile] > 1 {
				*z = append(*z, segment{ak, ak + counts[curPile], bytePos + 1})
			}
			ak += counts[curPile]
			piles[curPile] = ak
		}
		fmt.Println(piles)

		// Permute home
		var c byte
		for ak := from; ak < end; ak += counts[c] {
			for {
				c = getByte(data, ak, bytePos)
				piles[c]--
				if piles[c] <= ak {
					break
				}
				data.Swap(int(piles[c]), int(ak)) // Breaks stability
			}
		}
	}
}
