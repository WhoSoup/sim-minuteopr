package main

type TopX []uint64

func (t *TopX) Add(n uint64) {
	l := len(*t)
	for k, v := range *t {
		if n > v {
			*t = append((*t)[:k], append(TopX{n}, (*t)[k:]...)...)
			break
		}
	}
	*t = (*t)[:l]
}
