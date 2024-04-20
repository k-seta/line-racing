/**
 * Greedyの改良版。行き詰まったらBacktrackさせることで袋小路対策する。
 * Greedy同様、Backtrack時の探索方向の順番によってパフォーマンスが変動する。
 * そこで、全ての順番の結果をシミュレートして最も良い結果を採用する。
 * ...正直Greedyは不要でこっちだけでよさそう。
 *
 * 長所: 袋小路に負けない。割と盤面で取れる最長経路を探せる。
 * 短所: 盤面の変動を考慮してない。
 */
package main

func GreedyBT(
	myHead Head,
	board [][]int,
	bumpGuard *BumpGuard,
) (Ops, int) {
	aggregator := NewScoreAggregator()
	simulator := NewGreedyBTSimulator(aggregator, bumpGuard)
	for _, policy := range cyclicGreedyPolicies {
		simulator.simulate(myHead, board, policy)
	}

	maxOps, maxScore := aggregator.GetMaxScore()
	debug.SetScore("GreedyBT", maxScore)
	return maxOps, maxScore
}

type greedyBTSimulator struct {
	aggregator *ScoreAggregator
	bumpGuard  *BumpGuard
}

func NewGreedyBTSimulator(aggregator *ScoreAggregator, bumpGuard *BumpGuard) *greedyBTSimulator {
	return &greedyBTSimulator{
		aggregator: aggregator,
		bumpGuard:  bumpGuard,
	}
}

func (s *greedyBTSimulator) simulate(
	myHead Head,
	board [][]int,
	policy []Ops,
) {
	tmpBoard := CopyBoard(board)
	s.backtrack(myHead, 0, NONE, tmpBoard, policy)
}

func (s *greedyBTSimulator) backtrack(cur Head, prevScore int, prevOps Ops, board [][]int, policy []Ops) {
	board[cur.Y][cur.X] = cur.ID /// boardを踏む
	if s.bumpGuard != nil {
		s.aggregator.SetScore(prevOps, s.bumpGuard.WithBumpGuard(prevOps, prevScore))
	} else {
		s.aggregator.SetScore(prevOps, prevScore)
	}

	for _, ops := range policy {
		switch ops {
		case UP:
			if IsOnBoard(cur.X, cur.Y-1) && board[cur.Y-1][cur.X] == 0 {
				nxt := Head{ID: cur.ID, X: cur.X, Y: cur.Y - 1}
				ops := prevOps
				if ops == NONE {
					ops = UP
				}
				s.backtrack(nxt, prevScore+1, ops, board, policy)
			}

		case DOWN:
			if IsOnBoard(cur.X, cur.Y+1) && board[cur.Y+1][cur.X] == 0 {
				nxt := Head{ID: cur.ID, X: cur.X, Y: cur.Y + 1}
				ops := prevOps
				if ops == NONE {
					ops = DOWN
				}
				s.backtrack(nxt, prevScore+1, ops, board, policy)
			}

		case LEFT:
			if IsOnBoard(cur.X-1, cur.Y) && board[cur.Y][cur.X-1] == 0 {
				nxt := Head{ID: cur.ID, X: cur.X - 1, Y: cur.Y}
				ops := prevOps
				if ops == NONE {
					ops = LEFT
				}
				s.backtrack(nxt, prevScore+1, ops, board, policy)
			}

		case RIGHT:
			if IsOnBoard(cur.X+1, cur.Y) && board[cur.Y][cur.X+1] == 0 {
				nxt := Head{ID: cur.ID, X: cur.X + 1, Y: cur.Y}
				ops := prevOps
				if ops == NONE {
					ops = RIGHT
				}
				s.backtrack(nxt, prevScore+1, ops, board, policy)
			}
		}
	}
}
