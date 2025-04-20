package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// Setting holds the schema definition for key-value setting entity.
type Setting struct {
	ent.Schema
}

func (Setting) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			Unique(),
		field.Text("value").
			Optional(),
	}
}

func (Setting) Edges() []ent.Edge {
	return nil
}

func (Setting) Mixin() []ent.Mixin {
	return []ent.Mixin{
		CommonMixin{},
	}
}
