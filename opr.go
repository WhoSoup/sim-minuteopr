package main

type MinuteOPR struct {
	Minimum uint64
	Chunks  []uint64
	Latest  uint64
}

func NewMinuteOPR(base uint64) *MinuteOPR {
	m := new(MinuteOPR)
	m.Minimum = base
	m.Latest = base
	return m
}

func (m *MinuteOPR) AddPOW(score uint64) {
	if score > m.Latest {
		m.Latest = score
	}
}

func (m *MinuteOPR) WantsMore() bool {
	return m.Latest < m.Minimum
}

func (m *MinuteOPR) Finish() {
	if len(m.Chunks) > 0 && m.Latest < m.Minimum {
		m.Minimum = m.Latest
	}
	m.Chunks = append(m.Chunks, m.Latest)
	m.Latest = 0
}
