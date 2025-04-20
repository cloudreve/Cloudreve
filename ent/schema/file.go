package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/cloudreve/Cloudreve/v4/inventory/types"
)

// File holds the schema definition for the File entity.
type File struct {
	ent.Schema
}

// Fields of the File.
func (File) Fields() []ent.Field {
	return []ent.Field{
		field.Int("type"),
		field.String("name"),
		field.Int("owner_id"),
		field.Int64("size").
			Default(0),
		field.Int("primary_entity").
			Optional(),
		field.Int("file_children").
			Optional(),
		field.Bool("is_symbolic").
			Default(false),
		field.JSON("props", &types.FileProps{}).Optional(),
		field.Int("storage_policy_files").
			Optional(),
	}
}

// Edges of the File.
func (File) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("owner", User.Type).
			Ref("files").
			Field("owner_id").
			Unique().
			Required(),
		edge.From("storage_policies", StoragePolicy.Type).
			Ref("files").
			Field("storage_policy_files").
			Unique(),
		edge.To("children", File.Type).
			From("parent").
			Field("file_children").
			Unique(),
		edge.To("metadata", Metadata.Type),
		edge.To("entities", Entity.Type),
		edge.To("shares", Share.Type),
		edge.To("direct_links", DirectLink.Type),
	}
}

// Indexes of the File.
func (File) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("file_children", "name").
			Unique(),
		index.Fields("file_children", "type", "updated_at"),
		index.Fields("file_children", "type", "size"),
	}
}

func (File) Mixin() []ent.Mixin {
	return []ent.Mixin{
		CommonMixin{},
	}
}
