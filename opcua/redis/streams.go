// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/go-redis/redis/v8"
	mflog "github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/opcua"
)

const (
	keyType      = "opcua"
	keyNodeID    = "node_id"
	keyServerURI = "server_uri"

	group  = "mainflux.opcua"
	stream = "mainflux.things"

	thingPrefix     = "thing."
	thingCreate     = thingPrefix + "create"
	thingUpdate     = thingPrefix + "update"
	thingRemove     = thingPrefix + "remove"
	thingConnect    = thingPrefix + "connect"
	thingDisconnect = thingPrefix + "disconnect"

	channelPrefix = "channel."
	channelCreate = channelPrefix + "create"
	channelUpdate = channelPrefix + "update"
	channelRemove = channelPrefix + "remove"

	exists = "BUSYGROUP Consumer Group name already exists"
)

var (
	errMetadataType = errors.New("metadatada is not of type opcua")

	errMetadataFormat = errors.New("malformed metadata")

	errMetadataServerURI = errors.New("ServerURI not found in channel metadatada")

	errMetadataNodeID = errors.New("NodeID not found in thing metadatada")
)

var _ opcua.EventStore = (*eventStore)(nil)

type eventStore struct {
	svc      opcua.Service
	client   *redis.Client
	consumer string
	logger   mflog.Logger
}

// NewEventStore returns new event store instance.
func NewEventStore(svc opcua.Service, client *redis.Client, consumer string, log mflog.Logger) opcua.EventStore {
	return eventStore{
		svc:      svc,
		client:   client,
		consumer: consumer,
		logger:   log,
	}
}

func (es eventStore) Subscribe(ctx context.Context, subject string) error {
	err := es.client.XGroupCreateMkStream(ctx, stream, group, "$").Err()
	if err != nil && err.Error() != exists {
		return err
	}

	for {
		streams, err := es.client.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    group,
			Consumer: es.consumer,
			Streams:  []string{stream, ">"},
			Count:    100,
		}).Result()
		if err != nil || len(streams) == 0 {
			continue
		}

		for _, msg := range streams[0].Messages {
			event := msg.Values

			var err error
			switch event["operation"] {
			case thingCreate:
				cte, e := decodeCreateThing(event)
				if e != nil {
					err = e
					break
				}
				err = es.svc.CreateThing(ctx, cte.id, cte.opcuaNodeID)
			case thingUpdate:
				ute, e := decodeCreateThing(event)
				if e != nil {
					err = e
					break
				}
				err = es.svc.CreateThing(ctx, ute.id, ute.opcuaNodeID)
			case thingRemove:
				rte := decodeRemoveThing(event)
				err = es.svc.RemoveThing(ctx, rte.id)
			case channelCreate:
				cce, e := decodeCreateChannel(event)
				if e != nil {
					err = e
					break
				}
				err = es.svc.CreateChannel(ctx, cce.id, cce.opcuaServerURI)
			case channelUpdate:
				uce, e := decodeCreateChannel(event)
				if e != nil {
					err = e
					break
				}
				err = es.svc.CreateChannel(ctx, uce.id, uce.opcuaServerURI)
			case channelRemove:
				rce := decodeRemoveChannel(event)
				err = es.svc.RemoveChannel(ctx, rce.id)
			case thingConnect:
				rce := decodeConnectThing(event)
				err = es.svc.ConnectThing(ctx, rce.chanID, rce.thingID)
			case thingDisconnect:
				rce := decodeDisconnectThing(event)
				err = es.svc.DisconnectThing(ctx, rce.chanID, rce.thingID)
			}
			if err != nil && err != errMetadataType {
				es.logger.Warn(fmt.Sprintf("Failed to handle event sourcing: %s", err.Error()))
				break
			}
			es.client.XAck(ctx, stream, group, msg.ID)
		}
	}
}

func decodeCreateThing(event map[string]interface{}) (createThingEvent, error) {
	strmeta := read(event, "metadata", "{}")
	var metadata map[string]interface{}
	if err := json.Unmarshal([]byte(strmeta), &metadata); err != nil {
		return createThingEvent{}, err
	}

	cte := createThingEvent{
		id: read(event, "id", ""),
	}

	metadataOpcua, ok := metadata[keyType]
	if !ok {
		return createThingEvent{}, errMetadataType
	}

	metadataVal, ok := metadataOpcua.(map[string]interface{})
	if !ok {
		return createThingEvent{}, errMetadataFormat
	}

	val, ok := metadataVal[keyNodeID].(string)
	if !ok || val == "" {
		return createThingEvent{}, errMetadataNodeID
	}

	cte.opcuaNodeID = val
	return cte, nil
}

func decodeRemoveThing(event map[string]interface{}) removeThingEvent {
	return removeThingEvent{
		id: read(event, "id", ""),
	}
}

func decodeCreateChannel(event map[string]interface{}) (createChannelEvent, error) {
	strmeta := read(event, "metadata", "{}")
	var metadata map[string]interface{}
	if err := json.Unmarshal([]byte(strmeta), &metadata); err != nil {
		return createChannelEvent{}, err
	}

	cce := createChannelEvent{
		id: read(event, "id", ""),
	}

	metadataOpcua, ok := metadata[keyType]
	if !ok {
		return createChannelEvent{}, errMetadataType
	}

	metadataVal, ok := metadataOpcua.(map[string]interface{})
	if !ok {
		return createChannelEvent{}, errMetadataFormat
	}

	val, ok := metadataVal[keyServerURI].(string)
	if !ok || val == "" {
		return createChannelEvent{}, errMetadataServerURI
	}

	cce.opcuaServerURI = val
	return cce, nil
}

func decodeRemoveChannel(event map[string]interface{}) removeChannelEvent {
	return removeChannelEvent{
		id: read(event, "id", ""),
	}
}

func decodeConnectThing(event map[string]interface{}) connectThingEvent {
	return connectThingEvent{
		chanID:  read(event, "chan_id", ""),
		thingID: read(event, "thing_id", ""),
	}
}

func decodeDisconnectThing(event map[string]interface{}) connectThingEvent {
	return connectThingEvent{
		chanID:  read(event, "chan_id", ""),
		thingID: read(event, "thing_id", ""),
	}
}

func read(event map[string]interface{}, key, def string) string {
	val, ok := event[key].(string)
	if !ok {
		return def
	}

	return val
}
