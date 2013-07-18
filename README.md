radixsort
=========

General purpose Radix sort in go.

General purpose means that you can sort anything, not just int arrays.

You're supposed to use it like this:

	type cell struct {
		rowKey, colKey uint
	}
	
	func (s []cell) SortKey(pos, priority int) (key int, ok bool) {
		switch priority {
			case 0:
				return s[pos].rowKey, true
			case 1:
				return s[pos].colKey, true
			default:
				return 0, false
		}
	}
	
	func (s []cell) Len() { return len(s) }
	func (s []cell) Swap(i, j int) { s[i, j] = s[j, i] }
	
	var cells []cell
	cells = ...
	
	radix.Sort(cells)
	
This should outperform the standard sort package, and does in my benchmarks.
Alas, it isn't quite done.