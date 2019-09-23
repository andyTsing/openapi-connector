package service

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"

	"github.com/go-ocf/cqrs/eventbus"
	cqrsEventStore "github.com/go-ocf/cqrs/eventstore"
	"github.com/go-ocf/kit/log"
	"github.com/go-ocf/kit/net/grpc"
	"github.com/go-ocf/kit/security"
	connectorStore "github.com/go-ocf/openapi-connector/store"

	pbAS "github.com/go-ocf/authorization/pb"
	projectionRA "github.com/go-ocf/resource-aggregate/cqrs/projection"
	pbRA "github.com/go-ocf/resource-aggregate/pb"
)

//Server handle HTTP request
type Server struct {
	server  *http.Server
	cfg     Config
	handler *RequestHandler
	ln      net.Listener
}

type loadDeviceSubscriptionsHandler struct {
	resourceProjection *projectionRA.Projection
}

func (h *loadDeviceSubscriptionsHandler) Handle(ctx context.Context, iter connectorStore.SubscriptionIter) error {
	var sub connectorStore.Subscription
	for iter.Next(ctx, &sub) {
		_, err := h.resourceProjection.Register(ctx, sub.DeviceID)
		if err != nil {
			log.Errorf("cannot register device %v subscription to resource projection: %v", sub.DeviceID, err)
		}
	}
	return iter.Err()
}

//New create new Server with provided store and bus
func New(config Config, resourceEventStore cqrsEventStore.EventStore, resourceSubscriber eventbus.Subscriber, store connectorStore.Store) *Server {

	tlsConfig, err := security.NewTLSConfigFromConfiguration(config.TLSConfig, security.VerifyClientCertificate)
	if err != nil {
		log.Fatalf("cannot setup tls configuration for service: %v", err)
	}
	ln, err := tls.Listen("tcp", config.Addr, tlsConfig)
	if err != nil {
		log.Fatalf("cannot listen and serve: %v", err)
	}

	raConn, err := grpc.NewClientConn(config.ResourceAggregateHost, config.TLSConfig)
	if err != nil {
		log.Fatalf("cannot create server: %v", err)
	}
	raClient := pbRA.NewResourceAggregateClient(raConn)

	asConn, err := grpc.NewClientConn(config.AuthHost, config.TLSConfig)
	if err != nil {
		log.Fatalf("cannot create server: %v", err)
	}
	asClient := pbAS.NewAuthorizationServiceClient(asConn)

	ctx := context.Background()

	resourceProjection, err := projectionRA.NewProjection(ctx, config.FQDN, resourceEventStore, resourceSubscriber, newResourceCtx(store, raClient))
	if err != nil {
		log.Fatalf("cannot create server: %v", err)
	}

	// load resource subscritpion
	h := loadDeviceSubscriptionsHandler{
		resourceProjection: resourceProjection,
	}
	err = store.LoadSubscriptions(ctx, []connectorStore.SubscriptionQuery{
		connectorStore.SubscriptionQuery{
			Type: connectorStore.Type_Device,
		},
	}, &h)
	if err != nil {
		log.Fatalf("cannot create server: %v", err)
	}

	requestHandler := NewRequestHandler(config.OriginCloud, config.OAuthCallback, NewSubscriptionManager(config.EventsURL, asClient, raClient, store, resourceProjection), asClient, raClient, resourceProjection, store)

	server := Server{
		server:  NewHTTP(requestHandler),
		cfg:     config,
		handler: requestHandler,
		ln:      ln,
	}

	return &server
}

// Serve starts the service's HTTP server and blocks.
func (s *Server) Serve() error {
	return s.server.Serve(s.ln)
}

// Shutdown ends serving
func (s *Server) Shutdown() error {
	return s.server.Shutdown(context.Background())
}
