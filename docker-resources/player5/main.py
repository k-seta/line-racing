from enum import Enum
from fastapi import FastAPI
import dataclasses
import uvicorn

# 独自import
import random

# スコアリング用定数
AREASIZE_BONUS = 0.5
COLLISION_PENALTY = 0.1

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
@app.post("/v1/next")
def create_user(body: RequestBody):

  # TODO: ここからを独自のアルゴリズムに修正する(5秒以内にレスポンスを返せるようにすること)

  # 自機のIDと現在座標を取得
  head = next(filter(lambda x: x.id == body.id, body.heads), Coordinate(-1, -1, -1))

  for ops in EnumOps:
      if(ops == EnumOps.up):
        dest = Coordinate(body.id, head.x, head.y - 1)
      elif(ops == EnumOps.right):
        dest = Coordinate(body.id, head.x + 1, head.y)
      elif(ops == EnumOps.left):
        dest = Coordinate(body.id, head.x - 1, head.y)
      elif(ops == EnumOps.down):
        dest = Coordinate(body.id, head.x, head.y + 1)

      if(dest.x >= 0 and dest.y >= 0 and dest.x < len(body.board[0]) and dest.y < len(body.board) and body.board[dest.y][dest.x] == 0):
        return ResponseModel(ops)

  # 次ターンの進路スコア
  scores = {
    EnumOps.up: 0,
    EnumOps.right: 0,
    EnumOps.left: 0,
    EnumOps.down: 0
  }

  # 前提チェック、進めない方角を除外
  if util_check_dest(head.x, head.y - 1, body.board) == False: del scores[EnumOps.up]
  if util_check_dest(head.x + 1, head.y, body.board) == False: del scores[EnumOps.right]
  if util_check_dest(head.x - 1, head.y, body.board) == False: del scores[EnumOps.left]
  if util_check_dest(head.x, head.y + 1, body.board) == False: del scores[EnumOps.down]
  if len(scores) == 0:
    return ResponseModel(EnumOps.checkmated)

  # これ以降、進むこと自体は可能なのでスコアリングで優先度を計算する

  # 進行方向に広がる進入可能領域のサイズを評価
  scores = scoring_areasize(head, body, scores)

  # 次ターンに他プレイヤーと衝突しうる行動の評価を下げる
  scores = scoring_collision(body, scores)

  # 同スコアの場合の進路決定
  # return strategy_default(scores)
  return strategy_random(scores)

  # TODO: ここまでを独自のアルゴリズムに修正する

  return ResponseModel(EnumOps.checkmated)

### utils ##########
# スコアリングや方向決定戦略に絡まない便利機能
####################

# 進行可能なマスかのチェック
def util_check_dest(x, y, board):
  return True if (x >= 0 and y >= 0 and x < len(board[0]) and y < len(board) and board[y][x] == 0) else False


### scoring ##########
# 各方向の選択可能性の重みづけ
######################

# 上下左右に広がる領域(連続する0マス)のサイズをもとに優先度を決定
def scoring_areasize(head, body, scores):
  board = body.board

  def count_same_area(x, y, board, visited):
    # ボード範囲外や探索済みの場合0を返す
    if (util_check_dest(x, y, board) == False) or (x, y) in visited:
      return 0

    # 現在のマスを訪れたとしてマークする
    visited.add((x, y))

    # 上下左右のマスに対して再起的に探索し、同じ領域に属するマスの数を返す
    count = 1
    count += count_same_area(x + 1, y, board, visited)  # 右
    count += count_same_area(x - 1, y, board, visited)  # 左
    count += count_same_area(x, y + 1, board, visited)  # 下
    count += count_same_area(x, y - 1, board, visited)  # 上

    return count

  scores_fixed = scores

  areasizes = {}
  visited = set()
  for ops in scores.keys():
    if ops == EnumOps.up:
      areasizes[ops] = count_same_area(head.x, head.y - 1, board, visited)
    elif ops == EnumOps.right:
      areasizes[ops] = count_same_area(head.x + 1, head.y, board, visited)
    elif ops == EnumOps.left:
      areasizes[ops] = count_same_area(head.x - 1, head.y, board, visited)
    elif ops == EnumOps.down:
      areasizes[ops] = count_same_area(head.x, head.y + 1, board, visited)
    visited = set()  # initialize

  # 各方向に広がる領域について、最大領域との比で加点
  max_areasize = max(areasizes.values())
  for ops in scores.keys():
    scores_fixed[ops] += round((areasizes[ops] / max_areasize) * AREASIZE_BONUS, 3)

  return scores_fixed

# 次ターンに他プレイヤーと衝突しうる行動の評価を下げる
def scoring_collision(body, scores):
  scores_fixed = scores
  my_head = next(filter(lambda x: x.id == body.id, body.heads), Coordinate(-1, -1, -1))
  heads = filter(lambda x: x.id != body.id, body.heads)

  for ops in scores.keys():
    if ops == EnumOps.up:
      num_neighbors = len(list(filter(lambda h: abs(my_head.x - h.x) + abs(my_head.y - 1 - h.y) == 1, heads)))
    elif ops == EnumOps.right:
      num_neighbors = len(list(filter(lambda h: abs(my_head.x + 1 - h.x) + abs(my_head.y - h.y) == 1, heads)))
    elif ops == EnumOps.left:
      num_neighbors = len(list(filter(lambda h: abs(my_head.x - 1 - h.x) + abs(my_head.y - h.y) == 1, heads)))
    elif ops == EnumOps.down:
      num_neighbors = len(list(filter(lambda h: abs(my_head.x - h.x) + abs(my_head.y + 1 - h.y) == 1, heads)))
    scores_fixed[ops] -= COLLISION_PENALTY * num_neighbors

  return scores_fixed


### strategy ##########
# 最終的な方向決定戦略
#######################

# 戦略0: デフォルト戦略
def strategy_default(scores):
  return ResponseModel(list(scores.keys())[0])

# 戦略1: ランダムウォーク（スコアは考慮）
def strategy_random(scores):
  max_score = max(scores.values())
  dirc_alt = [dirc for dirc, score in scores.items() if score == max_score]
  random.shuffle(dirc_alt)
  return ResponseModel(dirc_alt[0])


@app.get("/health")
async def health():
    return {"status": "up"}

if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=8000)
