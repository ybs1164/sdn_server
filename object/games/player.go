package games

import (
	"encoding/binary"
	"io"
	"log"
	"net"
	"strconv"
	"time"
)

type Player struct {
	game *Game
	conn net.Conn
	id   uint16
	name string // todo : 데이터 크기 한정

	ping     int64
	lastTime []int64

	team       byte
	energy     uint16
	energyTime uint16
	maxEnergy  uint16
	deck       [8]Card // Card 는 꼭 8 개여야 하는가?

	order [8]uint8 // deck 의 순서결정. deck 자체의 인덱스를 바꿀 시 클라이언트에서 식별할만한 데이터가 없다.
}

func PlayerSet(game *Game, con net.Conn) *Player {
	p := Player{}
	p.game = game
	p.conn = con
	return &p
}

func (p *Player) ConnHandler() {
	if p.conn == nil {
		return
	}
	defer p.game.Left(p)
	defer p.conn.Close()

	recvBuf := make([]byte, 4096)
	for {
		n, err := p.conn.Read(recvBuf)
		if err != nil {
			if io.EOF == err {
				log.Println("Failed to Connect", err)
				return
			}
			log.Println(err)
			return
		}
		if n > 0 { // 근데 클라에서도 패킷 뭉침 현상이 일어나면 서버에서도 일어나야 하는게 아닌가?
			data := recvBuf[:n]
			dataType := data[0]

			p.game.mutex.Lock()
			switch dataType {
			case 0: // player join
				if p.game.status == 1 {
					break
				}
				p.name = string(data[1 : len(data)-1])
				log.Println(string(p.name) + " Join Server")
			case 1: // change team
				if p.game.status == 1 {
					break
				}
				if data[1] > 2 {
					break
				}
				p.team = data[1]
			case 2: // chat
				if p.game.status == 1 {
					break
				}
				p.Chatting(data[1 : len(data)-1]) // remove empty byte
			case 3: // use card
				if p.game.status == 0 {
					break
				}
				order := p.order[data[1]]
				x := int16(binary.BigEndian.Uint16(data[2:4]))
				if x > 100 && x < -100 {

				}

				waitframe := 80 - uint8(p.ping/int64(time.Millisecond)*60/1000)

				//log.Println(waitframe)

				//
				if p.deck[order].UseCard(p, p.energy, x, waitframe) { // using card -> change order
					p.energy -= p.deck[order].cost
					if p.deck[p.order[4]].id != 0 {
						p.order[data[1]] = p.order[4]
						for true {
							for i := 5; i < 8; i++ {
								p.order[i-1] = p.order[i]
							}
							p.order[7] = order
							order = p.order[4]
							if p.deck[order].id != 0 {
								break
							}
						}
					}
				}

			case 4: // deck set
				if p.game.status == 1 {
					break
				}
				p.SetDeck(data[1:])

			case 5:
				p.ping = time.Now().UnixNano() - p.lastTime[0]
				//log.Println(p.ping / int64(time.Millisecond))
				p.lastTime = p.lastTime[1:]
			}
			p.game.mutex.Unlock()
		}
	}
}

func (p *Player) Chatting(data []byte) {
	switch string(data) {
	case "/start":
		p.game.Start()
	default:
		data = append([]byte(p.name+": "), data...)
		data = append([]byte{2}, data...)
		p.game.Broadcast(data)
		/*for i := 0; i < p.game.PlayerCount; i++ {
			other := p.game.players[i]
			if p != other {
				other.Send(data)
			}
		}*/
	}
}

func (p *Player) CardUsingMethod(data []byte) {

}

func (p *Player) SetDeck(b []byte) {
	var cardidList [8]int
	for i := 0; len(b) > 0; i++ {
		cardidList[i] = int(binary.BigEndian.Uint32(b[0:4]))
		b = b[4:]
		if cardidList[i] == 0 {
			i--
		}
	}
	for i := 0; i < 8; i++ {
		p.deck[i] = CardList[cardidList[i]]
	}
	var logg string = "Set Deck : "
	for _, card := range p.deck {
		logg += strconv.FormatUint(uint64(card.id), 10) + ", "
	}
	log.Println(logg)
}

func (p Player) Data() []byte {
	var data []byte
	var idData []byte = make([]byte, 2)
	binary.BigEndian.PutUint16(idData, p.id)
	data = append(data, idData...)
	data = append(data, p.team)

	var nameData [20]byte
	for i, b := range []byte(p.name) {
		if i >= 20 {
			break
		}
		nameData[i] = b
	}
	//log.Println(nameData)
	//log.Println([]byte(p.name))

	data = append(data, nameData[:]...)

	return data
}

func (p Player) Send(data []byte) {
	size := make([]byte, 2)
	binary.BigEndian.PutUint16(size, uint16(len(data)))
	_, err := p.conn.Write(append(size, data...))
	if err != nil {
		log.Println(err)
		return
	}
}

func (p *Player) GetEnergy(e uint16) uint16 {
	p.energy += e
	if p.energy > p.maxEnergy {
		p.energy = p.maxEnergy
	}
	return p.energy
}
