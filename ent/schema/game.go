package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Game holds the schema definition for the Game entity.
type Game struct {
	ent.Schema
}

// 그런데 이건 단순히 게임 로그 기록용으로... 하는거니까
// 동적인 부분들은 싸그리 다 무시해야 할테니까
// 중요한 기록들 빼고 싹다 빼자
// Fields of the Game.
func (Game) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").
			Positive(),
		field.JSON("events", []string{}).
			Optional(),
		field.Time("created_at").
			Default(time.Now),
	}
}

// Edges of the Game.
func (Game) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("players", Player.Type),
	}
}

func (Game) Index() []ent.Index {
	return []ent.Index{
		index.Fields("id").Unique(),
	}
}
