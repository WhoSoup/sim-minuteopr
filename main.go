package main

import (
	"flag"
	"fmt"
	"math/big"
	"math/rand"
	"sort"
	"time"
)

const MINUTES = 8

var rng *rand.Rand
var hashpower int
var threshold uint64 = 0xffff000000000000
var samples int
var compare bool

func init() {
	rng = rand.New(rand.NewSource(time.Now().UnixNano()))
}

type Rank struct {
	Power int
	Rank  uint64
}

func CompareHashpower() {
	var powers []int
	for i := 8192; i < 20000000; i *= 2 {
		powers = append(powers, i)
	}

	ranking := make([]Rank, 0)

	for _, power := range powers {
		opr := make([]uint64, MINUTES)

		for min := 0; min < MINUTES; min++ {
			for i := 0; i < power; i++ {
				pow := rng.Uint64()
				if pow > opr[min] {
					opr[min] = pow
				}
			}
		}

		min := opr[0]
		for _, pow := range opr {
			if pow < min {
				min = pow
			}
		}

		ranking = append(ranking, Rank{Power: power, Rank: min})
	}

	sort.Slice(ranking, func(i, j int) bool { return ranking[j].Rank < ranking[i].Rank })

	for _, r := range ranking {
		fmt.Printf("%d ", r.Power)
	}
	fmt.Println()
}

func CreateBases(limit int) []*MinuteOPR {
	results := make(TopX, limit)
	for h := 0; h < hashpower; h++ {
		results.Add(rng.Uint64())
	}

	res := make([]*MinuteOPR, limit)
	for k, v := range results {
		res[k] = NewMinuteOPR(v)
		res[k].Finish()
	}

	return res
}

func TryStrategy(name string, f func() []*MinuteOPR) {
	sum := big.NewInt(0)

	if !compare {
		fmt.Println(name, "\n=====================")
	}

	var best *MinuteOPR
	found := 0
	for i := 0; i < samples; i++ {
		bases := f()
		var submitted []*MinuteOPR
		for _, b := range bases {
			if b.Minimum > threshold {
				a1 := new(big.Int).SetUint64(b.Minimum)
				sum.Add(sum, a1)
				found++
				submitted = append(submitted, b)
			}
			if best == nil || b.Minimum > best.Minimum {
				best = b
			}
		}

		if !compare && len(submitted) > 0 {
			fmt.Printf("%d", i)
			for _, b := range submitted {
				fmt.Printf(";%d", b.Minimum-threshold)
			}
			fmt.Printf("\n")
		}
	}

	if compare {
		fmt.Println(name, found)
	} else if found == 0 {
		fmt.Println(name)
		fmt.Println("No OPRs above threshold found")
		return
	}
}

func main() {
	hashpowerF := flag.Int("hashpower", 1000000, "the simulated amount of hashpower, in hashes/chunk")
	samplesF := flag.Int("samples", 1000, "number of samples")
	hpCompare := flag.Bool("compare", false, "compare hash power for one strategy")
	flag.Parse()
	hashpower = *hashpowerF
	samples = *samplesF
	compare = *hpCompare

	if compare {
		for i := 0; i < samples; i++ {
			CompareHashpower()
		}
	} else {
		//TryStrategy("Old Mining", OldMining)
		TryStrategy("Strategy One (1)", func() []*MinuteOPR { return StrategyOne(threshold, 1) })
		TryStrategy("Strategy One (2)", func() []*MinuteOPR { return StrategyOne(threshold, 2) })
		TryStrategy("Strategy One (4)", func() []*MinuteOPR { return StrategyOne(threshold, 4) })
		TryStrategy("Strategy One (8)", func() []*MinuteOPR { return StrategyOne(threshold, 8) })
		TryStrategy("Strategy One (16)", func() []*MinuteOPR { return StrategyOne(threshold, 16) })
		TryStrategy("Strategy Two", StrategyTwo)
		TryStrategy("Strategy Three (1)", func() []*MinuteOPR { return StrategyThree(1) })
		TryStrategy("Strategy Three (2)", func() []*MinuteOPR { return StrategyThree(2) })
		TryStrategy("Strategy Three (4)", func() []*MinuteOPR { return StrategyThree(4) })
		TryStrategy("Strategy Three (8)", func() []*MinuteOPR { return StrategyThree(8) })
		TryStrategy("Strategy Three (16)", func() []*MinuteOPR { return StrategyThree(16) })
	}
}

func OldMining() []*MinuteOPR {
	results := make(TopX, 16)

	hashes := hashpower
	for hashes > 0 {
		results.Add(rng.Uint64())
		hashes--
	}

	oprs := make([]*MinuteOPR, 0)
	for _, r := range results {
		if r >= threshold {
			oprs = append(oprs, &MinuteOPR{Minimum: r})
		}
	}
	return oprs
}

func StrategyOne(threshold uint64, limit int) []*MinuteOPR {
	var bases []*MinuteOPR
	for h := 0; h < hashpower; h++ {
		r := rng.Uint64()
		if r > threshold {
			bases = append(bases, NewMinuteOPR(r))
		}
	}

	if len(bases) == 0 {
		//fmt.Println("Strategy One: No bases found above treshold")
		return nil
	}

	sort.Slice(bases, func(i, j int) bool {
		return bases[j].Minimum < bases[i].Minimum
	})

	if len(bases) > limit {
		bases = bases[:limit]
	}

	for b := range bases {
		bases[b].Finish()
	}

	//start := len(bases)

	for m := 1; m < MINUTES; m++ {
		hashes := hashpower
		baseID := 0

		// first phase, get all above treshold
		for hashes > 0 && baseID < len(bases) {
			r := rng.Uint64()
			hashes--
			bases[baseID].AddPOW(r)
			if bases[baseID].Latest > threshold {
				baseID++
			}
		}

		if baseID < len(bases) {
			bases = bases[:baseID+1]
		}

		baseID = 0

		// second phase, improve pow
		for hashes > 0 && baseID < len(bases) {
			r := rng.Uint64()
			hashes--
			bases[baseID].AddPOW(r)

			if bases[baseID].Latest >= bases[baseID].Minimum {
				baseID++
			}
		}

		if len(bases) == 0 {
			//fmt.Println("Strategy One ended up with no viable OPRs")
			return nil
		}

		for b := range bases {
			bases[b].Finish()
		}

		sort.Slice(bases, func(i, j int) bool {
			return bases[j].Minimum < bases[i].Minimum
		})
	}

	return bases
}

func StrategyTwo() []*MinuteOPR {
	bases := CreateBases(16)

	for m := 1; m < MINUTES; m++ {
		hashes := hashpower
		baseID := 0

		// get all above minimum
		for hashes > 0 && baseID < len(bases) {
			r := rng.Uint64()
			hashes--
			bases[baseID].AddPOW(r)
			if bases[baseID].Latest > bases[baseID].Minimum {
				baseID++
			}
		}

		if baseID < len(bases) {
			bases = bases[:baseID+1]
		}

		for k := range bases {
			bases[k].Finish()
		}
	}

	return bases
}

func StrategyThree(amount int) []*MinuteOPR {
	bases := CreateBases(amount)
	for m := 1; m < MINUTES; m++ {
		hashes := hashpower

		base := 0
		done := make(map[int]bool)
		for hashes > 0 && len(done) < len(bases) {
			if bases[base].Latest < bases[base].Chunks[0] {
				bases[base].AddPOW(rng.Uint64())
				hashes--
			} else {
				// help best first
				for b := range bases {
					if bases[b].Latest < bases[b].Chunks[0] {
						bases[b].AddPOW(rng.Uint64())
						hashes--
						break
					}
				}

				done[base] = true
			}
			base++
			base = base % len(bases)
		}

		for hashes > 0 {
			bases[base].AddPOW(rng.Uint64())
			hashes--
			base++
			base = base % len(bases)
		}

		for k := range bases {
			bases[k].Finish()
		}

		sort.Slice(bases, func(i, j int) bool {
			return bases[i].Minimum > bases[j].Minimum
		})
	}

	return bases
}
