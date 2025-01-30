package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla/websocket"
)

type GameState struct {
	Players    []Point
	Towers     []Point
	Explosions []Point
}

type Point struct {
	X int `json:"x"`
	Y int `json:"y"`
}

func main() {
	conn, _, err := websocket.DefaultDialer.Dial("wss://53ef-185-156-108-122.ngrok-free.app/ws", nil)
	if err != nil {
		log.Fatal("Ошибка при подключении к серверу:", err)
	}
	defer conn.Close()

	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			var state GameState
			if err := conn.ReadJSON(&state); err != nil {
				log.Println("Ошибка при чтении состояния:", err)
				return
			}
			renderGame(state)
		}
	}()

	// Обработка команд с клавиатуры
	go func() {
		for {
			var dx, dy int
			k := readKey()
			switch k {
			case 'w':
				dy = -1
			case 's':
				dy = 1
			case 'a':
				dx = -1
			case 'd':
				dx = 1
			case 'q':
				return
			default:
				continue
			}

			command := map[string]int{"dx": dx, "dy": dy}
			if err := conn.WriteJSON(command); err != nil {
				log.Println("Ошибка при отправке команды:", err)
				return
			}
		}
	}()

	// Ожидание сигнала завершения
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
}

// Отображение игрового поля
func renderGame(state GameState) {
	const width, height = 30, 15
	field := make([][]rune, height)
	for y := range field {
		field[y] = make([]rune, width)
		for x := range field[y] {
			field[y][x] = '.'
		}
	}

	for _, tower := range state.Towers {
		if tower.X >= 0 && tower.X < width && tower.Y >= 0 && tower.Y < height {
			field[tower.Y][tower.X] = '◘' // Игрок
		}
	}

	for _, explosion := range state.Explosions {
		if explosion.X >= 0 && explosion.X < width && explosion.Y >= 0 && explosion.Y < height {
			field[explosion.Y][explosion.X] = '💥' // Игрок
		}
	}

	for _, player := range state.Players {
		if player.X >= 0 && player.X < width && player.Y >= 0 && player.Y < height {
			field[player.Y][player.X] = '✈' // Игрок
		}
		//log.Printf("Игрок %s: (%d, %d)\n", id, player.X, player.Y)
	}

	for y := range field {
		for x := range field[y] {
			print(string(field[y][x]), " ")
		}
		println()
	}
	println()
}

// Чтение клавиши
func readKey() rune {
	var buf [1]byte
	_, err := os.Stdin.Read(buf[:])
	if err != nil {
		return 0
	}
	return rune(buf[0])
}
