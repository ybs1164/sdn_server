package games

import (
	"encoding/binary"
	"log"
	"strconv"
)

type ICard interface {
	UseCard(uint16, int16) bool
	AddCost(uint16) uint16
	SubCost(uint16) uint16
	Data() []byte
}

type Card struct {
	id       uint32
	cost     uint16
	overcost uint16

	width       float64
	height      float64
	isCollision bool

	spawnSpeed uint16
	spawnList  []func() interface{}
}

func (c Card) UseCard(player *Player, cost uint16, x int16, time uint8) bool {
	if c.cost > cost {
		return false
	}
	if c.isCollision {
		switch player.team {
		case 0:
			if x < -50 || float64(x)+c.width > 50 {
				return false
			}
		case 1:
			if x > -50 && float64(x)+c.width < 50 {
				return false
			}
		case 2:
		default:
			log.Println("player team error : UseCard")
		}
	}
	log.Println(player.name + " use Card id : " + strconv.Itoa(int(c.id)))
	for _, s := range c.spawnList {
		o := s()
		player.game.Spawn(Spawn{
			owner: player,
			x:     float64(x),
			y:     0,
			time:  time, //+ i * c.spawnSpeed,
			obj:   o,
		})
	}
	return true
}

func (c *Card) AddCost(cost uint16) uint16 {
	if c.overcost > cost {
		c.overcost -= cost
		cost = 0
	} else {
		cost -= c.overcost
		c.overcost = 0
	}
	c.cost += cost
	return c.cost
}

func (c *Card) SubCost(cost uint16) uint16 {
	if c.cost < cost {
		c.overcost += cost - c.cost
		c.cost = 0
	} else {
		c.cost -= cost
	}
	return c.cost
}

func (c Card) Data() []byte {
	var data []byte = make([]byte, 6)
	binary.BigEndian.PutUint32(data, c.id)
	binary.BigEndian.PutUint16(data[4:], c.cost)
	return data
}

var CardList []Card = []Card{
	{},
	CardFlask(),
	CardNote(),
	CardBigPencil(),
	CardPaint(),
	CardErasers(),
	CardSharpener(),
	CardBag(),
	CardAlarm(),
	CardDictionary(),
	CardPaintBrush(),
}

func CardFlask() Card {
	flask := Card{}
	flask.id = 1
	flask.cost = 5
	flask.width = 1.62
	flask.height = 2.71
	flask.isCollision = true
	flask.spawnList = append(flask.spawnList, func() interface{} {
		return NewFlask()
	})
	return flask
}

func CardNote() Card {
	note := Card{}
	note.id = 2
	note.cost = 7
	note.width = 2.25
	note.height = 2.92
	note.isCollision = true
	note.spawnList = append(note.spawnList, func() interface{} {
		return NewNote()
	})
	return note
}

func CardBigPencil() Card {
	bigPencil := Card{}
	bigPencil.id = 3
	bigPencil.cost = 5
	bigPencil.width = 6.52
	bigPencil.height = 1.49
	bigPencil.isCollision = true
	bigPencil.spawnList = append(bigPencil.spawnList, func() interface{} {
		return NewBigPencil()
	})
	return bigPencil
}

func CardPaint() Card {
	hm := *NewHealMagic(100, 5)
	paint := Card{}
	paint.id = 4
	paint.cost = 3
	paint.spawnList = append(paint.spawnList, func() interface{} {
		return NewThrower(3, 10, hm)
	})
	return paint
}

func CardErasers() Card {
	dm := *NewDamageMagic(100, 4)
	erasers := Card{}
	erasers.id = 5
	erasers.cost = 2
	erasers.spawnList = append(erasers.spawnList, func() interface{} {
		return NewThrower(3, 11, dm)
	})
	erasers.spawnList = append(erasers.spawnList, func() interface{} {
		return NewThrower(4, 11, dm)
	})
	erasers.spawnList = append(erasers.spawnList, func() interface{} {
		return NewThrower(5, 11, dm)
	})
	return erasers
}

func CardSharpener() Card {
	sharpener := Card{}
	sharpener.id = 6
	sharpener.cost = 6
	sharpener.width = 3.49
	sharpener.height = 2.34
	sharpener.isCollision = true
	sharpener.spawnList = append(sharpener.spawnList, func() interface{} {
		return NewSharpener()
	})
	return sharpener
}

func CardBag() Card {
	bag := Card{}
	bag.id = 7
	bag.cost = 4
	bag.width = 3.07
	bag.height = 3.59
	bag.isCollision = true
	bag.spawnList = append(bag.spawnList, func() interface{} {
		return NewBag()
	})
	return bag
}

func CardAlarm() Card {
	alarm := Card{}
	alarm.id = 8
	alarm.cost = 2
	alarm.width = 1.69
	alarm.height = 1.96
	alarm.isCollision = true
	alarm.spawnList = append(alarm.spawnList, func() interface{} {
		return NewAlarm()
	})
	return alarm
}

func CardDictionary() Card {
	dictionary := Card{}
	dictionary.id = 9
	dictionary.cost = 8
	dictionary.width = 1.06
	dictionary.height = 3.04
	dictionary.isCollision = true
	dictionary.spawnList = append(dictionary.spawnList, func() interface{} {
		return NewDictionary()
	})
	return dictionary
}

func CardPaintBrush() Card {
	paintBrush := Card{}
	paintBrush.id = 10
	paintBrush.cost = 3
	paintBrush.width = 4.08
	paintBrush.height = 1.29
	paintBrush.isCollision = true
	paintBrush.spawnList = append(paintBrush.spawnList, func() interface{} {
		return NewPaintBrush()
	})
	return paintBrush
}
