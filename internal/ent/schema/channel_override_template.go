package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/looplj/axonhub/internal/ent/schema/schematype"
	"github.com/looplj/axonhub/internal/objects"
	"github.com/looplj/axonhub/internal/scopes"
)

// ChannelOverrideTemplate holds the schema definition for the ChannelOverrideTemplate entity.
type ChannelOverrideTemplate struct {
	ent.Schema
}

func (ChannelOverrideTemplate) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimeMixin{},
		schematype.SoftDeleteMixin{},
	}
}

func (ChannelOverrideTemplate) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("name", "deleted_at").
			StorageKey("channel_override_templates_by_name").
			Unique(),
	}
}

func (ChannelOverrideTemplate) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			NotEmpty().
			Comment("Template name"),
		field.String("description").
			Optional().
			Comment("Template description"),
		field.String("override_parameters").
			DefaultFunc(func() string { return "{}" }).
			Deprecated("Use body_override_operations instead").
			Annotations(
				entgql.Skip(entgql.SkipMutationCreateInput, entgql.SkipMutationUpdateInput),
			).
			Comment("Override request body parameters as JSON string"),

		field.JSON("override_headers", []objects.HeaderEntry{}).
			Default([]objects.HeaderEntry{}).
			Deprecated("Use header_override_operations instead").
			Annotations(
				entgql.Skip(entgql.SkipMutationCreateInput, entgql.SkipMutationUpdateInput),
			).
			Comment("Override request headers"),

		field.JSON("header_override_operations", []objects.OverrideOperation{}).
			Default([]objects.OverrideOperation{}).
			Optional().
			Annotations(
				entgql.Directives(forceResolver()),
			).
			Comment("Override request headers"),

		field.JSON("body_override_operations", []objects.OverrideOperation{}).
			Default([]objects.OverrideOperation{}).
			Optional().
			Annotations(
				entgql.Directives(forceResolver()),
			).
			Comment("Override request body parameters"),
	}
}

func (ChannelOverrideTemplate) Edges() []ent.Edge {
	return []ent.Edge{}
}

func (ChannelOverrideTemplate) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.QueryField(),
		entgql.RelayConnection(),
		entgql.Mutations(entgql.MutationCreate(), entgql.MutationUpdate()),
	}
}

// Policy defines the privacy policy for ChannelOverrideTemplate.
// Templates are private to users - users can only access their own templates.
// Owners can access all templates.
func (ChannelOverrideTemplate) Policy() ent.Policy {
	return scopes.Policy{
		Query: scopes.QueryPolicy{
			scopes.OwnerRule(),
		},
		Mutation: scopes.MutationPolicy{
			scopes.OwnerRule(),
		},
	}
}
