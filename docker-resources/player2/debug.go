package main

import "sort"

var (
	DEBUG = false
)

type Debug struct {
	scores map[string]int
}

var debug = Debug{scores: make(map[string]int)}

func (d *Debug) Reset() {
	if !DEBUG {
		return
	}

	d.scores = make(map[string]int)
}

func (d *Debug) SetScore(strategy string, s int) {
	if !DEBUG {
		return
	}

	d.scores[strategy] = s
}

// PrintScores 各戦略のスコアをすべて出力
func (d *Debug) PrintScores() {
	if !DEBUG {
		return
	}

	print("[DEBUG] ")
	strategies := make([]string, 0, len(d.scores))
	for s := range d.scores {
		strategies = append(strategies, s)
	}
	sort.Strings(strategies)
	for _, strategy := range strategies {
		print(strategy, ": ", d.scores[strategy], " ")
	}
	println()
}

// PrintDecision 採択した戦略とスコアを出力
func (d *Debug) PrintDecision() {
	if !DEBUG {
		return
	}

	print("[DEBUG] ")
	maxStrategy := ""
	maxScore := 0
	for strategy, score := range d.scores {
		if score > maxScore {
			maxScore = score
			maxStrategy = strategy
		}
	}
	if maxScore > 0 {
		print("-> ", maxStrategy, ":", maxScore)
	} else {
		print("DEFEAT")
	}
	println()
}

func (d *Debug) PrintBumpGuard(bg *BumpGuard) {
	if !DEBUG {
		return
	}

	for _, ops := range []Ops{UP, LEFT, RIGHT, DOWN} {
		rate := bg.rate[ops]
		print("[DEBUG] BumpGuard: ", ops, ": ", rate, ": ")
		if rate != 0 {
			print("calc: ", int(bg.calcPenalty(rate)*100), "%")
		}
		println()
	}
}
