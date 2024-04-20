from enum import Enum
from fastapi import FastAPI
import dataclasses
import uvicorn

app = FastAPI()

@dataclasses.dataclass
class Coordinate:
  id: int
  x: int
  y: int


@dataclasses.dataclass
class RequestBody:
  id: int
  heads: list[Coordinate]
  board: list[list[int]] # 左上原点。x 軸右向き、y 軸左向き。(x, y) = board[y][x]


class EnumOps(Enum):
    up = "up"
    right = "right"
    left = "left"
    down = "down"
    checkmated = "checkmated"

    @classmethod
    def values(cls):
        return [i.value for i in cls]


@dataclasses.dataclass
class ResponseModel:
  ops: EnumOps

#移動可能なマスの数: 周囲2マスが空かいなかをスコア化
def get_empty_neighbors(x, y, board, id):
    empty_neighbors = 0
    for dx in range(-4, 4):
        for dy in range(-4, 4):
            new_x = x + dx
            new_y = y + dy
            if 0 <= new_x < len(board[0]) and 0 <= new_y < len(board):
                if board[new_y][new_x] == 0 or board[new_y][new_x] == id:
                    empty_neighbors += 1
    return empty_neighbors

 #敵との距離：最も近い敵との距離((x座標の2乗+y座標の2乗)の0.5乗) 
def calculate_enemy_distance(coord, heads):
    min_distance = float('inf')
    for head in heads:
        distance = ((coord.x - head.x) ** 2 + (coord.y - head.y) ** 2) ** 0.5
        min_distance = min(min_distance, distance)
    return min_distance  

#壁との距離：最も近い壁との距離
def calculate_wall_distance(coord, board):
    top_distance = coord.y
    bottom_distance = len(board) - coord.y - 1
    left_distance = coord.x
    right_distance = len(board[0]) - coord.x - 1
    return min(top_distance, bottom_distance, left_distance, right_distance)
  
def heuristic_evaluation(body, ops):
    head = next(filter(lambda x: x.id == body.id, body.heads), Coordinate(-1, -1, -1))
    
    #create_userと処理が重複しており冗長だが、良いコードの作り方がわからぬ。。この処理だけ外だししてもいまいちだし。。
    dest = None
    if ops == EnumOps.up:
        dest = Coordinate(body.id, head.x, head.y - 1)
    elif ops == EnumOps.right:
        dest = Coordinate(body.id, head.x + 1, head.y)
    elif ops == EnumOps.left:
        dest = Coordinate(body.id, head.x - 1, head.y)
    elif ops == EnumOps.down:
        dest = Coordinate(body.id, head.x, head.y + 1)
    
    if dest is None or not (0 <= dest.x < len(body.board[0]) and 0 <= dest.y < len(body.board)):
        return 0
    
    empty_neighbors = get_empty_neighbors(dest.x, dest.y, body.board, body.id)
    enemy_distance = calculate_enemy_distance(dest, body.heads)
    wall_distance = calculate_wall_distance(dest, body.board)

    # これらの要素に重みをつけて総合的な評価を行う
    # 周りの空マスが多く、敵との距離が小さく、壁との距離が小さくなる場合にスコアが高くなる
    score = empty_neighbors * 1 - enemy_distance * 0.2 + wall_distance * 0.3
    return score

@app.post("/v1/next")
def create_user(body: RequestBody):
  
  head = next(filter(lambda x: x.id == body.id, body.heads), Coordinate(-1, -1, -1))
    
  best_ops = EnumOps.checkmated
  best_score = float('-inf')
    
  for ops in EnumOps:
      if ops == EnumOps.up:
          dest = Coordinate(body.id, head.x, head.y - 1)
      elif ops == EnumOps.right:
          dest = Coordinate(body.id, head.x + 1, head.y)
      elif ops == EnumOps.left:
          dest = Coordinate(body.id, head.x - 1, head.y)
      elif ops == EnumOps.down:
          dest = Coordinate(body.id, head.x, head.y + 1)
            
      if 0 <= dest.x < len(body.board[0]) and 0 <= dest.y < len(body.board) and body.board[dest.y][dest.x] == 0:
          score = heuristic_evaluation(body, ops)
          if score > best_score:
              best_score = score
              best_ops = ops
  print(best_ops)
  return ResponseModel(best_ops)

# TODO: ここまでを独自のアルゴリズムに修正する
#  return ResponseModel(EnumOps.checkmated)


@app.get("/health")
async def health():
    return {"status": "up"}

if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=8000)
