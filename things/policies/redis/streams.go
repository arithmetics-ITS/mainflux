// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package redis

import (
	"context"

	"github.com/go-redis/redis/v8"
	mfredis "github.com/mainflux/mainflux/internal/clients/redis"
	"github.com/mainflux/mainflux/things/policies"
)

const (
	streamID  = "mainflux.things"
	streamLen = 1000
)

var _ policies.Service = (*eventStore)(nil)

type eventStore struct {
	mfredis.Publisher
	svc    policies.Service
	client *redis.Client
}

// NewEventStoreMiddleware returns wrapper around policy service that sends
// events to event store.
func NewEventStoreMiddleware(ctx context.Context, svc policies.Service, client *redis.Client) policies.Service {
	es := &eventStore{
		svc:       svc,
		client:    client,
		Publisher: mfredis.NewEventStore(client, streamID, streamLen),
	}

	go es.StartPublishingRoutine(ctx)

	return es
}

func (es *eventStore) Authorize(ctx context.Context, ar policies.AccessRequest) (policies.Policy, error) {
	policy, err := es.svc.Authorize(ctx, ar)
	if err != nil {
		return policy, err
	}

	event := authorizeEvent{
		ar,
	}
	if err := es.Publish(ctx, event); err != nil {
		return policy, err
	}

	return policy, nil
}

func (es eventStore) AddPolicy(ctx context.Context, token string, external bool, policy policies.Policy) (policies.Policy, error) {
	policy, err := es.svc.AddPolicy(ctx, token, external, policy)
	if err != nil {
		return policy, err
	}

	event := policyEvent{
		policy, policyAdd,
	}
	if err := es.Publish(ctx, event); err != nil {
		return policy, err
	}

	return policy, nil
}

func (es *eventStore) UpdatePolicy(ctx context.Context, token string, policy policies.Policy) (policies.Policy, error) {
	policy, err := es.svc.UpdatePolicy(ctx, token, policy)
	if err != nil {
		return policy, err
	}

	event := policyEvent{
		policy, policyUpdate,
	}
	if err := es.Publish(ctx, event); err != nil {
		return policy, err
	}

	return policy, nil
}

func (es *eventStore) ListPolicies(ctx context.Context, token string, page policies.Page) (policies.PolicyPage, error) {
	pp, err := es.svc.ListPolicies(ctx, token, page)
	if err != nil {
		return pp, err
	}

	event := listPoliciesEvent{
		page,
	}
	if err := es.Publish(ctx, event); err != nil {
		return pp, err
	}

	return pp, nil
}

func (es *eventStore) DeletePolicy(ctx context.Context, token string, policy policies.Policy) error {
	if err := es.svc.DeletePolicy(ctx, token, policy); err != nil {
		return err
	}

	event := policyEvent{
		policy, policyDelete,
	}

	return es.Publish(ctx, event)
}
