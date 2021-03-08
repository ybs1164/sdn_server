package main

import (
	"context"
	"log"
	"net"
	"runtime"

	"app/ent"
	g "app/object/games"

	_ "github.com/mattn/go-sqlite3"
)

var games []*g.Game

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	// todo : init DB
	client, err := ent.Open("sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	if err != nil {
		log.Fatalf("failed opening connection to sqlite: %v", err)
	}
	defer client.Close()

	if err := client.Schema.Create(context.Background()); err != nil {
		log.Fatalf("failed creating schema resources: %v", err)
	}

	l, err := net.Listen("tcp", ":30004")
	if err != nil {
		log.Println(err)
		return
	} else {
		log.Println("Server Open")
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Println(err)
		} else {
			if len(games) == 0 {
				games = append(games, g.NewGame())
			}
			isJoin := false
			for _, g := range games {
				if g.PlayerCount < 3 {
					g.Join(conn)
					isJoin = true
				}
			}
			if !isJoin {
				games = append(games, g.NewGame())
				games[len(games)-1].Join(conn)
			}
		}
	}
}
