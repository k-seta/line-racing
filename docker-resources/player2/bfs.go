/**
 * 幅優先探索を用いた戦略。
 * 現在地から最も遠いマスまで最短経路で移動するよう方向選択する。
 *
 * 長所: 遠くに移動できる。
 * 短所: 最短経路を取るので、マス数を稼げない。盤面の変動を考慮してない。
 */
package main

func BFS(
	myHead Head,
	board [][]int,
	bumpGuard *BumpGuard,
) (Ops, int) {
	type Coord struct {
		X, Y, Score int
		Ops         Ops
	}

	tmpBoard := CopyBoard(board)
	queue := make([]Coord, 0, 30*20)
	aggregator := NewScoreAggregator()

	queue = append(queue, Coord{X: myHead.X, Y: myHead.Y, Score: 0, Ops: NONE})
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]

		// 左 (x -> x-1)
		if IsOnBoard(cur.X-1, cur.Y) && tmpBoard[cur.Y][cur.X-1] == 0 {
			nxt := Coord{X: cur.X - 1, Y: cur.Y, Score: cur.Score + 1, Ops: cur.Ops}
			if nxt.Score == 1 {
				nxt.Ops = LEFT
			}
			tmpBoard[nxt.Y][nxt.X] = myHead.ID
			queue = append(queue, nxt)
			if bumpGuard != nil {
				aggregator.SetScore(nxt.Ops, bumpGuard.WithBumpGuard(nxt.Ops, nxt.Score))
			} else {
				aggregator.SetScore(nxt.Ops, nxt.Score)
			}
		}

		// 右 (x -> x+1)
		if IsOnBoard(cur.X+1, cur.Y) && tmpBoard[cur.Y][cur.X+1] == 0 {
			nxt := Coord{X: cur.X + 1, Y: cur.Y, Score: cur.Score + 1, Ops: cur.Ops}
			if nxt.Score == 1 {
				nxt.Ops = RIGHT
			}
			tmpBoard[nxt.Y][nxt.X] = myHead.ID
			queue = append(queue, nxt)
			if bumpGuard != nil {
				aggregator.SetScore(nxt.Ops, bumpGuard.WithBumpGuard(nxt.Ops, nxt.Score))
			} else {
				aggregator.SetScore(nxt.Ops, nxt.Score)
			}
		}

		// 上 (y -> y-1)
		if IsOnBoard(cur.X, cur.Y-1) && tmpBoard[cur.Y-1][cur.X] == 0 {
			nxt := Coord{X: cur.X, Y: cur.Y - 1, Score: cur.Score + 1, Ops: cur.Ops}
			if nxt.Score == 1 {
				nxt.Ops = UP
			}
			tmpBoard[nxt.Y][nxt.X] = myHead.ID
			queue = append(queue, nxt)
			if bumpGuard != nil {
				aggregator.SetScore(nxt.Ops, bumpGuard.WithBumpGuard(nxt.Ops, nxt.Score))
			} else {
				aggregator.SetScore(nxt.Ops, nxt.Score)
			}
		}

		// 下 (y -> y+1)
		if IsOnBoard(cur.X, cur.Y+1) && tmpBoard[cur.Y+1][cur.X] == 0 {
			nxt := Coord{X: cur.X, Y: cur.Y + 1, Score: cur.Score + 1, Ops: cur.Ops}
			if nxt.Score == 1 {
				nxt.Ops = DOWN
			}
			tmpBoard[nxt.Y][nxt.X] = myHead.ID
			queue = append(queue, nxt)
			if bumpGuard != nil {
				aggregator.SetScore(nxt.Ops, bumpGuard.WithBumpGuard(nxt.Ops, nxt.Score))
			} else {
				aggregator.SetScore(nxt.Ops, nxt.Score)
			}
		}
	}

	maxOps, maxScore := aggregator.GetMaxScore()
	debug.SetScore("BFS", maxScore)
	return maxOps, maxScore
}
