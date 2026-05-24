package scopes

import (
	"context"

	"github.com/looplj/axonhub/internal/ent"
	"github.com/looplj/axonhub/internal/ent/privacy"
)

// APIKeyScopeQueryRule checks API Key permissions for queries.
func APIKeyScopeQueryRule(requiredScope ScopeSlug) privacy.QueryRule {
	return apiKeyQueryRule{requiredScope: requiredScope}
}

// apiKeyQueryRule custom QueryRule implementation for checking API Key scopes.
type apiKeyQueryRule struct {
	requiredScope ScopeSlug
}

func (r apiKeyQueryRule) EvalQuery(ctx context.Context, q ent.Query) error {
	_, err := getAPIKeyFromContext(ctx)
	if err != nil {
		return err
	}

	return privacy.Allow
}

// APIKeyScopeMutationRule checks API Key write permissions.
func APIKeyScopeMutationRule(requiredScope ScopeSlug) privacy.MutationRule {
	return privacy.MutationRuleFunc(func(ctx context.Context, m ent.Mutation) error {
		_, err := getAPIKeyFromContext(ctx)
		if err != nil {
			return err
		}

		return privacy.Allow
	})
}
