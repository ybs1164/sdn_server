package games

type IMagic interface {
	Run(*Player, float64, float64)
}

type DamageMagic struct {
	IMagic
	damage   uint32
	distance uint32
}

func NewDamageMagic(damage, distance uint32) *DamageMagic {
	dm := DamageMagic{}
	dm.damage = damage
	dm.distance = distance
	return &dm
}

func (d DamageMagic) Run(p *Player, x, y float64) {
	var g *Game = p.game
	var unitList []IUnit = g.Collision(x-float64(d.distance), y, float64(d.distance)*2, float64(d.distance))
	for _, u := range unitList {
		u.GetDamage(d.damage, p.team)
	}
}

type HealMagic struct {
	IMagic
	heal     uint32
	distance uint32
}

func NewHealMagic(heal, distance uint32) *HealMagic {
	hm := HealMagic{}
	hm.heal = heal
	hm.distance = distance
	return &hm
}

func (h HealMagic) Run(p *Player, x, y float64) {
	var g *Game = p.game
	var unitList []IUnit = g.Collision(x-float64(h.distance), y, float64(h.distance)*2, float64(h.distance))
	for _, u := range unitList {
		u.GetHeal(h.heal, p.team)
	}
}

type CreditCard struct {
	IMagic
}

func NewCreditCardMagic() *CreditCard {
	cc := CreditCard{}
	return &cc
}

func (c CreditCard) Run(p *Player, x, y float64) {

}

type Clip struct {
	IMagic
	order byte
}

func NewClipMagic() *Clip {
	cp := Clip{}
	return &cp
}

func (c Clip) Run(p *Player, x, y float64) {

}
