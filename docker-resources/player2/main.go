package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/v1/next", nextHandler)

	fmt.Println("Server listening on port 8000...")
	log.Fatal(http.ListenAndServe(":8000", nil))
}

type Head struct {
	ID int `json:"id"`
	X  int `json:"x"`
	Y  int `json:"y"`
}

type RequestModel struct {
	ID    int     `json:"id"`
	Heads []Head  `json:"heads"`
	Board [][]int `json:"board"`
}

type Ops string

const (
	NONE  Ops = "checkmated"
	UP    Ops = "up"
	DOWN  Ops = "down"
	LEFT  Ops = "left"
	RIGHT Ops = "right"
)

type ResponseModel struct {
	Ops Ops `json:"ops"`
}

func WriteResponse(w http.ResponseWriter, res *ResponseModel) {
	b, err := json.Marshal(res)
	if err != nil {
		panic(err)
	}
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

func ReadRequest(r *http.Request) *RequestModel {
	b := make([]byte, r.ContentLength)
	if _, err := r.Body.Read(b); err != nil {
		if err.Error() != "EOF" {
			panic(err)
		}
	}

	var req RequestModel
	if err := json.Unmarshal(b, &req); err != nil {
		panic(err)
	}
	return &req
}

// ヘルスチェック

func healthHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("GET /health 200 OK")
	type healthResp struct {
		Status string `json:"status"`
	}
	b, err := json.Marshal(healthResp{Status: "up"})
	if err != nil {
		panic(err)
	}

	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

// コマンド

func nextHandler(w http.ResponseWriter, r *http.Request) {
	debug.Reset()

	// リクエストから情報を取得
	req := ReadRequest(r)
	id := req.ID
	heads := req.Heads
	board := req.Board

	var myHead Head       // 自分のヘッド
	var otherHeads []Head // 他のヘッド
	for _, head := range heads {
		if head.ID == id {
			myHead = head
		} else {
			otherHeads = append(otherHeads, head)
		}
	}

	var turn int // 自分のターン数 = 自idマスの数
	for _, row := range board {
		for _, cell := range row {
			if cell == id {
				turn++
			}
		}
	}

	// 初動戦略
	if turn <= 15 {
		ops := Initial(myHead, board, turn)
		debug.PrintDecision()
		WriteResponse(w, &ResponseModel{Ops: ops})
		return
	}

	// BG初期化
	bumpGuard := NewBumpGuard(myHead, otherHeads, board)
	debug.PrintBumpGuard(bumpGuard)

	// ボードに衝突回避を適用
	board = BoardBumpGuard(board, myHead, heads)

	/*
	 * NOTE:
	 * BFS, Greedy, GreedyBT の３つの戦略を試行し、最もスコアの高い戦略を採択する。
	 * Greedy, GreedyBT は BumpGuard あり。BFS は BumpGuard なし。
	 *
	 * 結果的に、
	 * - 他Headと近い → BFSで遠くに避難
	 * - 他Headと遠い → Greedy系で最長経路を辿る
	 * というアルゴリズムになった。
	 * 近さと戦略のバランスはbumpGuard.CalcPenaltyのべき数で調整できる。
	 * 14だと、だいたい7マス以内に敵がいるとBFSを選ぶ。
	 */
	var maxScore int
	var maxOps Ops
	if ops, score := BFS(myHead, board, nil); score > maxScore {
		maxScore = score
		maxOps = ops
	}
	if ops, score := Greedy(myHead, board, bumpGuard); score > maxScore {
		maxScore = score
		maxOps = ops
	}
	if ops, score := GreedyBT(myHead, board, bumpGuard); score > maxScore {
		maxOps = ops
	}

	debug.PrintScores()
	debug.PrintDecision()

	WriteResponse(w, &ResponseModel{Ops: maxOps})
}
