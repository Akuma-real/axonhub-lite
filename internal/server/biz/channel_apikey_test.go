package biz

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/looplj/axonhub/internal/authz"
	"github.com/looplj/axonhub/internal/ent"
	"github.com/looplj/axonhub/internal/ent/channel"
	"github.com/looplj/axonhub/internal/objects"
)

// ==================== DeleteDisabledAPIKeys Tests ====================.
func TestChannelService_DeleteDisabledAPIKeys_SingleKey(t *testing.T) {
	svc, client := setupTestChannelService(t)
	defer client.Close()

	ctx := context.Background()
	ctx = ent.NewContext(ctx, client)
	ctx = authz.WithTestBypass(ctx)

	// Create a test channel with multiple keys, one disabled
	ch, err := client.Channel.Create().
		SetType(channel.TypeOpenai).
		SetName("Test Channel").
		SetBaseURL("https://api.openai.com/v1").
		SetCredentials(objects.ChannelCredentials{
			APIKeys: []string{"key1", "key2", "key3"},
		}).
		SetSupportedModels([]string{"gpt-4"}).
		SetDefaultTestModel("gpt-4").
		SetDisabledAPIKeys([]objects.DisabledAPIKey{
			{Key: "key2", ErrorCode: 401, Reason: "invalid key"},
		}).
		Save(ctx)
	require.NoError(t, err)

	// Delete the disabled key
	result, err := svc.DeleteDisabledAPIKeys(ctx, ch.ID, []string{"key2"})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.True(t, result.Success)
	require.Empty(t, result.Message)

	// Verify key is removed from both disabled list and credentials
	updatedCh, err := client.Channel.Get(ctx, ch.ID)
	require.NoError(t, err)
	require.Len(t, updatedCh.DisabledAPIKeys, 0)
	require.Len(t, updatedCh.Credentials.APIKeys, 2)
	require.NotContains(t, updatedCh.Credentials.APIKeys, "key2")
	require.Contains(t, updatedCh.Credentials.APIKeys, "key1")
	require.Contains(t, updatedCh.Credentials.APIKeys, "key3")
}

func TestChannelService_DeleteDisabledAPIKeys_MultipleKeys(t *testing.T) {
	svc, client := setupTestChannelService(t)
	defer client.Close()

	ctx := context.Background()
	ctx = ent.NewContext(ctx, client)
	ctx = authz.WithTestBypass(ctx)

	// Create a test channel with multiple keys, some disabled
	ch, err := client.Channel.Create().
		SetType(channel.TypeOpenai).
		SetName("Test Channel").
		SetBaseURL("https://api.openai.com/v1").
		SetCredentials(objects.ChannelCredentials{
			APIKeys: []string{"key1", "key2", "key3", "key4"},
		}).
		SetSupportedModels([]string{"gpt-4"}).
		SetDefaultTestModel("gpt-4").
		SetDisabledAPIKeys([]objects.DisabledAPIKey{
			{Key: "key2", ErrorCode: 401, Reason: "invalid key"},
			{Key: "key4", ErrorCode: 429, Reason: "rate limited"},
		}).
		Save(ctx)
	require.NoError(t, err)

	// Delete multiple disabled keys
	result, err := svc.DeleteDisabledAPIKeys(ctx, ch.ID, []string{"key2", "key4"})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.True(t, result.Success)

	// Verify keys are removed
	updatedCh, err := client.Channel.Get(ctx, ch.ID)
	require.NoError(t, err)
	require.Len(t, updatedCh.DisabledAPIKeys, 0)
	require.Len(t, updatedCh.Credentials.APIKeys, 2)
	require.Contains(t, updatedCh.Credentials.APIKeys, "key1")
	require.Contains(t, updatedCh.Credentials.APIKeys, "key3")
	require.NotContains(t, updatedCh.Credentials.APIKeys, "key2")
	require.NotContains(t, updatedCh.Credentials.APIKeys, "key4")
}

func TestChannelService_DeleteDisabledAPIKeys_OAuthChannel(t *testing.T) {
	svc, client := setupTestChannelService(t)
	defer client.Close()

	ctx := context.Background()
	ctx = ent.NewContext(ctx, client)
	ctx = authz.WithTestBypass(ctx)

	// Create an OAuth channel
	ch, err := client.Channel.Create().
		SetType(channel.TypeOpenai).
		SetName("OAuth Channel").
		SetBaseURL("https://api.openai.com/v1").
		SetCredentials(objects.ChannelCredentials{
			OAuth: &objects.OAuthCredentials{
				AccessToken: "test-token",
			},
		}).
		SetSupportedModels([]string{"gpt-4"}).
		SetDefaultTestModel("gpt-4").
		Save(ctx)
	require.NoError(t, err)

	// Try to delete keys from OAuth channel
	result, err := svc.DeleteDisabledAPIKeys(ctx, ch.ID, []string{"some-key"})
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "cannot delete API keys for OAuth channels")
}

func TestChannelService_DeleteDisabledAPIKeys_PreserveAtLeastOneKey(t *testing.T) {
	svc, client := setupTestChannelService(t)
	defer client.Close()

	ctx := context.Background()
	ctx = ent.NewContext(ctx, client)
	ctx = authz.WithTestBypass(ctx)

	// Create a test channel with only one key (disabled)
	ch, err := client.Channel.Create().
		SetType(channel.TypeOpenai).
		SetName("Test Channel").
		SetBaseURL("https://api.openai.com/v1").
		SetCredentials(objects.ChannelCredentials{
			APIKeys: []string{"only-key"},
		}).
		SetSupportedModels([]string{"gpt-4"}).
		SetDefaultTestModel("gpt-4").
		SetDisabledAPIKeys([]objects.DisabledAPIKey{
			{Key: "only-key", ErrorCode: 401, Reason: "invalid key"},
		}).
		Save(ctx)
	require.NoError(t, err)

	// Try to delete the only key - should preserve it
	result, err := svc.DeleteDisabledAPIKeys(ctx, ch.ID, []string{"only-key"})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.True(t, result.Success)
	require.Equal(t, "ONE_KEY_PRESERVED", result.Message)

	// Verify the key is preserved
	updatedCh, err := client.Channel.Get(ctx, ch.ID)
	require.NoError(t, err)
	require.Len(t, updatedCh.Credentials.APIKeys, 1)
	require.Contains(t, updatedCh.Credentials.APIKeys, "only-key")
	// The key should be removed from disabled list since it's the only one
	require.Len(t, updatedCh.DisabledAPIKeys, 0)
}

func TestChannelService_DeleteDisabledAPIKeys_WithLegacyAPIKey(t *testing.T) {
	svc, client := setupTestChannelService(t)
	defer client.Close()

	ctx := context.Background()
	ctx = ent.NewContext(ctx, client)
	ctx = authz.WithTestBypass(ctx)

	// Create a test channel with legacy APIKey field
	ch, err := client.Channel.Create().
		SetType(channel.TypeOpenai).
		SetName("Test Channel").
		SetBaseURL("https://api.openai.com/v1").
		SetCredentials(objects.ChannelCredentials{
			APIKey:  "legacy-key",
			APIKeys: []string{"new-key-1", "new-key-2"},
		}).
		SetSupportedModels([]string{"gpt-4"}).
		SetDefaultTestModel("gpt-4").
		SetDisabledAPIKeys([]objects.DisabledAPIKey{
			{Key: "legacy-key", ErrorCode: 401, Reason: "invalid key"},
		}).
		Save(ctx)
	require.NoError(t, err)

	// Delete the legacy key
	result, err := svc.DeleteDisabledAPIKeys(ctx, ch.ID, []string{"legacy-key"})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.True(t, result.Success)

	// Verify legacy key is removed
	updatedCh, err := client.Channel.Get(ctx, ch.ID)
	require.NoError(t, err)
	require.Empty(t, updatedCh.Credentials.APIKey)
	require.Len(t, updatedCh.Credentials.APIKeys, 2)
}

func TestChannelService_DeleteDisabledAPIKeys_PartialKeysNotInDisabledList(t *testing.T) {
	svc, client := setupTestChannelService(t)
	defer client.Close()

	ctx := context.Background()
	ctx = ent.NewContext(ctx, client)
	ctx = authz.WithTestBypass(ctx)

	// Create a test channel with some disabled keys
	ch, err := client.Channel.Create().
		SetType(channel.TypeOpenai).
		SetName("Test Channel").
		SetBaseURL("https://api.openai.com/v1").
		SetCredentials(objects.ChannelCredentials{
			APIKeys: []string{"key1", "key2", "key3"},
		}).
		SetSupportedModels([]string{"gpt-4"}).
		SetDefaultTestModel("gpt-4").
		SetDisabledAPIKeys([]objects.DisabledAPIKey{
			{Key: "key2", ErrorCode: 401, Reason: "invalid key"},
		}).
		Save(ctx)
	require.NoError(t, err)

	// Try to delete key2 (disabled) and key3 (not disabled)
	result, err := svc.DeleteDisabledAPIKeys(ctx, ch.ID, []string{"key2", "key3"})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.True(t, result.Success)

	// Verify both keys are removed from credentials
	updatedCh, err := client.Channel.Get(ctx, ch.ID)
	require.NoError(t, err)
	require.Len(t, updatedCh.DisabledAPIKeys, 0)
	require.Len(t, updatedCh.Credentials.APIKeys, 1)
	require.Contains(t, updatedCh.Credentials.APIKeys, "key1")
	require.NotContains(t, updatedCh.Credentials.APIKeys, "key2")
	require.NotContains(t, updatedCh.Credentials.APIKeys, "key3")
}

func TestChannelService_DeleteDisabledAPIKeys_NoDisabledKeys(t *testing.T) {
	svc, client := setupTestChannelService(t)
	defer client.Close()

	ctx := context.Background()
	ctx = ent.NewContext(ctx, client)
	ctx = authz.WithTestBypass(ctx)

	// Create a test channel with no disabled keys
	ch, err := client.Channel.Create().
		SetType(channel.TypeOpenai).
		SetName("Test Channel").
		SetBaseURL("https://api.openai.com/v1").
		SetCredentials(objects.ChannelCredentials{
			APIKeys: []string{"key1", "key2"},
		}).
		SetSupportedModels([]string{"gpt-4"}).
		SetDefaultTestModel("gpt-4").
		Save(ctx)
	require.NoError(t, err)

	// Try to delete a key that's not disabled
	result, err := svc.DeleteDisabledAPIKeys(ctx, ch.ID, []string{"key1"})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.True(t, result.Success)

	// Verify key is removed from credentials even if not disabled
	updatedCh, err := client.Channel.Get(ctx, ch.ID)
	require.NoError(t, err)
	require.Len(t, updatedCh.Credentials.APIKeys, 1)
	require.Contains(t, updatedCh.Credentials.APIKeys, "key2")
	require.NotContains(t, updatedCh.Credentials.APIKeys, "key1")
}
