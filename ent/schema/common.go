package schema

import (
	"context"
	"fmt"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
	gen "github.com/cloudreve/Cloudreve/v4/ent"
	"github.com/cloudreve/Cloudreve/v4/ent/hook"
	"github.com/cloudreve/Cloudreve/v4/ent/intercept"
)

// CommonMixin implements the soft delete pattern for schemas and common audit features.
type CommonMixin struct {
	mixin.Schema
}

// Fields of the CommonMixin.
func (CommonMixin) Fields() []ent.Field {
	return commonFields()
}

type softDeleteKey struct{}

// SkipSoftDelete returns a new context that skips the soft-delete interceptor/mutators.
func SkipSoftDelete(parent context.Context) context.Context {
	return context.WithValue(parent, softDeleteKey{}, true)
}

// Interceptors of the CommonMixin.
func (d CommonMixin) Interceptors() []ent.Interceptor {
	return softDeleteInterceptors(d)
}

// Hooks of the CommonMixin.
func (d CommonMixin) Hooks() []ent.Hook {
	return commonHooks(d)
}

// P adds a storage-level predicate to the queries and mutations.
func (d CommonMixin) P(w interface{ WhereP(...func(*sql.Selector)) }) {
	p(d, w)
}

// Indexes of the CommonMixin.
func (CommonMixin) Indexes() []ent.Index {
	return []ent.Index{}
}

func softDeleteInterceptors(d interface {
	P(w interface {
		WhereP(...func(*sql.Selector))
	})
}) []ent.Interceptor {
	return []ent.Interceptor{
		intercept.TraverseFunc(func(ctx context.Context, q intercept.Query) error {
			// Skip soft-delete, means include soft-deleted entities.
			if skip, _ := ctx.Value(softDeleteKey{}).(bool); skip {
				return nil
			}
			d.P(q)
			return nil
		}),
	}
}

func p(d interface{ Fields() []ent.Field }, w interface{ WhereP(...func(*sql.Selector)) }) {
	w.WhereP(
		sql.FieldIsNull(d.Fields()[2].Descriptor().Name),
	)
}

func commonFields() []ent.Field {
	return []ent.Field{
		field.Time("created_at").
			Immutable().
			Default(time.Now).
			SchemaType(map[string]string{
				dialect.MySQL: "datetime",
			}),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now).
			SchemaType(map[string]string{
				dialect.MySQL: "datetime",
			}),
		field.Time("deleted_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{
				dialect.MySQL: "datetime",
			}),
	}
}

func commonHooks(d interface {
	P(w interface {
		WhereP(...func(*sql.Selector))
	})
}) []ent.Hook {
	return []ent.Hook{
		hook.On(
			func(next ent.Mutator) ent.Mutator {
				return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
					// Skip soft-delete, means delete the entity permanently.
					if skip, _ := ctx.Value(softDeleteKey{}).(bool); skip {
						return next.Mutate(ctx, m)
					}
					mx, ok := m.(interface {
						SetOp(ent.Op)
						Client() *gen.Client
						SetDeletedAt(time.Time)
						WhereP(...func(*sql.Selector))
					})
					if !ok {
						return nil, fmt.Errorf("unexpected mutation type in soft-delete %T", m)
					}
					d.P(mx)
					mx.SetOp(ent.OpUpdate)
					mx.SetDeletedAt(time.Now())
					return mx.Client().Mutate(ctx, m)
				})
			},
			ent.OpDeleteOne|ent.OpDelete,
		),
	}
}
