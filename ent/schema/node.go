package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/cloudreve/Cloudreve/v4/inventory/types"
	"github.com/cloudreve/Cloudreve/v4/pkg/boolset"
)

// Node holds the schema definition for the Node entity.
type Node struct {
	ent.Schema
}

// Fields of the Node.
func (Node) Fields() []ent.Field {
	return []ent.Field{
		field.Enum("status").
			Values("active", "suspended"),
		field.String("name"),
		field.Enum("type").
			Values("master", "slave"),
		field.String("server").
			Optional(),
		field.String("slave_key").Optional(),
		field.Bytes("capabilities").GoType(&boolset.BooleanSet{}),
		field.JSON("settings", &types.NodeSetting{}).
			Default(&types.NodeSetting{}).
			Optional(),
		field.Int("weight").Default(0),
	}
}

// Edges of the Node.
func (Node) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("storage_policy", StoragePolicy.Type),
	}
}

func (Node) Mixin() []ent.Mixin {
	return []ent.Mixin{
		CommonMixin{},
	}
}
