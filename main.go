package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/nsf/termbox-go"
)

// GameConfig содержит настройки игры
type GameConfig struct {
	PlayerX           int           // Позиция игрока по X
	GroundY           int           // Позиция земли
	JumpHeight        int           // Высота прыжка
	JumpDuration      int           // Продолжительность прыжка
	InitialLives      int           // Начальное количество жизней
	BaseObstacleSpeed int           // Базовая скорость препятствий
	FrameRate         time.Duration // Частота обновления кадров
	BaseSpawnRate     int           // Базовая частота появления препятствий
	ScreenWidth       int           // Ширина экрана
	MaxSpeed          int           // Максимальная скорость препятствий
	MinSpawnRate      int           // Минимальная частота появления препятствий
}

// DefaultGameConfig возвращает конфигурацию игры по умолчанию
func DefaultGameConfig() GameConfig {
	return GameConfig{
		PlayerX:           10,
		GroundY:           15,
		JumpHeight:        10,
		JumpDuration:      18,
		InitialLives:      5,
		BaseObstacleSpeed: 1,
		FrameRate:         32 * time.Millisecond,
		BaseSpawnRate:     25,
		ScreenWidth:       80,
		MaxSpeed:          5,
		MinSpawnRate:      10,
	}
}

// Типы препятствий
const (
	ObstacleRock = iota
	ObstacleBox
	ObstacleTree
	ObstacleTypesCount
)

// Sprite представляет графическое изображение объекта
type Sprite []string

// Obstacle представляет игровое препятствие
type Obstacle struct {
	X    int
	Type int
}

// Game содержит игровое состояние
type Game struct {
	Config     GameConfig
	PlayerY    int
	IsJumping  bool
	JumpTime   int
	Lives      int
	Obstacles  []Obstacle
	FrameCount int
	Score      int
	Sprites    struct {
		Player    Sprite
		Obstacles []Sprite
	}
}

// LevelConfig содержит настройки сложности текущего уровня
type LevelConfig struct {
	ObstacleSpeed int
	SpawnRate     int
}

// NewGame создаёт новую игру с указанной конфигурацией
func NewGame(config GameConfig) *Game {
	game := &Game{
		Config:     config,
		PlayerY:    config.GroundY,
		IsJumping:  false,
		JumpTime:   0,
		Lives:      config.InitialLives,
		Obstacles:  []Obstacle{},
		FrameCount: 0,
		Score:      0,
	}

	// Инициализация спрайтов
	game.Sprites.Player = Sprite{
		" O ",
		"/|\\",
		"/ \\",
	}

	game.Sprites.Obstacles = []Sprite{
		// Камень
		{
			" /\\ ",
			"/__\\",
		},
		// Ящик
		{
			"+--+",
			"|  |",
			"+--+",
		},
		// Дерево
		{
			" /\\ ",
			"/  \\",
			" || ",
			" || ",
		},
	}

	return game
}

// GetLevelConfig возвращает настройки сложности на основе текущего счета
func (g *Game) GetLevelConfig() LevelConfig {
	// Базовые значения
	speed := g.Config.BaseObstacleSpeed
	spawnRate := g.Config.BaseSpawnRate

	// Увеличиваем сложность каждые 10 очков
	levelIncrease := g.Score / 10

	// Увеличиваем скорость (до максимума)
	speed += levelIncrease
	if speed > g.Config.MaxSpeed {
		speed = g.Config.MaxSpeed
	}

	// Уменьшаем частоту появления препятствий (делаем их чаще)
	spawnRate -= levelIncrease
	if spawnRate < g.Config.MinSpawnRate {
		spawnRate = g.Config.MinSpawnRate
	}

	return LevelConfig{
		ObstacleSpeed: speed,
		SpawnRate:     spawnRate,
	}
}

// DrawText выводит текст на заданной позиции
func DrawText(x, y int, msg string, fg, bg termbox.Attribute) {
	for i, ch := range msg {
		termbox.SetCell(x+i, y, ch, fg, bg)
	}
}

// DrawSprite выводит многострочный спрайт на экран
func DrawSprite(x, y int, sprite Sprite, fg, bg termbox.Attribute) {
	for dy, line := range sprite {
		for dx, ch := range line {
			termbox.SetCell(x+dx, y+dy, rune(ch), fg, bg)
		}
	}
}

// Render отрисовывает текущее состояние игры
func (g *Game) Render() {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)

	// Отрисовка земли
	for x := 0; x < g.Config.ScreenWidth; x++ {
		termbox.SetCell(x, g.Config.GroundY+1, '_', termbox.ColorGreen, termbox.ColorDefault)
	}

	// Отрисовка игрока
	playerY := g.PlayerY - len(g.Sprites.Player) + 1
	DrawSprite(g.Config.PlayerX, playerY, g.Sprites.Player, termbox.ColorYellow, termbox.ColorDefault)

	// Отрисовка препятствий
	for _, o := range g.Obstacles {
		model := g.Sprites.Obstacles[o.Type]
		y := g.Config.GroundY - len(model) + 1
		DrawSprite(o.X, y, model, termbox.ColorRed, termbox.ColorDefault)
	}

	// Отрисовка информации
	levelConfig := g.GetLevelConfig()
	info := fmt.Sprintf("Lives: %d | Score: %d | Speed: %d", g.Lives, g.Score, levelConfig.ObstacleSpeed)
	DrawText(0, 0, info, termbox.ColorWhite, termbox.ColorDefault)

	// Отрисовка инструкций
	instructions := "Space: Jump | ESC/Q: Quit"
	DrawText(g.Config.ScreenWidth-len(instructions), 0, instructions, termbox.ColorWhite, termbox.ColorDefault)

	termbox.Flush()
}

// Update обновляет состояние игры
func (g *Game) Update() {
	// Получаем текущую конфигурацию уровня
	levelConfig := g.GetLevelConfig()

	// Обработка прыжка
	if g.IsJumping {
		g.JumpTime++
		if g.JumpTime < g.Config.JumpDuration/2 {
			// Подъем
			jumpProgress := float64(g.JumpTime) / float64(g.Config.JumpDuration/2)
			g.PlayerY = g.Config.GroundY - int(float64(g.Config.JumpHeight)*jumpProgress)
		} else if g.JumpTime < g.Config.JumpDuration {
			// Спуск
			fallProgress := float64(g.JumpTime-g.Config.JumpDuration/2) / float64(g.Config.JumpDuration/2)
			g.PlayerY = g.Config.GroundY - g.Config.JumpHeight + int(float64(g.Config.JumpHeight)*fallProgress)
		} else {
			// Приземление
			g.PlayerY = g.Config.GroundY
			g.IsJumping = false
			g.JumpTime = 0
		}
	}

	// Обновление препятствий
	newObstacles := []Obstacle{}
	for _, o := range g.Obstacles {
		o.X -= levelConfig.ObstacleSpeed

		// Проверка столкновений
		obstacleWidth := len(g.Sprites.Obstacles[o.Type][0])
		obstacleHeight := len(g.Sprites.Obstacles[o.Type])
		playerWidth := len(g.Sprites.Player[0])
		playerHeight := len(g.Sprites.Player)

		// Более точное определение столкновений
		if CheckCollision(
			g.Config.PlayerX, g.PlayerY-playerHeight+1, playerWidth, playerHeight,
			o.X, g.Config.GroundY-obstacleHeight+1, obstacleWidth, obstacleHeight,
		) {
			g.Lives--
		}

		if o.X > -obstacleWidth {
			newObstacles = append(newObstacles, o)
		} else {
			// Увеличение счета при успешном пропуске препятствия
			g.Score++
		}
	}
	g.Obstacles = newObstacles

	// Создание новых препятствий
	if g.FrameCount%levelConfig.SpawnRate == 0 && rand.Intn(3) > 0 {
		obstacleType := rand.Intn(ObstacleTypesCount)
		g.Obstacles = append(g.Obstacles, Obstacle{
			X:    g.Config.ScreenWidth,
			Type: obstacleType,
		})
	}

	g.FrameCount++
}

// CheckCollision проверяет столкновение двух прямоугольников
func CheckCollision(x1, y1, w1, h1, x2, y2, w2, h2 int) bool {
	return x1 < x2+w2 && x1+w1 > x2 && y1 < y2+h2 && y1+h1 > y2
}

// HandleInput обрабатывает пользовательский ввод
func (g *Game) HandleInput(ev termbox.Event) bool {
	if ev.Type == termbox.EventKey {
		switch {
		case ev.Key == termbox.KeyEsc || ev.Ch == 'q':
			return false
		case ev.Key == termbox.KeySpace && !g.IsJumping:
			g.IsJumping = true
		}
	}
	return true
}

// DrawGameOver отображает экран окончания игры
func DrawGameOver(score int) {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	width, height := termbox.Size()

	// Графика "GAME OVER"
	gameOverArt := []string{
		"  ___    _    __  __   ___    _____   _   _  ___  ___ ",
		" / __|  /_\\  |  \\/  | | __|  / _ \\ \\ | | | |/ _ \\| _ \\",
		"| (_ | / _ \\ | |\\/| | | _|  |  __/\\ V /  | |  __/|   /",
		" \\___|/_/ \\_\\|_|  |_| |___|  \\___| \\_/   |_|\\___||_|_\\",
	}

	// Рисуем ASCII-арт
	artY := height/2 - len(gameOverArt) - 2
	for i, line := range gameOverArt {
		DrawText(width/2-len(line)/2, artY+i, line, termbox.ColorRed, termbox.ColorDefault)
	}

	// Сообщение о завершении игры
	finalScore := fmt.Sprintf("Final Score: %d", score)
	exitMsg := "Press any key to exit"

	DrawText(width/2-len(finalScore)/2, height/2+3, finalScore, termbox.ColorYellow, termbox.ColorDefault)
	DrawText(width/2-len(exitMsg)/2, height/2+5, exitMsg, termbox.ColorWhite, termbox.ColorDefault)

	termbox.Flush()

	// Ожидание нажатия клавиши
	termbox.PollEvent()
}

// RunGame запускает игровой цикл
func RunGame(game *Game) {
	// Создание игрового цикла
	gameLoop := time.NewTicker(game.Config.FrameRate)
	defer gameLoop.Stop()

	// Канал для событий пользовательского ввода
	eventQueue := make(chan termbox.Event)
	go func() {
		for {
			eventQueue <- termbox.PollEvent()
		}
	}()

	// Главный цикл игры
	running := true
	for running && game.Lives > 0 {
		select {
		case <-gameLoop.C:
			game.Update()
			game.Render()
		case ev := <-eventQueue:
			running = game.HandleInput(ev)
		}
	}

	// Отображение экрана завершения игры
	DrawGameOver(game.Score)
}

func main() {
	// Инициализация генератора случайных чисел
	rand.Seed(time.Now().UnixNano())

	// Инициализация termbox
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	// Настройка событий
	termbox.SetInputMode(termbox.InputEsc)

	// Получение размеров терминала
	width, _ := termbox.Size()

	// Настройка конфигурации с учетом размера экрана
	config := DefaultGameConfig()
	config.ScreenWidth = width

	// Создание и запуск игры
	game := NewGame(config)
	RunGame(game)
}
