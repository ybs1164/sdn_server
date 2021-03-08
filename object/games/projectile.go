package games

import (
	"encoding/binary"
	"math"

	quadtree "github.com/ybs1164/quadtree-go"
)

type IProjectile interface {
	HitBox() *quadtree.Bounds
	Run(*Player, uint16)
	Frame()
	SetTarget(*IUnit)
	IsUsing() bool
	Move(float64, float64)
	Death()
	Data() []byte
}

type Projectile struct {
	quadtree.Bounds
	owner  *Player
	id     uint16
	typeid uint16

	team  byte
	angle float64

	target *IUnit
	using  bool

	magic IMagic
}

func (p *Projectile) HitBox() *quadtree.Bounds {
	return &p.Bounds
}

// todo : round
func (p *Projectile) Move(x, y float64) {
	p.X += x
	p.Y += y
}

func (p *Projectile) Run(player *Player, id uint16) {
	p.owner = player
	p.id = id
	p.team = player.team
	p.using = false
	return
}

func (p *Projectile) Frame() {
	return
}

func (p *Projectile) SetTarget(u *IUnit) {
	p.target = u
}

func (p *Projectile) IsUsing() bool {
	return p.using
}

func (p Projectile) Data() []byte {
	var data []byte = make([]byte, 29)
	binary.BigEndian.PutUint16(data[0:2], p.id)
	binary.BigEndian.PutUint16(data[2:4], p.typeid)
	data[4] = p.team + 3
	// todo : pos set
	binary.BigEndian.PutUint64(data[5:13], math.Float64bits(p.X))
	binary.BigEndian.PutUint64(data[13:21], math.Float64bits(p.Y))
	binary.BigEndian.PutUint64(data[21:29], math.Float64bits(p.angle))
	return data
}

func (p *Projectile) Death() {
	if p.magic != nil {
		p.magic.Run(p.owner, p.X, p.Y)
	}
	var data []byte = make([]byte, 3)
	data[0] = 6
	binary.BigEndian.PutUint16(data[1:3], p.id)
	p.owner.game.Broadcast(data)
	return
}

type Thrower struct {
	Projectile
	height float64
}

func NewThrower(height float64, t uint16, m IMagic) *Thrower {
	th := Thrower{}
	th.height = height
	th.typeid = t
	th.magic = m
	return &th
}

func (t *Thrower) Run(player *Player, id uint16) {
	t.Projectile.Run(player, id)
	t.Y = t.height
}

func (t *Thrower) Frame() {
	t.Move(0, -0.1)
	if t.Y <= 0 {
		t.using = true
	}
}

// Pencil, Sharp
type Bullet struct {
	Projectile
	damage uint32
	speed  float64
	life   uint16
}

func NewBullet(damage uint32, speed float64, life uint16, t uint16) *Bullet {
	b := Bullet{}
	b.damage = damage
	b.speed = speed
	b.typeid = t
	b.life = life
	return &b
}

func (b *Bullet) Frame() {
	if b.target == nil {
		b.Move(math.Cos(b.angle)*b.speed, math.Sin(b.angle)*b.speed)

		for _, u := range b.owner.game.Collision(b.X-b.Width/2, b.Y-b.Height/2, b.Width, b.Height) {
			if u.Team() != b.team {
				u.GetDamage(b.damage, b.team)
				b.using = true
			}
		}
	} else {
		// todo : Move
	}
	b.life--
	if b.life == 0 {
		b.using = true
	}
}

type Crayon struct {
	Projectile
	damage float64
	speed  float64
	dx     float64
	dy     float64
}

func NewCrayon(damage, speed float64) *Crayon {
	c := Crayon{}
	c.damage = damage
	c.speed = speed
	c.typeid = 14
	return &c
}
