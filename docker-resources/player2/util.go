package main

// CopyBoard ボードのDeepCopyを作成
func CopyBoard(board [][]int) [][]int {
	tmpBoard := make([][]int, 20)
	for i := range 20 {
		tmpBoard[i] = make([]int, 30)
		copy(tmpBoard[i], board[i])
	}
	return tmpBoard
}

// IsOnBoard ボードの範囲内かどうか。
// 移動可能か否かは考慮してないので注意
func IsOnBoard(x, y int) bool {
	return 0 <= x && x < 30 && 0 <= y && y < 20
}

// ScoreAggregator スコア集計用
type ScoreAggregator struct {
	maxScores map[Ops]int
}

func NewScoreAggregator() *ScoreAggregator {
	return &ScoreAggregator{
		maxScores: make(map[Ops]int, 4),
	}
}

// SetScore スコアを記録
func (s *ScoreAggregator) SetScore(ops Ops, score int) {
	if s.maxScores[ops] < score {
		s.maxScores[ops] = score
	}
}

// GetMaxScore 最高スコアのコマンドとそのスコアを返す
func (s *ScoreAggregator) GetMaxScore() (Ops, int) {
	maxScore := 0
	maxOps := NONE
	for ops, score := range s.maxScores {
		if maxScore < score {
			maxScore = score
			maxOps = ops
		}
	}
	return maxOps, maxScore
}
