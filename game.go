package main

import (
	"fmt"
	"image"
	"log"
	"math"
	"rpg-sphere/animations"
	"rpg-sphere/components"
	"rpg-sphere/constants"
	"rpg-sphere/entities"
	"rpg-sphere/spritesheet"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type Game struct {
	player            *entities.Player
	PlayerSpriteSheet *spritesheet.SpriteSheet
	enemies           []*entities.Enemy
	potions           []*entities.Potion
	tilemapJSON       *TilemapJSON
	tilesets          []Tileset
	tilemapImg        *ebiten.Image
	camera            *Camera
	colliders         []image.Rectangle
}

func NewGame() *Game {
	playerImg, _, err := ebitenutil.NewImageFromFile("assets/images/ninja.png")
	if err != nil {
		// handle error
		fmt.Println("ninja Img")
		log.Fatal(err)
	}

	skeletonImg, _, err := ebitenutil.NewImageFromFile("assets/images/skeleton.png")
	if err != nil {
		// handle error
		fmt.Println("skeleton Img")
		log.Fatal(err)
	}

	potionImg, _, err := ebitenutil.NewImageFromFile("assets/images/potion.png")
	if err != nil {
		// handle error
		fmt.Println("potion Img")
		log.Fatal(err)
	}

	tilemapImg, _, err := ebitenutil.NewImageFromFile("assets/images/TilesetFloor.png")
	if err != nil {
		// handle error
		fmt.Println("TilesetFloor Img")
		log.Fatal(err)
	}

	tilemapJSON, err := NewTilemapJSON("assets/maps/spawn.json")
	if err != nil {
		fmt.Println("spawn Img")
		log.Fatal(err)
	}

	tilesets, err := tilemapJSON.GenTilesets()
	if err != nil {
		fmt.Println("Tilesets")
		log.Fatal(err)
	}

	playerSpriteSheet := spritesheet.NewSpriteSheet(4, 7, 16)

	return &Game{
		player: &entities.Player{
			Sprite: &entities.Sprite{
				Img: playerImg,
				X:   400.0,
				Y:   350.0,
			},
			Health: 5.0,
			Animations: map[entities.PlayerState]*animations.Animation{
				entities.Up:    animations.NewAnimation(5, 13, 4, 20.0),
				entities.Down:  animations.NewAnimation(4, 12, 4, 20.0),
				entities.Left:  animations.NewAnimation(6, 14, 4, 20.0),
				entities.Right: animations.NewAnimation(7, 15, 4, 20.0),
			},
			CombatComp: components.NewBasicCombat(3, 1),
		},
		PlayerSpriteSheet: playerSpriteSheet,
		enemies: []*entities.Enemy{
			{
				Sprite: &entities.Sprite{
					Img: skeletonImg,
					X:   100.0,
					Y:   100.0,
				},
				FollowsPlayer: false,
				CombatComp:    components.NewEnemyCombat(3, 1, 30),
			},
			{
				Sprite: &entities.Sprite{
					Img: skeletonImg,
					X:   150.0,
					Y:   150.0,
				},
				FollowsPlayer: true,
				CombatComp:    components.NewEnemyCombat(3, 1, 30),
			},
			{
				Sprite: &entities.Sprite{
					Img: skeletonImg,
					X:   75.0,
					Y:   75.0,
				},
				FollowsPlayer: false,
				CombatComp:    components.NewEnemyCombat(3, 1, 30),
			},
		},
		potions: []*entities.Potion{
			{
				Sprite: &entities.Sprite{
					Img: potionImg,
					X:   150.0,
					Y:   150.0,
				},
				AmtHeal: 1.0,
			},
		},
		tilemapJSON: tilemapJSON,
		tilemapImg:  tilemapImg,
		tilesets:    tilesets,
		camera:      NewCamera(400.0, 400.0),
		colliders: []image.Rectangle{
			image.Rect(100, 100, 116, 116),
		},
	}
}

func (g *Game) Update() error {
	//react to ket press

	g.player.Dx = 0.0
	g.player.Dy = 0.0
	if ebiten.IsKeyPressed(ebiten.KeyRight) {
		g.player.Dx = 2
	}
	if ebiten.IsKeyPressed(ebiten.KeyLeft) {
		g.player.Dx = -2
	}
	if ebiten.IsKeyPressed(ebiten.KeyUp) {
		g.player.Dy = -2
	}
	if ebiten.IsKeyPressed(ebiten.KeyDown) {
		g.player.Dy = 2
	}

	g.player.X += g.player.Dx
	CheckCollisionHorizontal(g.player.Sprite, g.colliders)
	g.player.Y += g.player.Dy
	CheckCollisionVeritcal(g.player.Sprite, g.colliders)

	activeAnim := g.player.ActiveAnimation(int(g.player.Dx), int(g.player.Dy))
	if activeAnim != nil {
		activeAnim.Update()
	}

	for _, enemy := range g.enemies {
		enemy.Dx = 0.0
		enemy.Dy = 0.0
		if enemy.FollowsPlayer {

			if enemy.X < g.player.X {
				enemy.Dx += 1
			} else if enemy.X > g.player.X {
				enemy.Dx -= 1
			}

			if enemy.Y > g.player.Y {
				enemy.Dy -= 1
			} else if enemy.Y < g.player.Y {
				enemy.Dy += 1
			}
		}

		enemy.X += enemy.Dx
		CheckCollisionHorizontal(enemy.Sprite, g.colliders)

		enemy.Y += enemy.Dy
		CheckCollisionVeritcal(enemy.Sprite, g.colliders)

	}

	clicked := inpututil.IsMouseButtonJustPressed(ebiten.MouseButton0)
	cX, cY := ebiten.CursorPosition()
	cX -= int(g.camera.X)
	cY -= int(g.camera.Y)
	g.player.CombatComp.Update()
	pRect := image.Rect(
		int(g.player.X),
		int(g.player.Y),
		int(g.player.X)+constants.Tilesize,
		int(g.player.Y)+constants.Tilesize,
	)

	deadEnemies := make(map[int]struct{})
	for index, enemy := range g.enemies {
		enemy.CombatComp.Update()
		rect := image.Rect(
			int(enemy.X),
			int(enemy.Y),
			int(enemy.X)+constants.Tilesize,
			int(enemy.Y)+constants.Tilesize,
		)

		// if enemy overlaps player1
		if rect.Overlaps(pRect) {
			if enemy.CombatComp.Attack() {
				g.player.CombatComp.Damage(enemy.CombatComp.AttackPower())
				fmt.Println(fmt.Printf("Player damaged, health: %d\n", g.player.CombatComp.Health()))
				if g.player.CombatComp.Health() <= 0 {
					fmt.Println("player has died")
				}
			}
		}

		//if cursor in react
		if cX > rect.Min.X && cX < rect.Max.X && cY > rect.Min.Y && cY < rect.Max.Y {
			if clicked &&
				math.Sqrt(
					math.Pow(float64(cX)-g.player.X+(constants.Tilesize/2), 2)+math.Pow(float64(cY)-g.player.Y+(constants.Tilesize/2), 2),
				) < constants.Tilesize*5 {
				fmt.Print("click")
				enemy.CombatComp.Damage(g.player.CombatComp.AttackPower())

				if enemy.CombatComp.Health() <= 0 {
					fmt.Print("eliminated")
					deadEnemies[index] = struct{}{}
				}
			}
		}
	}

	if len(deadEnemies) > 0 {
		newEnemies := make([]*entities.Enemy, 0)
		for index, enemy := range g.enemies {
			if _, exists := deadEnemies[index]; !exists {
				newEnemies = append(newEnemies, enemy)
			}
		}
		g.enemies = newEnemies
	}

	// for _, potion := range g.potions {
	// 	if g.player.X > potion.X {
	// 		g.player.Health += potion.AmtHeal
	// 		fmt.Printf("Picked up potion! Health: %d\n", g.player.Health)
	// 	}
	// }

	g.camera.FollowTarget(g.player.X+8, g.player.Y+8, 320, 240)
	g.camera.Constrain(
		float64(g.tilemapJSON.Layers[0].Width)*16.0,
		float64(g.tilemapJSON.Layers[0].Height)*16.0,
		320,
		240,
	)
	return nil
}
