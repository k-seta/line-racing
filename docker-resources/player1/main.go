package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"math"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"gonum.org/v1/gonum/graph/path"
	"gonum.org/v1/gonum/graph/simple"
)

type Health struct {
	Status string `json:"status"`
}

type Coordinate struct {
	ID     int `json:"id"`
	CoordX int `json:"x"`
	CoordY int `json:"y"`
}

type RequestBody struct {
	ID    int          `json:"id"`
	Heads []Coordinate `json:"heads"`
	Board [][]int      `json:"board"`
}

type ResponseModel struct {
	Ops string `json:"ops"`
}

func main() {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/health", health)
	e.POST("/v1/next", next)

	e.Logger.Fatal(e.Start(":8000"))
}

func health(c echo.Context) error {
	return c.JSON(http.StatusOK, Health{Status: "up"})
}

func next(c echo.Context) error {
	var body RequestBody
	if err := c.Bind(&body); err != nil {
		return err
	}

	evals := recursiveEval(body, 1)
	fmt.Println(body.Heads)
	fmt.Println(evals)

	npos := -1
	evalMax := math.Inf(-1)
	for mapKey, value := range evals {
		if value >= evalMax {
			evalMax = value
			npos = mapKey
		}
	}

	var selfHead Coordinate
	for _, head := range body.Heads {
		if head.ID == body.ID {
			selfHead = head
			break
		}
	}

	result := ops(key(selfHead.CoordX, selfHead.CoordY), npos)

	return c.JSON(http.StatusOK, ResponseModel{Ops: result})
}

func sliceCopy(in, out interface{}) {
	buf := new(bytes.Buffer)
	gob.NewEncoder(buf).Encode(in)
	gob.NewDecoder(buf).Decode(out)
}

func key(x int, y int) int {
	return 100*x + y
}

func next_generation(body RequestBody) map[int][]RequestBody {

	// 初期化
	candidates := map[int][]Coordinate{}
	for _, head := range body.Heads {
		candidates[head.ID] = []Coordinate{}
	}

	// 各 head の次の移動選択肢を列挙
	for _, head := range body.Heads {
		x := head.CoordX
		y := head.CoordY

		// up できるか？
		if y-1 >= 0 && body.Board[y-1][x] == 0 {
			candidates[head.ID] = append(candidates[head.ID], Coordinate{ID: head.ID, CoordX: head.CoordX, CoordY: head.CoordY - 1})
		}

		// right できるか？
		if x+1 < len(body.Board[y]) && body.Board[y][x+1] == 0 {
			candidates[head.ID] = append(candidates[head.ID], Coordinate{ID: head.ID, CoordX: head.CoordX + 1, CoordY: head.CoordY})
		}

		// left できるか？
		if x-1 >= 0 && body.Board[y][x-1] == 0 {
			candidates[head.ID] = append(candidates[head.ID], Coordinate{ID: head.ID, CoordX: head.CoordX - 1, CoordY: head.CoordY})
		}

		// down できるか？
		if y+1 < len(body.Board) && body.Board[y+1][x] == 0 {
			candidates[head.ID] = append(candidates[head.ID], Coordinate{ID: head.ID, CoordX: head.CoordX, CoordY: head.CoordY + 1})
		}
	}

	nheads := [][]Coordinate{[]Coordinate{}}
	for _, head := range body.Heads {

		if len(candidates[head.ID]) == 0 {
			continue
		}

		memo := [][]Coordinate{}
		for _, tmp := range nheads {
			for _, candidate := range candidates[head.ID] {
				memo = append(memo, append(tmp, candidate))
			}
		}
		nheads = memo
	}

	// 相手が詰んだ場合、通った線が消えるので、先に消しておく
	var nBaseboard [][]int
	sliceCopy(body.Board, &nBaseboard)
	for _, head := range body.Heads {
		if len(candidates[head.ID]) == 0 {
			for y := 0; y < len(nBaseboard); y++ {
				for x := 0; x < len(nBaseboard[y]); x++ {
					if nBaseboard[y][x] == head.ID {
						nBaseboard[y][x] = 0
					}
				}
			}
		}
	}

	nbodies := map[int][]RequestBody{}
	for _, heads := range nheads {

		// deep copy
		var nboard [][]int
		sliceCopy(nBaseboard, &nboard)

		var mapkey int
		for _, head := range heads {
			nboard[head.CoordY][head.CoordX] = head.ID

			if head.ID == body.ID {
				mapkey = key(head.CoordX, head.CoordY)
			}
		}

		if _, ok := nbodies[mapkey]; ok {
			nbodies[mapkey] = append(nbodies[mapkey], RequestBody{ID: body.ID, Heads: heads, Board: nboard})
		} else {
			nbodies[mapkey] = []RequestBody{RequestBody{ID: body.ID, Heads: heads, Board: nboard}}
		}
	}

	return nbodies
}

func eval(body RequestBody) int {
	// 初期化処理
	// graph の node 作成
	graph := simple.NewDirectedGraph()
	for y := 0; y < len(body.Board); y++ {
		for x := 0; x < len(body.Board[y]); x++ {
			id := key(x, y)
			node := simple.Node(id)
			graph.AddNode(node)
		}
	}

	// graph の edge 作成
	for y := 0; y < len(body.Board); y++ {
		for x := 0; x < len(body.Board[y]); x++ {

			isHead := false
			for _, head := range body.Heads {
				if head.CoordX == x && head.CoordY == y {
					isHead = true
				}
			}

			// head ではないかつ空きマスでないなら、早期 return
			if body.Board[y][x] != 0 && !isHead {
				continue
			}

			from := simple.Node(key(x, y))

			// up できるか？
			if y-1 >= 0 && body.Board[y-1][x] == 0 {
				to := simple.Node(key(x, y-1))
				edge := graph.NewEdge(from, to)
				graph.SetEdge(edge)
			}

			// right できるか？
			if x+1 < len(body.Board[y]) && body.Board[y][x+1] == 0 {
				to := simple.Node(key(x+1, y))
				edge := graph.NewEdge(from, to)
				graph.SetEdge(edge)
			}

			// left できるか？
			if x-1 >= 0 && body.Board[y][x-1] == 0 {
				to := simple.Node(key(x-1, y))
				edge := graph.NewEdge(from, to)
				graph.SetEdge(edge)
			}

			// down できるか？
			if y+1 < len(body.Board) && body.Board[y+1][x] == 0 {
				to := simple.Node(key(x, y+1))
				edge := graph.NewEdge(from, to)
				graph.SetEdge(edge)
			}
		}
	}

	shortests := map[int]path.Shortest{}
	evals := map[int]int{}
	for _, head := range body.Heads {

		// 各マスへの最短経路を計算
		shortests[head.ID] = path.DijkstraFrom(simple.Node(key(head.CoordX, head.CoordY)), graph)

		// 評価値格納 map 初期化
		evals[head.ID] = 0
	}

	// 各マスへの距離が threshold 以下のマス目の合計を計算
	threshold := 50
	for y := 0; y < len(body.Board); y++ {
		for x := 0; x < len(body.Board[y]); x++ {
			for _, head := range body.Heads {

				// from head to (x, y)
				_, length := shortests[head.ID].To(int64(key(x, y)))

				if length < float64(threshold) {
					evals[head.ID]++
				}

				for _, ohead := range body.Heads {
					if ohead.ID != head.ID && ohead.CoordX == head.CoordX && ohead.CoordY == head.CoordY {
						evals[head.ID] -= 100
					}
				}
			}
		}
	}

	result := 0
	for _, head := range body.Heads {
		if body.ID == head.ID {
			result += evals[head.ID]
		} else {
			result += evals[body.ID] - evals[head.ID]
		}
	}

	return result
}

func recursiveEval(body RequestBody, generation int) map[int]float64 {

	candidates := next_generation(body)

	evals := map[int]float64{}
	if generation == 0 {
		// 候補手ごとの自 ID 評価値の平均を取得
		for mapkey, ngens := range candidates {
			tmp := 0
			for _, ngen := range ngens {
				tmp += eval(ngen)
			}
			evals[mapkey] = float64(tmp) / float64(len(ngens))
		}
		return evals
	} else {
		for mapkey, ngens := range candidates {
			evalMax := math.Inf(-1)
			tmp := float64(0)
			for _, ngen := range ngens {
				recEvals := recursiveEval(ngen, generation-1)
				for _, score := range recEvals {
					tmp += score
				}
				tmp = tmp / float64(len(recEvals))
				if tmp >= evalMax {
					evalMax = tmp
				}
			}
			evals[mapkey] = evalMax
		}
		return evals
	}
}

func ops(from int, to int) string {
	diff := to - from

	if diff == -1 {
		return "up"
	} else if diff == 100 {
		return "right"
	} else if diff == -100 {
		return "left"
	} else if diff == 1 {
		return "down"
	} else {
		return "checkmated"
	}
}
