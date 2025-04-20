package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/go-webauthn/webauthn/webauthn"
)

// Passkey holds the schema definition for the Passkey entity.
type Passkey struct {
	ent.Schema
}

// Fields of the Passkey.
func (Passkey) Fields() []ent.Field {
	return []ent.Field{
		field.Int("user_id"),
		field.String("credential_id"),
		field.String("name"),
		field.JSON("credential", &webauthn.Credential{}).
			Sensitive(),
		field.Time("used_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{
				dialect.MySQL: "datetime",
			}),
	}
}

// Edges of the Passkey.
func (Passkey) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Field("user_id").
			Ref("passkey").
			Unique().
			Required(),
	}
}

func (Passkey) Mixin() []ent.Mixin {
	return []ent.Mixin{
		CommonMixin{},
	}
}

func (Passkey) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "credential_id").Unique(),
	}
}
