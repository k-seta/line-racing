import Phaser from "phaser";
import Board from "../domain/Board";
import Player from "../domain/Player";

export default class MyScene extends Phaser.Scene {

  isStarted: boolean = false;
  isUpdating: boolean = false;
  isFinished: boolean = false;

  board: Board | null = null;
  players: Player[] = new Array(0);

  constructor() {
    super("main");
  }

  preload() {
  }

  async create() {
    // 盤面描画
    this.board = new Board(this, 30, 20);
    this.board.draw();

    this.players.push(new Player("Player1", 1, 0xff0000, { x: 7, y: 7 }, "/p1/v1/next", this.add.text(1030, 100, "")))
    this.players.push(new Player("Player2", 2, 0x00ff00, { x: 22, y: 7 }, "/p2/v1/next", this.add.text(1030, 200, "")))
    this.players.push(new Player("Player3", 3, 0x00ffff, { x: 7, y: 12 }, "/p3/v1/next", this.add.text(1030, 300, "")))
    this.players.push(new Player("Player4", 4, 0xf5b2b2, { x: 22, y: 12 }, "/p4/v1/next", this.add.text(1030, 400, "")))
    // this.players.push(new Player("Player5", 5, 0xf5b2b2, { x: 22, y: 12 }, "/p5/v1/next", this.add.text(1030, 400, "")))

    await this.board.update(this.players);

    const readyText = this.add.text(1000, 735, "Click to Start..");
    this.input.on("pointerdown", () => {
      this.isStarted = true;
      readyText.setVisible(false);
    })
  }

  async update() {

    if (!this.isStarted || this.isFinished) {
      return;
    }

    this.players.forEach((player) => {
      if (!player.isDefeated || player.isDrawing()) {
        player.draw();
      } else {
        player.clear();
      }
    })

    const activePlayers = this.players.filter(p => !p.isDefeated)
    if (activePlayers.length == 1) {
      activePlayers[0].win();
      this.isFinished = true;
      return;
    }

    this.players.forEach((a) => {
      if (!a.isDefeated) {
        this.players.forEach((b) => {
          if (!b.isDefeated && a.id != b.id && a.head.x == b.head.x && a.head.y == b.head.y) {
            // 同じマスに侵入した場合敗北
            a.defeat();
            b.defeat();
          }
        })
      }
    })

    if (this.isUpdating) {
      return;
    }

    if (this.players.every((player) => {
      return !player.isDrawing()
    })) {
      this.isUpdating = true;
      const heads = this.players
        .filter((player) => {
          return !player.isDefeated
        })
        .map((player) => {
          return {
            id: player.id,
            x: player.head.x,
            y: player.head.y,
          };
        });

      await Promise.all(this.players.map(async (player) => {
        return player.next(this, this.board, heads);
      }));

      await this.board?.update(this.players);

      this.isUpdating = false;
    }
  }
}
