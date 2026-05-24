package biz

import (
	"context"
	"fmt"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/looplj/axonhub/internal/ent"
	"github.com/looplj/axonhub/internal/ent/user"
	"github.com/looplj/axonhub/internal/log"
	"github.com/looplj/axonhub/internal/objects"
	"github.com/looplj/axonhub/internal/pkg/xcache"
)

type UserServiceParams struct {
	fx.In

	CacheConfig xcache.Config
	Ent         *ent.Client
}

type UserService struct {
	*AbstractService

	UserCache xcache.Cache[ent.User]
}

type CreateUserInput struct {
	Email     string
	Password  string
	FirstName *string
	LastName  *string
}

type UpdateUserInput struct {
	Email          *string
	FirstName      *string
	LastName       *string
	PreferLanguage *string
	Avatar         *string
	ClearAvatar    bool
	Password       *string
}

func NewUserService(params UserServiceParams) *UserService {
	return &UserService{
		AbstractService: &AbstractService{
			db: params.Ent,
		},
		UserCache: xcache.NewFromConfig[ent.User](params.CacheConfig),
	}
}

// CreateUser creates a new user with hashed password.
func (s *UserService) CreateUser(ctx context.Context, input CreateUserInput) (*ent.User, error) {
	client := s.entFromContext(ctx)

	hashedPassword, err := HashPassword(input.Password)
	if err != nil {
		return nil, err
	}

	mut := client.User.Create().
		SetNillableFirstName(input.FirstName).
		SetNillableLastName(input.LastName).
		SetEmail(input.Email).
		SetPassword(hashedPassword)

	user, err := mut.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// UpdateUser updates an existing user.
func (s *UserService) UpdateUser(ctx context.Context, id int, input UpdateUserInput) (*ent.User, error) {
	client := s.entFromContext(ctx)

	mut := client.User.UpdateOneID(id).
		SetNillableEmail(input.Email).
		SetNillableFirstName(input.FirstName).
		SetNillableLastName(input.LastName).
		SetNillablePreferLanguage(input.PreferLanguage)

	if input.ClearAvatar {
		mut.ClearAvatar()
	} else {
		mut.SetNillableAvatar(input.Avatar)
	}

	if input.Password != nil {
		hashedPassword, err := HashPassword(*input.Password)
		if err != nil {
			return nil, err
		}

		mut.SetPassword(hashedPassword)
	}

	user, err := mut.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	// Invalidate cache
	s.invalidateUserCache(ctx, id)

	return user, nil
}

// UpdateUserStatus updates the status of a user.
func (s *UserService) UpdateUserStatus(ctx context.Context, id int, status user.Status) (*ent.User, error) {
	client := s.entFromContext(ctx)

	user, err := client.User.UpdateOneID(id).
		SetStatus(status).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to update user status: %w", err)
	}

	// Invalidate cache
	s.invalidateUserCache(ctx, id)

	return user, nil
}

// GetUserByID gets a user by ID with caching.
func (s *UserService) GetUserByID(ctx context.Context, id int) (*ent.User, error) {
	// Try cache first
	cacheKey := buildUserCacheKey(id)
	if user, err := s.UserCache.Get(ctx, cacheKey); err == nil {
		return &user, nil
	}

	// Query database
	client := s.entFromContext(ctx)
	if client == nil {
		return nil, fmt.Errorf("ent client not found in context")
	}

	user, err := client.User.Query().
		Where(user.IDEQ(id)).
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Cache the user
	// TODO: handle role scope changed.
	err = s.UserCache.Set(ctx, cacheKey, *user)
	if err != nil {
		log.Warn(ctx, "failed to cache user", zap.Error(err))
	}

	return user, nil
}

func buildUserCacheKey(id int) string {
	return fmt.Sprintf("user:%d", id)
}

// invalidateUserCache removes a user from cache.
func (s *UserService) invalidateUserCache(ctx context.Context, id int) {
	cacheKey := buildUserCacheKey(id)
	_ = s.UserCache.Delete(ctx, cacheKey)
}

// clearUserCache clears all user cache.
func (s *UserService) clearUserCache(ctx context.Context) {
	_ = s.UserCache.Clear(ctx)
}

// ConvertUserToUserInfo converts ent.User to objects.UserInfo.
func ConvertUserToUserInfo(ctx context.Context, u *ent.User) *objects.UserInfo {
	return &objects.UserInfo{
		ID:             objects.GUID{Type: ent.TypeUser, ID: u.ID},
		Email:          u.Email,
		FirstName:      u.FirstName,
		LastName:       u.LastName,
		IsOwner:        u.IsOwner,
		PreferLanguage: u.PreferLanguage,
		Avatar:         &u.Avatar,
		HasPassword:    u.Password != "",
	}
}
