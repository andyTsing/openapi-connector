package refImpl

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-ocf/kit/cqrs/eventbus/nats"
	"github.com/go-ocf/kit/cqrs/eventstore/mongodb"
	"github.com/go-ocf/kit/log"
	"github.com/go-ocf/openapi-connector/service"
	storeMongodb "github.com/go-ocf/openapi-connector/store/mongodb"
	"github.com/panjf2000/ants"
)

type Config struct {
	Log               log.Config
	MongoDB           mongodb.Config
	Nats              nats.Config
	Service           service.Config
	GoRoutinePoolSize int `envconfig:"GOROUTINE_POOL_SIZE" default:"16"`
	StoreMongoDB      storeMongodb.Config
}

//String return string representation of Config
func (c Config) String() string {
	b, _ := json.MarshalIndent(c, "", "  ")
	return fmt.Sprintf("config: \n%v\n", string(b))
}

func Init(config Config) (*service.Server, error) {
	log.Setup(config.Log)

	pool, err := ants.NewPool(config.GoRoutinePoolSize)
	if err != nil {
		return nil, fmt.Errorf("cannot create goroutine pool: %v", err)
	}

	resourceEventstore, err := mongodb.NewEventStore(config.MongoDB, pool.Submit)
	if err != nil {
		return nil, fmt.Errorf("cannot create resource mongodb eventstore %v", err)
	}

	resourceSubscriber, err := nats.NewSubscriber(config.Nats, pool.Submit, func(err error) { log.Errorf("error occurs during receiving event: %v", err) })
	if err != nil {
		return nil, fmt.Errorf("cannot create resource nats subscriber %v", err)
	}

	store, err := storeMongodb.NewStore(context.Background(), config.StoreMongoDB)
	if err != nil {
		return nil, fmt.Errorf("cannot create mongodb store %v", err)
	}

	log.Info(config.String())

	return service.New(config.Service, resourceEventstore, resourceSubscriber, store), nil
}
