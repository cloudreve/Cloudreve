package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// DirectLink holds the schema definition for the DirectLink entity.
type DirectLink struct {
	ent.Schema
}

// Fields of the DirectLink.
func (DirectLink) Fields() []ent.Field {
	return []ent.Field{
		field.String("name"),
		field.Int("downloads"),
		field.Int("file_id"),
		field.Int("speed"),
	}
}

// Edges of the DirectLink.
func (DirectLink) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("file", File.Type).
			Ref("direct_links").
			Field("file_id").
			Required().
			Unique(),
	}
}

func (DirectLink) Mixin() []ent.Mixin {
	return []ent.Mixin{
		CommonMixin{},
	}
}
