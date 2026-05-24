package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/looplj/axonhub/internal/ent/schema/schematype"
	"github.com/looplj/axonhub/internal/scopes"
)

// User holds the schema definition for the User entity.
type User struct {
	ent.Schema
}

func (User) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimeMixin{},
		schematype.SoftDeleteMixin{},
	}
}

// Indexes of the User.
// Unique index on (email, deleted_at) to allow same email after soft delete.
func (User) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("email", "deleted_at").Unique(),
	}
}

// Fields of the User.
func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("email"),
		field.Enum("status").Values("activated", "deactivated").Default("activated"),
		field.String("prefer_language").Default("en").Comment("用户偏好语言"),
		field.String("password").Sensitive(),
		field.String("first_name").Default(""),
		field.String("last_name").Default(""),
		field.String("avatar").Optional().Comment("用户头像URL").SchemaType(
			map[string]string{
				// The avatar is stored as base64 image, it is too long to store in varchar, so we use mediumtext to store it.
				dialect.MySQL: "mediumtext",
			},
		),
		field.Bool("is_owner").Default(false),
	}
}

// Edges of the User.
func (User) Edges() []ent.Edge {
	return []ent.Edge{}
}

func (User) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.RelayConnection(),
	}
}

func (User) Policy() ent.Policy {
	return scopes.Policy{
		Query: scopes.QueryPolicy{
			scopes.OwnerRule(),
			scopes.UserOwnedQueryRule(),
		},
		Mutation: scopes.MutationPolicy{
			scopes.OwnerRule(),
			scopes.UserOwnedMutationRule(),
		},
	}
}
