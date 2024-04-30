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
	bulletBuffer := BulletBuffer{}
	swarm := AlienSwarm{}
	player := Player{rl.Rectangle{float32(screenWidth) / 2, float32(screenHeight) - 50, float32(characterWidth), float32(characterHeight)}}

	for i := range servers {
		alien := Alien{rl.Rectangle{float32(i*2*characterWidth + 10), float32(characterHeight) * 2, float32(characterWidth), float32(characterHeight)}, true}
		swarm.aliens = append(swarm.aliens, alien)
	}
	rl.InitWindow(int32(screenWidth), int32(screenHeight), "K8s_Invaders")
	defer rl.CloseWindow()
	rl.SetTargetFPS(20)
	k8s.CreateDeployment(deploymentsClient, deployment)
	go k8s.UpdateDeployment(deploymentsClient, int32(servers))
	for !rl.IsKeyPressed(rl.KeySpace) {
		rl.BeginDrawing()
		rl.ClearBackground(rl.Black)
		rl.DrawText("Press Space to Continue", 200, 10, 30, rl.White)
		rl.EndDrawing()
	}
	for !rl.WindowShouldClose() && !gameOver {
		checkinput(&player, &bulletBuffer)
		checkPhysics(&bulletBuffer, deploymentsClient, &player, &swarm)

		rl.BeginDrawing()
		rl.ClearBackground(rl.Black)
		replicaText := fmt.Sprintf("Replicas: %d", servers)
		rl.DrawText(replicaText, 10, 10, 20, rl.Red)
		for i := range swarm.aliens {
			if swarm.aliens[i].enabled {
				rl.DrawRectangleRec(swarm.aliens[i].rect, rl.Red)
			}
		}
		for i := range bulletBuffer.Bullets {
			rl.DrawRectangleRec(bulletBuffer.Bullets[i].rect, rl.Blue)
		}
		rl.DrawRectangleRec(player.rect, rl.White)
		rl.EndDrawing()
	}
	k8s.DeleteDeployment(deploymentsClient)
}

func checkinput(player *Player, bulletBuffer *BulletBuffer) {
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
		bulletBuffer.Bullets = append(bulletBuffer.Bullets, bullet)
	}
}

func checkPhysics(bulletBuffer *BulletBuffer, deploymentsClient v1.DeploymentInterface, player *Player, swarm *AlienSwarm) {
	for i := range bulletBuffer.Bullets {
		for y := range swarm.aliens {
			if rl.CheckCollisionRecs(bulletBuffer.Bullets[i].rect, swarm.aliens[y].rect) {
				fmt.Println(swarm.aliens[y])
				if bulletBuffer.Bullets[i].enabled && swarm.aliens[y].enabled {
					bulletBuffer.Bullets[i].enabled = false
					swarm.aliens[y].enabled = false
					servers--
					go k8s.UpdateDeployment(deploymentsClient, int32(servers))
					if servers == 0 {
						gameOver = true
					}
				}
			}
		}
		bulletBuffer.Bullets[i].rect.Y -= float32(bulletSpeed)
	}
	for i := range swarm.aliens {
		swarm.aliens[i].rect.X += float32(alienSpeedX)
		if rl.CheckCollisionRecs(player.rect, swarm.aliens[i].rect) {
			gameOver = true
		}
	}
	for i := range swarm.aliens {
		if swarm.aliens[i].enabled {
			swarm.minLeft = int(swarm.aliens[i].rect.X)
			break
		}
	}
	for i := len(swarm.aliens) - 1; i >= 0; i-- {
		if swarm.aliens[i].enabled {
			swarm.maxRight = int(swarm.aliens[i].rect.X)
			break
		}
	}
	if swarm.minLeft < 0 {
		alienSpeedX = int(math.Abs(float64(alienSpeedX)))
		for i := range swarm.aliens {
			swarm.aliens[i].rect.Y += float32(alienSpeedY)
		}
	}
	if swarm.maxRight+characterWidth > screenWidth {
		for i := range swarm.aliens {
			swarm.aliens[i].rect.Y += float32(alienSpeedY)
		}
		alienSpeedX = int(math.Abs(float64(alienSpeedX))) * -1
	}
}
