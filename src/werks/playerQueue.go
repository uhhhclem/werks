package werks


// PlayerInfo is a container for putting players into
// the PlayerQueue.
type PlayerInfo struct {
	player *Player 	// the player we're keeping track of
	turnOrder int		// the turn order in the previous turn
	index int				// the index of the item in the heap
}

// PlayerQueue is a priority queue that orders players by
// money and previous turn order.  To do this, it implements
// heap.Interface.
type PlayerQueue []*PlayerInfo

func (pq PlayerQueue) Len() int {
	return len(pq)
}

func (pq PlayerQueue) Less(i, j int) bool {
	// Pop removes the "least" item in the heap, so we want
	// that to be the player with the highest player order.
	priority := func(p *PlayerInfo) int {
		// Subtract turn order so that going 4th is lower
		// than going 3rd.
		return (10 * p.player.Money) - p.turnOrder
	}

	pi := priority(pq[i])
	pj := priority(pq[j])
	return pi > pj
}

func (pq PlayerQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *PlayerQueue) Push(x interface{}) {
	// This is copied straight out of the example in containers/heap.
	a := *pq
	n := len(a)
	a = a[0 : n+1]
	item := x.(*PlayerInfo)
	item.index = n
	a[n] = item
	*pq = a
}

func (pq *PlayerQueue) Pop() interface{} {
	// This is copied straight out of the example in containers/heap.
	a := *pq
	n := len(a)
	item := a[n-1]
	item.index = -1 // for safety
	*pq = a[0 : n-1]
	return item
}
