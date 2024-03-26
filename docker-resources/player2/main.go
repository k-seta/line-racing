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
	X int `json:"x"`
	Y int `json:"y"`
}

type RequestModel struct {
	ID int `json:"id"`
	Heads []Head `json:"heads"`
	Board [][]int `json:"board"`
}

type Ops string
const (
	NONE Ops = "checkmated"
	UP Ops = "up"
	DOWN Ops = "down"
	LEFT Ops = "left"
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
	req := ReadRequest(r)

	id := req.ID
	heads := req.Heads
	board := req.Board

	var myHead Head // 自分のヘッド
	for _, head := range heads {
		if head.ID == id {
			myHead = head
			break
		}
	}

	var ops Ops
	switch {
	case myHead.Y > 0 && board[myHead.Y-1][myHead.X] == 0:
		ops = UP
	case myHead.X < 29 && board[myHead.Y][myHead.X+1] == 0:
		ops = RIGHT
	case myHead.X > 0 && board[myHead.Y][myHead.X-1] == 0:
		ops = LEFT
	case myHead.Y < 19 && board[myHead.Y+1][myHead.X] == 0:
		ops = DOWN
	default:
		ops = NONE
	}

	WriteResponse(w, &ResponseModel{Ops: ops})
}
