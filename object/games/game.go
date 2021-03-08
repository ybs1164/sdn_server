package games

import (
	"encoding/binary"
	"log"
	"math"
	"math/rand"
	"net"
	"sync"
	"time"

	quadtree "github.com/ybs1164/quadtree-go"
)

type Spawn struct {
	owner *Player
	x     float64
	y     float64
	time  uint8

	obj interface{}
}

type Game struct {
	id     uint64
	playID uint64 // this ID is using for DB

	players     [3]*Player
	PlayerCount int

	status      byte // 0: ready.. 1: gaming
	frame       uint
	energySpeed uint16

	objID       uint16
	units       []IUnit
	projectiles []IProjectile
	qt          *quadtree.Quadtree
	ticker      *time.Ticker
	spawner     []Spawn
	mutex       sync.Mutex
}

func NewGame() *Game {
	rand.Seed(time.Now().UnixNano())

	g := new(Game)
	g.qt = &quadtree.Quadtree{
		Bounds: quadtree.Bounds{
			X:      0,
			Y:      0,
			Width:  100,
			Height: 100,
		},
		MaxObjects: 5,
		MaxLevels:  6,
		Level:      0,
		Objects:    make([]quadtree.IBounds, 0),
		Nodes:      make([]quadtree.Quadtree, 0),
	}
	g.units = []IUnit{}
	g.mutex = sync.Mutex{}
	g.ticker = time.NewTicker(time.Second / 60)
	//defer ticker.Stop()
	go g.Frame()
	return g
}

func (g *Game) Join(c net.Conn) {
	g.mutex.Lock()
	p := PlayerSet(g, c)
	for i := 0; i < 3; i++ {
		if g.players[i] == nil {
			g.players[i] = p
			break
		}
	}
	g.PlayerCount++
	g.mutex.Unlock()

	go p.ConnHandler()
}

func (g *Game) Left(p *Player) {
	g.mutex.Lock()
	for i := 0; i < g.PlayerCount; i++ {
		if i > 0 && g.players[i-1] == nil {
			g.players[i-1] = g.players[i]
		}
		if g.players[i] == p {
			g.players[i] = nil
		}
	}
	g.PlayerCount--
	g.mutex.Unlock()

	// todo : leave event

	log.Println("left " + p.name)
}

func (g *Game) Start() {
	/*
		if g.PlayerCount < 3 {
			g.Broadcast(append([]byte{2}, string("플레이어가 3명 모여야 합니다.")...))
			return
		}

		teamOn := [3]bool{false, false, false}

		for _, p := range g.players {
			if teamOn[p.team] {
				g.Broadcast(append([]byte{2}, string("팀은 각각 하나씩 가져야 합니다.")...))
				return
			} else {
				teamOn[p.team] = true
			}
		}

		sort.Slice(g.players[:], func(i, j int) bool {
			return g.players[i].team < g.players[j].team
		})
	*/

	for i := 0; i < g.PlayerCount; i++ {
		p := g.players[i]
		p.order = [8]byte{0, 1, 2, 3, 4, 5, 6, 7}
		//rand.Shuffle(len(p.order), func(i, j int) { p.order[i], p.order[j] = p.order[j], p.order[i] })
		p.energy = 5
		p.energyTime = 120
		p.maxEnergy = 10
	}
	earth := NewEarth()
	earth.Run(g.players[0], 0)
	g.units = []IUnit{earth}

	g.energySpeed = 1
	g.status = 1
	g.frame = 60 * 180

	g.objID = 1

	g.Broadcast([]byte{1})

	log.Println("Game Start")
}

func (g *Game) End(team byte) {
	g.units = []IUnit{}
	g.projectiles = []IProjectile{}

	var data []byte

	data = append(data, 7)
	data = append(data, team)

	g.Broadcast(data)

	g.status = 0

	log.Println("Game End")
}

func (g *Game) Broadcast(data []byte) {
	for i := 0; i < g.PlayerCount; i++ {
		p := g.players[i]
		p.Send(data)
	}
}

func (g *Game) Spawn(u Spawn) {
	g.spawner = append(g.spawner, u)
}

func (g *Game) Collision(x, y, w, h float64) []IUnit {
	var b quadtree.IBounds = &quadtree.Bounds{
		X:      x,
		Y:      y,
		Width:  w,
		Height: h,
	}
	var list []IUnit

	/*
		for _, unit := range g.units {
			if unit.HitBox().Intersects(b) {
				list = append(list, unit)
			}
		}
		return list
	*/

	for _, u := range g.qt.Retrieve(b) {
		list = append(list, u.(IUnit))
	}

	return list
}

func (g *Game) Frame() {
	second := 60
	for range g.ticker.C {
		g.mutex.Lock()

		// ping
		if second < 1 {
			for i := 0; i < g.PlayerCount; i++ {
				g.players[i].lastTime = append(g.players[i].lastTime, time.Now().UnixNano())
			}
			g.Broadcast([]byte{5})
			second = 60
		}
		second--

		var data []byte

		data = append(data, 0)

		for i := 0; i < g.PlayerCount; i++ {
			data = append(data, g.players[i].Data()...)
		}

		g.Broadcast(data)

		// game
		switch g.status {
		case 0: // room
		case 1: // game
			// spawn
			for i := 0; i < len(g.spawner); i++ {
				s := &g.spawner[i]
				if s.time > 0 {
					s.time--
				} else {
					owner := s.owner
					x, y := s.x, s.y
					u := s.obj
					switch u.(type) {
					case IUnit:
						// todo : unit waiting time
						unit := u.(IUnit)
						g.units = append(g.units, unit)
						unit.HitBox().X = x - unit.HitBox().Width/2
						unit.HitBox().Y = y
						unit.Run(owner, g.objID)
					case IProjectile:
						proj := u.(IProjectile)
						g.projectiles = append(g.projectiles, proj)
						proj.HitBox().X = x - proj.HitBox().Width/2
						proj.HitBox().Y = y
						proj.Run(owner, g.objID)
					case IMagic:
						// todo
					default:
						log.Println("what?")
					}
					g.objID++
					g.spawner = append(g.spawner[:i], g.spawner[i+1:]...)
					i--

					var data []byte = make([]byte, 17)

					data[0] = 8

					binary.BigEndian.PutUint64(data[1:9], math.Float64bits(x))
					binary.BigEndian.PutUint64(data[9:17], math.Float64bits(y))

					owner.Send(data)
				}
			}

			// todo : quadtree

			g.qt.Clear()
			for _, unit := range g.units {
				var b quadtree.Bounds = *unit.HitBox()
				b.X -= b.Width / 2
				g.qt.Insert(unit)
			}

			for _, unit := range g.units {
				unit.Frame()
			}

			for _, p := range g.projectiles {
				p.Frame()
			}

			var isAllP = true

			for i := 0; i < len(g.units); i++ {
				unit := &g.units[i]
				if (*unit).Team() != 2 {
					isAllP = false
				}
				if (*unit).IsDead() {
					(*unit).Death()
					g.units = append(g.units[:i], g.units[i+1:]...)
					i--
				} else if (*unit).Team() != 2 && (*unit).IsPoisoned() {
					(*unit).Poisoned()
				}
			}

			for i := 0; i < len(g.projectiles); i++ {
				pr := &g.projectiles[i]
				if (*pr).IsUsing() {
					(*pr).Death()
					g.projectiles = append(g.projectiles[:i], g.projectiles[i+1:]...)
					i--
				}
			}

			// add energy to players
			for i := 0; i < g.PlayerCount; i++ {
				p := g.players[i]
				if p.energy < p.maxEnergy {
					if p.energyTime-g.energySpeed < 0 {
						p.energyTime = 0
					} else {
						p.energyTime = p.energyTime - g.energySpeed
					}

					if p.energyTime == 0 {
						p.GetEnergy(1)
						p.energyTime = 120
					}
				}
			}

			// send data
			for i := 0; i < g.PlayerCount; i++ {
				var data []byte
				p := g.players[i]

				data = append(data, 3)

				var value []byte = make([]byte, 2)
				binary.BigEndian.PutUint16(value, p.id)
				data = append(data, value...)
				binary.BigEndian.PutUint16(value, p.energy)
				data = append(data, value...)
				binary.BigEndian.PutUint16(value, p.maxEnergy)
				data = append(data, value...)

				// 경기 시간
				var time uint16 = uint16(g.frame / 60)
				binary.BigEndian.PutUint16(value, time)
				data = append(data, value...)
				//

				for _, value := range p.order {
					data = append(data, value)
				}
				for _, c := range p.deck {
					data = append(data, c.Data()...)
				}

				p.Send(data)
			}

			var data []byte

			data = append(data, 4)

			for _, unit := range g.units {
				data = append(data, unit.Data()...)
			}
			for _, proj := range g.projectiles {
				data = append(data, proj.Data()...)
			}

			g.Broadcast(data)

			g.frame--

			// semo win
			if isAllP {
				g.End(2)
			}

			// dongrami win
			if g.frame == 0 {
				g.End(0)
			}
		}
		g.mutex.Unlock()
	}
}
