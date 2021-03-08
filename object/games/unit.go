package games

import (
	"encoding/binary"
	"log"
	"math"

	quadtree "github.com/ybs1164/quadtree-go"
)

type IUnit interface {
	Run(*Player, uint16)
	Frame()
	HitBox() *quadtree.Bounds
	Team() byte
	Collision([]IUnit)
	GetDamage(uint32, byte)
	GetHeal(uint32, byte)
	IsPoisoned() bool
	Poisoned()
	Move(float64, float64)
	Data() []byte
	IsDead() bool
	Death()
}

// todo : Unit Status
type Unit struct {
	quadtree.Bounds
	owner *Player

	id     uint16
	team   byte
	typeid uint16

	health    uint32
	poison    uint32
	maxHealth uint32

	isVisible bool
	isStatic  bool
	isNoAI    bool
	isReverse bool
}

func (u *Unit) HitBox() *quadtree.Bounds {
	return &u.Bounds
}

func (u Unit) Team() byte {
	return u.team
}

func (u *Unit) Collision(list []IUnit) {
	return
}

func (u *Unit) GetDamage(d uint32, t byte) {
	if t == u.team {
		return
	}

	if t == 2 {
		u.poison += d
	} else if u.health > 0 {
		if u.health < d {
			u.health = 0
		} else {
			u.health -= d
		}
		if u.team == 2 {
			u.poison = u.health
		}
	}
}

func (u *Unit) GetHeal(h uint32, t byte) {
	if t != u.team {
		return
	}

	u.health += h
	if u.health > u.maxHealth {
		u.health = u.maxHealth
	}
	if u.team == 2 {
		u.poison = u.health
	}
}

func (u Unit) IsPoisoned() bool {
	return u.health <= u.poison
}

func (u *Unit) Poisoned() {
	u.owner = u.owner.game.players[2]
	u.team = 2
	u.poison = u.health
}

// todo : round
func (u *Unit) Move(x, y float64) {
	u.X += x
	u.Y += y
}

func (u *Unit) Run(p *Player, id uint16) {
	u.owner = p
	u.id = id
	u.team = p.team
	if u.team == 2 {
		u.poison = u.health
	}
	u.isVisible = true
	return
}

func (u *Unit) Frame() {
	if math.Abs(u.X) > 80 || math.Abs(u.Y) > 80 {
		u.health = 0
	}
	return
}

func (u *Unit) Death() {
	var data []byte = make([]byte, 3)
	data[0] = 6
	binary.BigEndian.PutUint16(data[1:3], u.id)
	u.owner.game.Broadcast(data)
	log.Println("dead unit id : ", u.id)
	return
}

func (u *Unit) IsDead() bool {
	return u.health <= 0
}

func (u Unit) Data() []byte {
	if !u.isVisible {
		return []byte{}
	}
	var data []byte = make([]byte, 34)
	binary.BigEndian.PutUint16(data[0:2], u.id)
	binary.BigEndian.PutUint16(data[2:4], u.typeid)
	data[4] = u.team
	// todo : pos set
	binary.BigEndian.PutUint64(data[5:13], math.Float64bits(u.X))
	binary.BigEndian.PutUint64(data[13:21], math.Float64bits(u.Y))
	binary.BigEndian.PutUint32(data[21:25], u.health)
	binary.BigEndian.PutUint32(data[25:29], u.maxHealth)
	binary.BigEndian.PutUint32(data[29:33], u.poison)
	data[33] = u.Status()
	return data
}

func (u Unit) Status() byte {
	var data byte = 0
	if u.isReverse {
		data += 1
	}
	return data
}

/*
	Unit List
*/
// Earth
type Earth struct {
	Unit
}

func NewEarth() *Earth {
	unit := Earth{}
	unit.typeid = 5

	unit.maxHealth = 5000
	unit.health = 5000

	unit.X = -3

	unit.Width = 8.25
	unit.Height = 4.56

	return &unit
}

func (e *Earth) Death() {
	e.Unit.Death()
	e.owner.game.End(1)
}

// Flask
type Flask struct {
	Unit
	time int
}

func NewFlask() *Flask {
	unit := Flask{}
	unit.typeid = 0

	unit.maxHealth = 350
	unit.health = 350

	unit.Width = 1.62
	unit.Height = 2.71

	unit.time = 60 * 5

	return &unit
}

func (f *Flask) Frame() {
	f.Unit.Frame()
	f.time--
	if f.time <= 0 {
		f.owner.GetEnergy(1)
		f.time = 60 * 5
	}
}

// Note
type Note struct {
	Unit
	subEnergy uint16
}

func NewNote() *Note {
	unit := Note{}
	unit.typeid = 1

	unit.maxHealth = 100
	unit.health = 100

	unit.Width = 2.25
	unit.Height = 2.92

	unit.subEnergy = 1

	return &unit
}

func (n *Note) Run(p *Player, id uint16) {
	n.Unit.Run(p, id)
	for i := range p.deck {
		c := &p.deck[i]
		c.SubCost(n.subEnergy)
	}
}

func (n *Note) Poisoned() {
	for i := range n.owner.deck {
		c := &n.owner.deck[i]
		c.AddCost(n.subEnergy)
	}
	n.Unit.Poisoned()
	for i := range n.owner.deck {
		c := &n.owner.deck[i]
		c.SubCost(n.subEnergy)
	}
}

func (n *Note) Death() {
	n.Unit.Death()
	for i := range n.owner.deck {
		c := &n.owner.deck[i]
		c.AddCost(n.subEnergy)
	}
}

// Bag
type Bag struct {
	Unit
	addEnergy uint16
}

func NewBag() *Bag {
	unit := Bag{}
	unit.typeid = 2

	unit.maxHealth = 150
	unit.health = 150

	unit.Width = 3.07
	unit.Height = 3.59

	unit.addEnergy = 1

	return &unit
}

func (b *Bag) Run(p *Player, id uint16) {
	b.Unit.Run(p, id)
	p.maxEnergy += b.addEnergy
}

func (b *Bag) Poisoned() {
	b.owner.maxEnergy -= b.addEnergy
	b.Unit.Poisoned()
	b.owner.maxEnergy += b.addEnergy
}

func (b *Bag) Death() {
	b.Unit.Death()
	b.owner.maxEnergy -= b.addEnergy
}

//
type Pen struct {
	Unit
}

func NewPen() *Pen {
	unit := Pen{}
	unit.typeid = 3

	unit.maxHealth = 500
	unit.health = 500

	unit.Width = 1
	unit.Height = 1

	return &unit
}

//
type BigPencil struct {
	Unit
	mcooltime uint8 // max CoolTime
	cooltime  uint8
	waittime  uint8

	damage   uint32
	distance float64

	target *IUnit
}

func NewBigPencil() *BigPencil {
	unit := BigPencil{}
	unit.typeid = 4

	unit.maxHealth = 1000
	unit.health = 1000

	unit.Width = 6.52
	unit.Height = 1.49

	unit.damage = 250

	unit.mcooltime = 120
	unit.waittime = 30

	unit.distance = 1

	return &unit
}

func (b *BigPencil) Run(p *Player, id uint16) {
	b.Unit.Run(p, id)
	switch b.team {
	case 0:
		b.isReverse = b.X > 0
	default:
		b.isReverse = b.X < 0
	}
	b.Y = 0.5
}

func (b *BigPencil) Frame() {
	b.Unit.Frame()

	var detectBound quadtree.Bounds

	if b.isReverse {
		detectBound = quadtree.Bounds{
			X:      b.X + b.Width,
			Y:      b.Y,
			Width:  b.distance,
			Height: b.Height,
		}
	} else {
		detectBound = quadtree.Bounds{
			X:      b.X - b.distance,
			Y:      b.Y,
			Width:  b.distance,
			Height: b.Height,
		}
	}

	if b.target == nil {
		var colList []IUnit = b.owner.game.Collision(
			detectBound.X,
			detectBound.Y,
			detectBound.Width,
			detectBound.Height,
		)

		for _, u := range colList {
			if u.Team() != b.team {
				b.target = &u
				break
			}
		}

		if b.isReverse {
			b.Move(0.01, 0)
		} else {
			b.Move(-0.01, 0)
		}
	}
	if b.target != nil {
		if !(*b.target).IsDead() && b.team != (*b.target).Team() && (*b.target).HitBox().Intersects(detectBound) {
			if b.cooltime == 0 {
				m := NewDamageMagic(b.damage, 3)
				if b.isReverse {
					m.Run(b.owner, b.X+b.Width, b.Y)
				} else {
					m.Run(b.owner, b.X, b.Y)
				}
				b.cooltime = b.mcooltime
			}
		} else {
			b.target = nil
		}
	}
	if b.cooltime > 0 {
		b.cooltime--
	}
}

//
type Sharpener struct {
	Unit
	mcooltime uint8 // max CoolTime
	cooltime  uint8
	waittime  uint8

	distance float64
	speed    float64
	damage   uint32

	target *IUnit
}

func NewSharpener() *Sharpener {
	unit := Sharpener{}
	unit.typeid = 6

	unit.maxHealth = 700
	unit.health = 700

	unit.Width = 3.49
	unit.Height = 2.34

	unit.distance = 15
	unit.speed = .5
	unit.damage = 80

	unit.mcooltime = 60
	//unit.waittime = 30

	return &unit
}

func (s *Sharpener) Run(p *Player, id uint16) {
	s.Unit.Run(p, id)
	switch s.team {
	case 0:
		s.isReverse = s.X > 0
	default:
		s.isReverse = s.X < 0
	}
}

// todo : 두개 이상의 중복되는 유닛들이 따로 작동되도록 제작
func (s *Sharpener) Frame() {
	s.Unit.Frame()

	var detectBound quadtree.Bounds

	if s.isReverse {
		detectBound = quadtree.Bounds{
			X:      s.X + s.Width,
			Y:      s.Y,
			Width:  s.distance,
			Height: s.distance,
		}
	} else {
		detectBound = quadtree.Bounds{
			X:      s.X - s.distance,
			Y:      s.Y,
			Width:  s.distance,
			Height: s.distance,
		}
	}

	if s.cooltime == 0 {
		if s.target == nil {
			var colList []IUnit = s.owner.game.Collision(
				detectBound.X,
				detectBound.Y,
				detectBound.Width,
				detectBound.Height,
			)

			for _, u := range colList {
				if u.Team() != s.team {
					s.target = &u
					break
				}
			}
		}
		if s.target != nil {
			if !(*s.target).IsDead() && s.team != (*s.target).Team() && (*s.target).HitBox().Intersects(detectBound) {
				var dx float64 = -1
				obj := NewBullet(s.damage, s.speed, 300, 12)
				obj.angle = math.Pi
				if s.isReverse {
					dx = s.Width + 1
					obj.angle = 0
				}
				s.owner.game.Spawn(Spawn{
					owner: s.owner,
					x:     s.X + dx,
					y:     s.Y + 1.5,
					obj:   obj,
				})
				s.cooltime = s.mcooltime
			} else {
				s.target = nil
			}
		}
	}
	if s.cooltime > 0 {
		s.cooltime--
	}
}

//
type Alarm struct {
	Unit

	damage   uint32
	distance uint32
}

func NewAlarm() *Alarm {
	unit := Alarm{}
	unit.typeid = 7

	unit.maxHealth = 200
	unit.health = 200

	unit.Width = 1.69
	unit.Height = 1.96

	unit.damage = 1
	unit.distance = 5

	return &unit
}

func (a *Alarm) Frame() {
	a.Unit.Frame()

	dm := *NewDamageMagic(a.damage, a.distance)
	dm.Run(a.owner, a.X+a.Width/2, a.Y)
}

//
type Dictionary struct {
	Unit

	healPercent float64
}

func NewDictionary() *Dictionary {
	unit := Dictionary{}
	unit.typeid = 8

	unit.maxHealth = 2000
	unit.health = 2000

	unit.Width = 1.06
	unit.Height = 3.04

	unit.healPercent = 0.2

	return &unit
}

func (dic *Dictionary) GetDamage(d uint32, t byte) {
	m := NewHealMagic(uint32(float64(d)*dic.healPercent), 10)
	m.Run(dic.owner, dic.X+dic.Width/2, dic.Y)

	dic.Unit.GetDamage(d, t)
}

type PaintBrush struct {
	Unit
	mcooltime uint8 // max CoolTime
	cooltime  uint8
	waittime  uint8

	heal     uint32
	distance float64

	target *IUnit
}

func NewPaintBrush() *PaintBrush {
	unit := PaintBrush{}
	unit.typeid = 9

	unit.maxHealth = 400
	unit.health = 400

	unit.Width = 4.08
	unit.Height = 1.29

	unit.mcooltime = 30

	unit.heal = 50
	unit.distance = 2

	return &unit
}

func (pb *PaintBrush) Run(p *Player, id uint16) {
	pb.Unit.Run(p, id)
	switch pb.team {
	case 0:
		pb.isReverse = pb.X > 0
	default:
		pb.isReverse = pb.X < 0
	}
	pb.Y = 0.5
}

func (p *PaintBrush) Frame() {
	p.Unit.Frame()

	var detectBound quadtree.Bounds

	if p.isReverse {
		detectBound = quadtree.Bounds{
			X:      p.X + p.Width,
			Y:      p.Y,
			Width:  p.distance,
			Height: p.Height,
		}
	} else {
		detectBound = quadtree.Bounds{
			X:      p.X - p.distance,
			Y:      p.Y,
			Width:  p.distance,
			Height: p.Height,
		}
	}

	if p.target == nil {
		var colList []IUnit = p.owner.game.Collision(
			detectBound.X,
			detectBound.Y,
			detectBound.Width,
			detectBound.Height,
		)

		for _, u := range colList {
			if u != p && u.Team() == p.team {
				p.target = &u
				break
			}
		}

		if p.isReverse {
			p.Move(0.03, 0)
		} else {
			p.Move(-0.03, 0)
		}
	}
	if p.target != nil {
		if !(*p.target).IsDead() && p.team == (*p.target).Team() && (*p.target).HitBox().Intersects(detectBound) {
			if p.cooltime == 0 {
				(*p.target).GetHeal(p.heal, p.team)
				/*m := NewHealMagic(p.heal, 4)
				if p.isReverse {
					m.Run(p.owner, p.bounds.X+p.bounds.Width, p.bounds.Y)
				} else {
					m.Run(p.owner, p.bounds.X, p.bounds.Y)
				}*/
				p.cooltime = p.mcooltime
			}
		} else {
			p.target = nil
		}
	}
	if p.cooltime > 0 {
		p.cooltime--
	}
}
