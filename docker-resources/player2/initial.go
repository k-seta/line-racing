/**
 * 初動戦略: 他プレイヤーの干渉がない段階での戦略。
 *
 * 1 - 7 turn: 最近接の壁に突進
 *     8 turn: 他Headがない方向にUターン
 * 9 -15 turn: 折り返す
 */
package main

type InitialDivision int

const (
	InitialDivisionNone InitialDivision = iota
	InitialDivisionLeftUp
	InitialDivisionLeftDown
	InitialDivisionRightUp
	InitialDivisionRightDown
)

func Initial(
	myHead Head,
	board [][]int,
	turn int,
) Ops {
	division := getInitialDivision(myHead, board)
	debug.SetScore("Initial", turn)

	switch division {
	case InitialDivisionLeftUp:
		switch {
		case 1 <= turn && turn <= 7:
			return LEFT
		case turn == 8:
			return UP
		case 9 <= turn && turn <= 15:
			return RIGHT
		}
	case InitialDivisionLeftDown:
		switch {
		case 1 <= turn && turn <= 7:
			return LEFT
		case turn == 8:
			return DOWN
		case 9 <= turn && turn <= 15:
			return RIGHT
		}
	case InitialDivisionRightUp:
		switch {
		case 1 <= turn && turn <= 7:
			return RIGHT
		case turn == 8:
			return UP
		case 9 <= turn && turn <= 15:
			return LEFT
		}
	case InitialDivisionRightDown:
		switch {
		case 1 <= turn && turn <= 7:
			return RIGHT
		case turn == 8:
			return DOWN
		case 9 <= turn && turn <= 15:
			return LEFT
		}
	}
	return NONE
}

// getInitialDivision 初動戦略の対象となる区画を取得
// 初期位置の数値をもとに判定する。
// 脱落者がいる場合は複数の初期位置が自分のIDで埋まりうる点に注意。
func getInitialDivision(
	myHead Head,
	board [][]int,
) InitialDivision {
	if board[7][7] == myHead.ID {
		return InitialDivisionLeftUp
	}
	if board[12][7] == myHead.ID {
		return InitialDivisionLeftDown
	}
	if board[7][22] == myHead.ID {
		return InitialDivisionRightUp
	}
	if board[12][22] == myHead.ID {
		return InitialDivisionRightDown
	}
	return InitialDivisionNone
}
