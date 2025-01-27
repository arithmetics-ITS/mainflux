// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package redis

import (
	"context"

	"github.com/go-redis/redis/v8"
	mfredis "github.com/mainflux/mainflux/internal/clients/redis"
	mfclients "github.com/mainflux/mainflux/pkg/clients"
	"github.com/mainflux/mainflux/users/clients"
	"github.com/mainflux/mainflux/users/jwt"
)

const (
	streamID  = "mainflux.users"
	streamLen = 1000
)

var _ clients.Service = (*eventStore)(nil)

type eventStore struct {
	mfredis.Publisher
	svc    clients.Service
	client *redis.Client
}

// NewEventStoreMiddleware returns wrapper around users service that sends
// events to event store.
func NewEventStoreMiddleware(ctx context.Context, svc clients.Service, client *redis.Client) clients.Service {
	es := eventStore{
		svc:       svc,
		client:    client,
		Publisher: mfredis.NewEventStore(client, streamID, streamLen),
	}

	go es.StartPublishingRoutine(ctx)

	return es
}

func (es eventStore) RegisterClient(ctx context.Context, token string, user mfclients.Client) (mfclients.Client, error) {
	user, err := es.svc.RegisterClient(ctx, token, user)
	if err != nil {
		return user, err
	}

	event := createClientEvent{
		user,
	}

	if err := es.Publish(ctx, event); err != nil {
		return user, err
	}

	return user, nil
}

func (es eventStore) UpdateClient(ctx context.Context, token string, user mfclients.Client) (mfclients.Client, error) {
	user, err := es.svc.UpdateClient(ctx, token, user)
	if err != nil {
		return user, err
	}

	return es.update(ctx, "", user)
}

func (es eventStore) UpdateClientOwner(ctx context.Context, token string, user mfclients.Client) (mfclients.Client, error) {
	user, err := es.svc.UpdateClientOwner(ctx, token, user)
	if err != nil {
		return user, err
	}

	return es.update(ctx, "owner", user)
}

func (es eventStore) UpdateClientTags(ctx context.Context, token string, user mfclients.Client) (mfclients.Client, error) {
	user, err := es.svc.UpdateClientTags(ctx, token, user)
	if err != nil {
		return user, err
	}

	return es.update(ctx, "tags", user)
}

func (es eventStore) UpdateClientSecret(ctx context.Context, token, oldSecret, newSecret string) (mfclients.Client, error) {
	user, err := es.svc.UpdateClientSecret(ctx, token, oldSecret, newSecret)
	if err != nil {
		return user, err
	}

	return es.update(ctx, "secret", user)
}

func (es eventStore) UpdateClientIdentity(ctx context.Context, token, id, identity string) (mfclients.Client, error) {
	user, err := es.svc.UpdateClientIdentity(ctx, token, id, identity)
	if err != nil {
		return user, err
	}

	return es.update(ctx, "identity", user)
}

func (es eventStore) update(ctx context.Context, operation string, user mfclients.Client) (mfclients.Client, error) {
	event := updateClientEvent{
		user, operation,
	}

	if err := es.Publish(ctx, event); err != nil {
		return user, err
	}

	return user, nil
}

func (es eventStore) ViewClient(ctx context.Context, token, id string) (mfclients.Client, error) {
	user, err := es.svc.ViewClient(ctx, token, id)
	if err != nil {
		return user, err
	}

	event := viewClientEvent{
		user,
	}

	if err := es.Publish(ctx, event); err != nil {
		return user, err
	}

	return user, nil
}

func (es eventStore) ViewProfile(ctx context.Context, token string) (mfclients.Client, error) {
	user, err := es.svc.ViewProfile(ctx, token)
	if err != nil {
		return user, err
	}

	event := viewProfileEvent{
		user,
	}

	if err := es.Publish(ctx, event); err != nil {
		return user, err
	}

	return user, nil
}

func (es eventStore) ListClients(ctx context.Context, token string, pm mfclients.Page) (mfclients.ClientsPage, error) {
	cp, err := es.svc.ListClients(ctx, token, pm)
	if err != nil {
		return cp, err
	}
	event := listClientEvent{
		pm,
	}

	if err := es.Publish(ctx, event); err != nil {
		return cp, err
	}

	return cp, nil
}

func (es eventStore) ListMembers(ctx context.Context, token, groupID string, pm mfclients.Page) (mfclients.MembersPage, error) {
	mp, err := es.svc.ListMembers(ctx, token, groupID, pm)
	if err != nil {
		return mp, err
	}
	event := listClientByGroupEvent{
		pm, groupID,
	}

	if err := es.Publish(ctx, event); err != nil {
		return mp, err
	}

	return mp, nil
}

func (es eventStore) EnableClient(ctx context.Context, token, id string) (mfclients.Client, error) {
	user, err := es.svc.EnableClient(ctx, token, id)
	if err != nil {
		return user, err
	}

	return es.delete(ctx, user)
}

func (es eventStore) DisableClient(ctx context.Context, token, id string) (mfclients.Client, error) {
	user, err := es.svc.DisableClient(ctx, token, id)
	if err != nil {
		return user, err
	}

	return es.delete(ctx, user)
}

func (es eventStore) delete(ctx context.Context, user mfclients.Client) (mfclients.Client, error) {
	event := removeClientEvent{
		id:        user.ID,
		updatedAt: user.UpdatedAt,
		updatedBy: user.UpdatedBy,
		status:    user.Status.String(),
	}

	if err := es.Publish(ctx, event); err != nil {
		return user, err
	}

	return user, nil
}

func (es eventStore) Identify(ctx context.Context, token string) (string, error) {
	userID, err := es.svc.Identify(ctx, token)
	if err != nil {
		return userID, err
	}
	event := identifyClientEvent{
		userID: userID,
	}

	if err := es.Publish(ctx, event); err != nil {
		return userID, err
	}

	return userID, nil
}

func (es eventStore) GenerateResetToken(ctx context.Context, email, host string) error {
	if err := es.svc.GenerateResetToken(ctx, email, host); err != nil {
		return err
	}
	event := generateResetTokenEvent{
		email: email,
		host:  host,
	}

	return es.Publish(ctx, event)
}

func (es eventStore) IssueToken(ctx context.Context, identity, secret string) (jwt.Token, error) {
	token, err := es.svc.IssueToken(ctx, identity, secret)
	if err != nil {
		return token, err
	}
	event := issueTokenEvent{
		identity: identity,
	}

	if err := es.Publish(ctx, event); err != nil {
		return token, err
	}

	return token, nil
}

func (es eventStore) RefreshToken(ctx context.Context, refreshToken string) (jwt.Token, error) {
	token, err := es.svc.RefreshToken(ctx, refreshToken)
	if err != nil {
		return token, err
	}
	event := refreshTokenEvent{}

	if err := es.Publish(ctx, event); err != nil {
		return token, err
	}

	return token, nil
}

func (es eventStore) ResetSecret(ctx context.Context, resetToken, secret string) error {
	if err := es.svc.ResetSecret(ctx, resetToken, secret); err != nil {
		return err
	}
	event := resetSecretEvent{}

	return es.Publish(ctx, event)
}

func (es eventStore) SendPasswordReset(ctx context.Context, host, email, user, token string) error {
	if err := es.svc.SendPasswordReset(ctx, host, email, user, token); err != nil {
		return err
	}
	event := sendPasswordResetEvent{
		host:  host,
		email: email,
		user:  user,
	}

	return es.Publish(ctx, event)
}
