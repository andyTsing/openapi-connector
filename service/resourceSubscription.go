package service

import (
	"context"
	"fmt"

	"github.com/go-ocf/go-coap"

	pbCQRS "github.com/go-ocf/kit/cqrs/pb"
	kitHttp "github.com/go-ocf/kit/http"
	"github.com/go-ocf/openapi-connector/events"
	"github.com/go-ocf/openapi-connector/store"
	raCqrs "github.com/go-ocf/resource-aggregate/cqrs"
	pbRA "github.com/go-ocf/resource-aggregate/pb"
)

func (s *SubscribeManager) subscribeToResource(ctx context.Context, l store.LinkedAccount, correlationID, signingSecret, deviceID, resourceHrefLink string) (string, error) {
	resp, err := subscribe(ctx, "/devices/"+deviceID+"/"+resourceHrefLink+"/subscriptions", correlationID, events.SubscriptionRequest{
		URL:           s.eventsURL,
		EventType:     []events.EventType{events.EventType_ResourceContentChanged},
		SigningSecret: signingSecret,
	}, l)
	if err != nil {
		return "", fmt.Errorf("cannot subscribe to device %v for %v: %v", deviceID, l.ID, err)
	}
	return resp.SubscriptionId, nil
}

func cancelResourceSubscription(ctx context.Context, l store.LinkedAccount, deviceID, resourceID, subscriptionID string) error {
	err := cancelSubscription(ctx, "/devices/"+deviceID+"/"+resourceID+"/subscriptions/"+subscriptionID, l)
	if err != nil {
		return fmt.Errorf("cannot cancel resource subscription for %v: %v", l.ID, err)
	}
	return nil
}

func (s *SubscribeManager) HandleResourceContentChangedEvent(ctx context.Context, subscriptionData subscriptionData, header events.EventHeader, body []byte) error {
	userID, err := subscriptionData.linkedAccount.OriginCloud.AccessToken.GetSubject()
	if err != nil {
		return fmt.Errorf("cannot get userID for device (%v) resource (%v) content changed: %v", subscriptionData.subscription.DeviceID, subscriptionData.subscription.Href, err)
	}

	coapContentFormat := int32(-1)
	switch header.ContentType {
	case coap.AppCBOR.String():
		coapContentFormat = int32(coap.AppCBOR)
	case coap.AppOcfCbor.String():
		coapContentFormat = int32(coap.AppOcfCbor)
	case coap.AppJSON.String():
		coapContentFormat = int32(coap.AppJSON)
	}

	_, err = s.raClient.NotifyResourceContentChanged(ctx, &pbRA.NotifyResourceContentChangedRequest{
		AuthorizationContext: &pbCQRS.AuthorizationContext{
			UserId:      userID,
			AccessToken: string(subscriptionData.linkedAccount.OriginCloud.AccessToken),
			DeviceId:    subscriptionData.subscription.DeviceID,
		},
		ResourceId: raCqrs.MakeResourceId(subscriptionData.subscription.DeviceID, kitHttp.CanonicalHref(subscriptionData.subscription.Href)),
		CommandMetadata: &pbCQRS.CommandMetadata{
			ConnectionId: OpenapiConnectorConnectionId,
			Sequence:     header.SequenceNumber,
		},
		Content: &pbRA.Content{
			Data:              body,
			ContentType:       header.ContentType,
			CoapContentFormat: coapContentFormat,
		},
	})
	if err != nil {
		return fmt.Errorf("cannot update resource aggregate (%v) resource (%v) content changed: %v", subscriptionData.subscription.DeviceID, subscriptionData.subscription.Href, err)
	}

	return nil

}

func (s *SubscribeManager) HandleResourceEvent(ctx context.Context, header events.EventHeader, body []byte, subscriptionData subscriptionData) error {
	switch header.EventType {
	case events.EventType_ResourceContentChanged:
		return s.HandleResourceContentChangedEvent(ctx, subscriptionData, header, body)
	}
	return fmt.Errorf("cannot handle resource event: unsupported Event-Type %v", header.EventType)
}
