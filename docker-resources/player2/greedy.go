/**
 * デフォルトの「特定の順番で行ける方向を探す」という戦略、これが意外に強い。
 * ただ、方向を探す順番によってパフォーマンスが変動する。
 * そこで、全ての順番の結果をシミュレートして最も良い結果を採用する。
 *
 * 長所: 簡単な割にスコアがいい。
 * 短所: 袋小路に弱い。盤面の変動を考慮してない。
 */
package main

var (
	/*
		greedyPolicies = [][]Ops{
			{DOWN, LEFT, RIGHT, UP}, {DOWN, LEFT, UP, RIGHT}, {DOWN, RIGHT, LEFT, UP}, {DOWN, RIGHT, UP, LEFT}, {DOWN, UP, LEFT, RIGHT}, {DOWN, UP, RIGHT, LEFT},
			{LEFT, DOWN, RIGHT, UP}, {LEFT, DOWN, UP, RIGHT}, {LEFT, RIGHT, DOWN, UP}, {LEFT, RIGHT, UP, DOWN}, {LEFT, UP, DOWN, RIGHT}, {LEFT, UP, RIGHT, DOWN},
			{RIGHT, DOWN, LEFT, UP}, {RIGHT, DOWN, UP, LEFT}, {RIGHT, LEFT, DOWN, UP}, {RIGHT, LEFT, UP, DOWN}, {RIGHT, UP, DOWN, LEFT}, {RIGHT, UP, LEFT, DOWN},
			{UP, DOWN, LEFT, RIGHT}, {UP, DOWN, RIGHT, LEFT}, {UP, LEFT, DOWN, RIGHT}, {UP, LEFT, RIGHT, DOWN}, {UP, RIGHT, DOWN, LEFT}, {UP, RIGHT, LEFT, DOWN},
		}
	*/

	// cyclicGreedyPolicies ポリシーのうち、探す順番が時計回りor半時計回りのもの
	// 完全ランダムだとポリシーの変更頻度が高すぎて蛇行が下手になるため、こちらを採用。。
	cyclicGreedyPolicies = [][]Ops{
		{DOWN, LEFT, UP, RIGHT}, {DOWN, RIGHT, UP, LEFT},
		{LEFT, DOWN, RIGHT, UP}, {LEFT, UP, RIGHT, DOWN},
		{RIGHT, DOWN, LEFT, UP}, {RIGHT, UP, LEFT, DOWN},
		{UP, LEFT, DOWN, RIGHT}, {UP, RIGHT, DOWN, LEFT},
	}
)

func Greedy(
	myHead Head,
	board [][]int,
	bumpGuard *BumpGuard,
) (Ops, int) {
	aggregator := NewScoreAggregator()
	simulator := NewGreedySimulator(aggregator, bumpGuard)
	for _, policy := range cyclicGreedyPolicies {
		simulator.simulate(myHead, board, policy)
	}

	maxOps, maxScore := aggregator.GetMaxScore()
	debug.SetScore("Greedy", maxScore)
	return maxOps, maxScore
}

type greedySimulator struct {
	aggregator *ScoreAggregator
	bumpGuard  *BumpGuard
}

func NewGreedySimulator(aggregator *ScoreAggregator, bumpGuard *BumpGuard) *greedySimulator {
	return &greedySimulator{
		aggregator: aggregator,
		bumpGuard:  bumpGuard,
	}
}

func (s *greedySimulator) simulate(
	myHead Head,
	board [][]int,
	policy []Ops,
) {
	tmpBoard := CopyBoard(board)

	cur := myHead
	score := 0
	firstOps := NONE
	for {
		way := NONE
		for _, ops := range policy {
			nxt := cur
			switch ops {
			case UP:
				nxt.Y--
			case DOWN:
				nxt.Y++
			case LEFT:
				nxt.X--
			case RIGHT:
				nxt.X++
			}
			if !IsOnBoard(nxt.X, nxt.Y) || tmpBoard[nxt.Y][nxt.X] != 0 {
				// 進めない箇所なら次の方向を試す
				continue
			}

			// 進めるので進む
			way = ops
			cur = nxt
			tmpBoard[cur.Y][cur.X] = cur.ID
			break
		}
		if way != NONE {
			score++
			if firstOps == NONE {
				firstOps = way // 1歩めの方向を記録
			}
			continue
		}
		// どの方向もだめなら終了
		break
	}

	if s.bumpGuard != nil {
		s.aggregator.SetScore(firstOps, s.bumpGuard.WithBumpGuard(firstOps, score))
	} else {
		s.aggregator.SetScore(firstOps, score)
	}
}
