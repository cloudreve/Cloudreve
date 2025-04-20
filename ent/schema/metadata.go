package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Metadata holds the schema definition for the Metadata entity.
type Metadata struct {
	ent.Schema
}

// Fields of the Metadata.
func (Metadata) Fields() []ent.Field {
	return []ent.Field{
		field.String("name"),
		field.Text("value"),
		field.Int("file_id"),
		field.Bool("is_public").
			Default(false),
	}
}

// Edges of the Metadata.
func (Metadata) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("file", File.Type).
			Ref("metadata").
			Field("file_id").
			Required().
			Unique(),
	}
}

func (Metadata) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("file_id", "name").
			Unique(),
	}
}

func (Metadata) Mixin() []ent.Mixin {
	return []ent.Mixin{
		CommonMixin{},
	}
}
