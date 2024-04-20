import dataclasses
import random
import time
from collections import deque
from enum import Enum

import numpy as np
import uvicorn
from fastapi import FastAPI

app = FastAPI()
np.set_printoptions(
    threshold=10000,
    linewidth=1000,
)

WIDTH = 30
HEIGHT = 20


@dataclasses.dataclass
class Coordinate:
    id: int
    x: int
    y: int


@dataclasses.dataclass
class RequestBody:
    id: int
    heads: list[Coordinate]
    board: list[list[int]]  # 左上原点。x 軸右向き、y 軸左向き。(x, y) = board[y][x]


class EnumOps(Enum):
    up = "up"
    right = "right"
    left = "left"
    down = "down"
    checkmated = "checkmated"

    @classmethod
    def values(cls):
        return [i.value for i in cls]


def direction(ops):
    if ops == EnumOps.up:
        return (0, -1)
    elif ops == EnumOps.right:
        return (1, 0)
    elif ops == EnumOps.left:
        return (-1, 0)
    elif ops == EnumOps.down:
        return (0, 1)
    return (0, 0)


def convert_to_ops(dx: int, dy: int):
    if dx == 0 and dy == -1:
        return EnumOps.up
    elif dx == 1 and dy == 0:
        return EnumOps.right
    elif dx == -1 and dy == 0:
        return EnumOps.left
    elif dx == 0 and dy == 1:
        return EnumOps.down
    return EnumOps.checkmated


@dataclasses.dataclass
class ResponseModel:
    ops: EnumOps


@dataclasses.dataclass
class RouteNode:
    x: int
    y: int
    prev_node: "RouteNode"
    next_nodes: dict[EnumOps, "RouteNode"]
    score: int
    searched: bool


OPS_LIST = [EnumOps.up, EnumOps.right, EnumOps.left, EnumOps.down]
RESPONSE_TIME_LIMIT = 4.5


@app.post("/v1/next")
def create_user(body: RequestBody):
    # TODO: ここからを独自のアルゴリズムに修正する(5秒以内にレスポンスを返せるようにすること)
    start_time = time.time()
    print(f"start_time: {start_time}")
    head = next(filter(lambda x: x.id == body.id, body.heads), Coordinate(-1, -1, -1))
    # return ResponseModel(
    #     convert_to_ops(*search_longest_route(body.board, head.x, head.y, body.id))
    # )

    voronnoi_count_by_ops = calculate_voronoi_count_by_ops(
        body.board, body.heads, body.id
    )
    print(f"voronnoi_count_by_ops: {voronnoi_count_by_ops}")
    if len(voronnoi_count_by_ops) > 1:
        max_voronoi_count = max(voronnoi_count_by_ops.values())
        max_voronoi_count_ops_list = [
            ops
            for ops, count in voronnoi_count_by_ops.items()
            if count == max_voronoi_count
        ]
        if len(max_voronoi_count_ops_list) == 1:
            return ResponseModel(max_voronoi_count_ops_list[0])
        window = 5 if max_voronoi_count > 100 else 10
        return ResponseModel(
            search_longest_route_v2(
                body.board,
                head.x,
                head.y,
                body.id,
                max_voronoi_count_ops_list,
                start_time + RESPONSE_TIME_LIMIT,
                window,
                max_voronoi_count,
            )
        )
    elif len(voronnoi_count_by_ops) == 1:
        max_ops = max(voronnoi_count_by_ops, key=voronnoi_count_by_ops.get)
        return ResponseModel(max_ops)

    return ResponseModel(EnumOps.checkmated)


@app.get("/health")
async def health():
    return {"status": "up"}


def calculate_voronoi_count_by_ops(board, heads, id):
    head = next(filter(lambda x: x.id == id, heads), Coordinate(-1, -1, -1))
    voronoi_count_by_ops = {}
    for ops in OPS_LIST:
        dx, dy = direction(ops)
        if not is_movable(board, head.x + dx, head.y + dy):
            continue
        _heads = [
            Coordinate(
                h.id,
                h.x + dx,
                h.y + dy,
            )
            if h.id == head.id
            else h
            for h in heads
        ]
        count = calculate_voronoi_counts(board, _heads).get(head.id, 0)
        voronoi_count_by_ops[ops] = count
    return voronoi_count_by_ops


BFS_MAX_VALUE = 10000


def is_movable(board: (list[list[int]] | np.ndarray), x: int, y: int):
    if isinstance(board, np.ndarray):
        return 0 <= x < WIDTH and 0 <= y < HEIGHT and board[y, x] == 0
    return 0 <= x < WIDTH and 0 <= y < HEIGHT and board[y][x] == 0


def calculate_voronoi_counts(board, heads):
    CONFLICT = 10
    voronoi_board = np.zeros((HEIGHT, WIDTH))
    bfs_boards = [bfs(board, head.x, head.y) for head in heads]
    for y in range(HEIGHT):
        for x in range(WIDTH):
            min_distance = BFS_MAX_VALUE
            for i, head in enumerate(heads):
                if bfs_boards[i][y, x] == BFS_MAX_VALUE:
                    continue
                elif min_distance > bfs_boards[i][y, x]:
                    voronoi_board[y, x] = head.id
                    min_distance = bfs_boards[i][y, x]
                elif min_distance == bfs_boards[i][y, x]:
                    voronoi_board[y, x] = CONFLICT
    voronoi_heads, counts = np.unique(voronoi_board, return_counts=True)
    return dict(zip(voronoi_heads, counts))


def kill(board, id):
    for y in range(HEIGHT):
        for x in range(WIDTH):
            if board[y][x] == id:
                board[y][x] = 0
    return board


def bfs(board, x, y):
    queue = deque()
    bfs_board = np.full((HEIGHT, WIDTH), BFS_MAX_VALUE)
    bfs_board[y, x] = 0
    queue.append((x, y))
    directions = [(1, 0), (-1, 0), (0, 1), (0, -1)]
    while len(queue) > 0:
        x, y = queue.popleft()
        for dx, dy in directions:
            next_x, next_y = x + dx, y + dy
            if next_x < 0 or next_x >= WIDTH or next_y < 0 or next_y >= HEIGHT:
                continue
            if board[next_y][next_x] != 0:
                continue
            if bfs_board[next_y, next_x] > bfs_board[y, x] + 1:
                bfs_board[next_y, next_x] = bfs_board[y, x] + 1
                queue.append((next_x, next_y))
            continue
    return bfs_board


def search_longest_route_v2(
    board, x, y, id, root_ops_list, abort_time, window=20, terminate_score=None
):
    ops_list = [EnumOps.up, EnumOps.right, EnumOps.left, EnumOps.down]
    root = RouteNode(
        x=x,
        y=y,
        prev_node=None,
        next_nodes={},
        score=0,
        searched=False,
    )
    route = deque()
    route_board = np.array(board)
    current_node = root
    searched_count = 0
    max_score = 0
    while time.time() < abort_time:
        if terminate_score is not None and max_score >= terminate_score:
            break
        next_ops_list = list(
            filter(
                lambda ops: is_movable(
                    route_board,
                    current_node.x + direction(ops)[0],
                    current_node.y + direction(ops)[1],
                )
                # and x - window <= current_node.x + direction(ops)[0] <= x + window
                # and y - window <= current_node.y + direction(ops)[1] <= y + window
                and not (
                    ops in current_node.next_nodes
                    and current_node.next_nodes[ops].searched
                ),
                ops_list if len(route) > 0 else root_ops_list,
            )
        )
        if (
            len(next_ops_list) == 0
            or (current_node.x < x - window or current_node.x > x + window)
            or (current_node.y < y - window or current_node.y > y + window)
        ):
            if len(route) == 0:
                break
            route_node = route.pop()
            node_scores = [node.score for node in route_node.next_nodes.values()]
            node_scores.append(len(route) + 1)
            route_node.score = max(node_scores)
            route_node.searched = True
            route_board[route_node.y, route_node.x] = 0
            if max_score < route_node.score:
                max_score = route_node.score
            while route:
                route_node = route.pop()
                node_scores = [node.score for node in route_node.next_nodes.values()]
                route_node.score = max(node_scores)
                nops_list = list(
                    filter(
                        lambda ops: is_movable(
                            route_board,
                            route_node.x + direction(ops)[0],
                            route_node.y + direction(ops)[1],
                        )
                        and not (
                            ops in route_node.next_nodes
                            and route_node.next_nodes[ops].searched
                        ),
                        ops_list if len(route) > 0 else root_ops_list,
                    )
                )
                if len(nops_list) == 0:
                    route_node.searched = True
                route_board[route_node.y, route_node.x] = 0
            searched_count += 1
            current_node = root
            continue

        next_ops = random.choice(next_ops_list)
        if next_ops not in current_node.next_nodes:
            current_node.next_nodes[next_ops] = RouteNode(
                x=current_node.x + direction(next_ops)[0],
                y=current_node.y + direction(next_ops)[1],
                prev_node=current_node,
                next_nodes={},
                score=0,
                searched=False,
            )
        current_node = current_node.next_nodes[next_ops]
        route.append(current_node)
        route_board[current_node.y, current_node.x] = id
    best_ops = max(root.next_nodes, key=lambda ops: root.next_nodes[ops].score)
    print(
        f"id: {id}, (x, y): ({x}, {y}), window: {window}, best_ops: {best_ops}, score: {root.next_nodes[best_ops].score}, searched_count: {searched_count}"
    )

    return best_ops


def search_longest_route(board, x, y, id, root_ops_list, abort_time):
    directions = [(1, 0), (-1, 0), (0, 1), (0, -1)]
    longest_distance = 0
    best_direction = (0, 0)
    loop_count = 0
    root_directions = [direction(ops) for ops in root_ops_list]
    print(f"root_directions: {root_directions}")
    while time.time() < abort_time:
        loop_count += 1
        route_board = np.array(board)
        route = deque()
        next_directions = list(
            filter(
                lambda d: 0 <= x + d[0] < WIDTH
                and 0 <= y + d[1] < HEIGHT
                and route_board[y + d[1], x + d[0]] == 0,
                root_directions,
            )
        )
        if len(next_directions) == 0:
            break
        next_direction = random.choice(next_directions)
        route.append(next_direction)
        next_x, next_y = x + next_direction[0], y + next_direction[1]
        route_board[y + next_direction[1], x + next_direction[0]] = id
        while True:
            if time.time() >= abort_time:
                break
            next_directions = list(
                filter(
                    lambda d: 0 <= next_x + d[0] < WIDTH
                    and 0 <= next_y + d[1] < HEIGHT
                    and route_board[next_y + d[1], next_x + d[0]] == 0,
                    directions,
                )
            )
            if len(next_directions) == 0:
                if len(route) > longest_distance:
                    longest_distance = len(route)
                    best_direction = route[0]
                    print(
                        f"longest_distance: {longest_distance}, best_direction: {best_direction}, time: {time.time()}, abort_time: {abort_time}"
                    )
                break
            next_direction = random.choice(next_directions)
            route.append(next_direction)
            next_x += next_direction[0]
            next_y += next_direction[1]
            route_board[next_y, next_x] = id
    best_ops = convert_to_ops(*best_direction)
    print(
        f"longest_distance: {longest_distance}, best_ops: {best_ops}, loop_count: {loop_count}"
    )
    return best_ops


if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=8000)
