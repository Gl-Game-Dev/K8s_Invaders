package main

import (
	"fmt"
	rl "github.com/gen2brain/raylib-go/raylib"
	v1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	"math"

	"github.com/Gl-Game-Dev/K8s_Invaders/internal/k8s"
)

type Player struct {
	rect rl.Rectangle
}

type Alien struct {
	rect    rl.Rectangle
	enabled bool
}

type AlienSwarm struct {
	aliens   []Alien
	minLeft  int
	maxRight int
}

type Bullet struct {
	rect    rl.Rectangle
	enabled bool
}
type BulletBuffer struct {
	Bullets []Bullet
}

var screenWidth = 800
var screenHeight = 450
var deploymentNumber = 0
var framecounter = 0
var alienSpeedX = 1
var alienSpeedY = 40
var playerSpeed float32 = 50
var bulletSpeed = 30
var servers = 10
var gameOver bool = false
var characterWidth = 40
var characterHeight = 40

func main() {
	deploymentsClient, deployment := k8s.ConnectK8s()
	bullets := BulletBuffer{}
	player := Player{rl.Rectangle{float32(screenWidth) / 2, float32(screenHeight) - 50, float32(characterWidth), float32(characterHeight)}}

	//aliens := []Alien{}
	aliens := AlienSwarm{}
	for i := range servers {
		alien := Alien{rl.Rectangle{float32(i*2*characterWidth + 10), float32(characterHeight) * 2, float32(characterWidth), float32(characterHeight)}, true}
		aliens.aliens = append(aliens.aliens, alien)
	}
	//serverBlock := AlienBounds{}
	rl.InitWindow(int32(screenWidth), int32(screenHeight), "K8s_Invaders")
	defer rl.CloseWindow()
	rl.SetTargetFPS(20)
	k8s.CreateDeployment(deploymentsClient, deployment)
	go k8s.UpdateDeployment(deploymentsClient, int32(servers))
	for !rl.WindowShouldClose() && !gameOver {
		checkinput(&player, &bullets)
		checkPhysics(&bullets, deploymentsClient, &player, &aliens)

		rl.BeginDrawing()
		rl.ClearBackground(rl.Black)
		replicaText := fmt.Sprintf("Replicas: %d", servers)
		rl.DrawText(replicaText, 10, 10, 20, rl.Red)
		for i := range aliens.aliens {
			if aliens.aliens[i].enabled {
				rl.DrawRectangleRec(aliens.aliens[i].rect, rl.Red)
			}
		}
		for i := range bullets.Bullets {
			rl.DrawRectangleRec(bullets.Bullets[i].rect, rl.Blue)
		}
		rl.DrawRectangleRec(player.rect, rl.White)
		rl.EndDrawing()
	}
	k8s.DeleteDeployment(deploymentsClient)
}

func checkinput(player *Player, bullets *BulletBuffer) {
	if rl.IsKeyDown(rl.KeyD) {
		if player.rect.X+float32(characterWidth) < float32(screenWidth-characterWidth) {
			player.rect.X += playerSpeed
		}
	}
	if rl.IsKeyDown(rl.KeyA) {
		if player.rect.X > 0 {
			player.rect.X -= playerSpeed
		}
	}
	if rl.IsKeyPressed(rl.KeySpace) {
		bullet := Bullet{rl.Rectangle{player.rect.X + (float32(characterWidth) / 2), player.rect.Y, 10, 5}, true}
		bullets.Bullets = append(bullets.Bullets, bullet)
	}
}

func checkPhysics(bullets *BulletBuffer, deploymentsClient v1.DeploymentInterface, player *Player, aliens *AlienSwarm) {
	for i := range bullets.Bullets {
		for y := range aliens.aliens {
			if rl.CheckCollisionRecs(bullets.Bullets[i].rect, aliens.aliens[y].rect) {
				fmt.Println(aliens.aliens[y])
				if bullets.Bullets[i].enabled && aliens.aliens[y].enabled {
					bullets.Bullets[i].enabled = false
					aliens.aliens[y].enabled = false
					servers--
					go k8s.UpdateDeployment(deploymentsClient, int32(servers))
					if servers == 0 {
						gameOver = true
					}
				}
			}
		}
		bullets.Bullets[i].rect.Y -= float32(bulletSpeed)
	}
	for i := range aliens.aliens {
		aliens.aliens[i].rect.X += float32(alienSpeedX)
	}
	for i := range aliens.aliens {
		if aliens.aliens[i].enabled {
			aliens.minLeft = int(aliens.aliens[i].rect.X)
			break
		}
	}
	for i := len(aliens.aliens) - 1; i >= 0; i-- {
		if aliens.aliens[i].enabled {
			aliens.maxRight = int(aliens.aliens[i].rect.X)
			break
		}
	}
	if aliens.minLeft < 0 {
		alienSpeedX = int(math.Abs(float64(alienSpeedX)))
		for i := range aliens.aliens {
			aliens.aliens[i].rect.Y += float32(alienSpeedY)
		}
	}
	if aliens.maxRight+characterWidth > screenWidth {
		for i := range aliens.aliens {
			aliens.aliens[i].rect.Y += float32(alienSpeedY)
		}
		alienSpeedX = int(math.Abs(float64(alienSpeedX))) * -1
	}
	for i := range aliens.aliens {
		if rl.CheckCollisionRecs(player.rect, aliens.aliens[i].rect) {
			gameOver = true
		}
	}
}
