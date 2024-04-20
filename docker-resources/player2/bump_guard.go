/**
 * 各戦略は、盤面の変動を考慮していない。
 * 盤面を変動させるのは他Headなので、他Headから離れておけば変動を考慮しないことによるデメリットを減らせる。
 * 特に、他Headと同じマスに突っ込んで死ぬのは最悪。
 */
package main

import "math"

// BoardBumpGuard ボードを細工して衝突回避する。
// 他Headの位置から+1マス前後左右を壁にする。
func BoardBumpGuard(board [][]int, myHead Head, heads []Head) [][]int {
	newBoard := CopyBoard(board)

	for _, head := range heads {
		if head.ID == myHead.ID {
			continue
		}
		x, y := head.X, head.Y
		// +-1マスを壁にする
		if IsOnBoard(x-1, y) {
			newBoard[y][x-1] = head.ID
		}
		if IsOnBoard(x+1, y) {
			newBoard[y][x+1] = head.ID
		}
		if IsOnBoard(x, y-1) {
			newBoard[y-1][x] = head.ID
		}
		if IsOnBoard(x, y+1) {
			newBoard[y+1][x] = head.ID
		}
	}
	return newBoard
}

// BumpGuard 各Opsに対するペナルティを管理
type BumpGuard struct {
	rate map[Ops]int
}

func NewBumpGuard(myHead Head, otherHeads []Head, board [][]int) *BumpGuard {
	return &BumpGuard{
		rate: map[Ops]int{
			LEFT:  GetNearestHeadsDistance(Head{ID: myHead.ID, X: myHead.X - 1, Y: myHead.Y}, otherHeads, board),
			RIGHT: GetNearestHeadsDistance(Head{ID: myHead.ID, X: myHead.X + 1, Y: myHead.Y}, otherHeads, board),
			UP:    GetNearestHeadsDistance(Head{ID: myHead.ID, X: myHead.X, Y: myHead.Y - 1}, otherHeads, board),
			DOWN:  GetNearestHeadsDistance(Head{ID: myHead.ID, X: myHead.X, Y: myHead.Y + 1}, otherHeads, board),
		},
	}
}

// GetNearestHeadsDistance 他Headとの最短距離を取得
func GetNearestHeadsDistance(myHead Head, otherHeads []Head, board [][]int) int {
	if IsOnBoard(myHead.X, myHead.Y) && board[myHead.Y][myHead.X] != 0 {
		return 0
	}

	// ボード複製 & 他Headの位置を別の色に塗る
	tmpBoard := CopyBoard(board)
	for _, head := range otherHeads {
		tmpBoard[head.Y][head.X] = -1
	}

	// 他Headが見つかるまでBFS
	type Coord struct{ X, Y, Dist int }
	queue := make([]Coord, 0, 30*20)
	queue = append(queue, Coord{X: myHead.X, Y: myHead.Y, Dist: 0})
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]

		if IsOnBoard(cur.X-1, cur.Y) {
			switch tmpBoard[cur.Y][cur.X-1] {
			case -1:
				return cur.Dist + 1
			case 0:
				nxt := Coord{X: cur.X - 1, Y: cur.Y, Dist: cur.Dist + 1}
				tmpBoard[nxt.Y][nxt.X] = myHead.ID
				queue = append(queue, nxt)
			}
		}

		if IsOnBoard(cur.X+1, cur.Y) {
			switch tmpBoard[cur.Y][cur.X+1] {
			case -1:
				return cur.Dist + 1
			case 0:
				nxt := Coord{X: cur.X + 1, Y: cur.Y, Dist: cur.Dist + 1}
				tmpBoard[nxt.Y][nxt.X] = myHead.ID
				queue = append(queue, nxt)
			}
		}

		if IsOnBoard(cur.X, cur.Y-1) {
			switch tmpBoard[cur.Y-1][cur.X] {
			case -1:
				return cur.Dist + 1
			case 0:
				nxt := Coord{X: cur.X, Y: cur.Y - 1, Dist: cur.Dist + 1}
				tmpBoard[nxt.Y][nxt.X] = myHead.ID
				queue = append(queue, nxt)
			}
		}

		if IsOnBoard(cur.X, cur.Y+1) {
			switch tmpBoard[cur.Y+1][cur.X] {
			case -1:
				return cur.Dist + 1
			case 0:
				nxt := Coord{X: cur.X, Y: cur.Y + 1, Dist: cur.Dist + 1}
				tmpBoard[nxt.Y][nxt.X] = myHead.ID
				queue = append(queue, nxt)
			}
		}
	}

	// 他Headとぶつからなかった = 他Headと接触し得ない
	return 0
}

func (bg *BumpGuard) WithBumpGuard(ops Ops, score int) int {
	rate := bg.rate[ops]
	if rate == 0 {
		// そもそも行けない or 他Headと接触し得ない
		return score
	}
	return int(float64(score)*bg.calcPenalty(rate)) + 1
}

// calcPenalty ペナルティを計算
func (bg *BumpGuard) calcPenalty(rate int) float64 {
	return math.Pow(1-1/float64(rate), 14)
}
