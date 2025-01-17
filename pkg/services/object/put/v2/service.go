package putsvc

import (
	"context"
	"fmt"

	"github.com/TrueCloudLab/frostfs-node/pkg/services/object"
	putsvc "github.com/TrueCloudLab/frostfs-node/pkg/services/object/put"
	"github.com/TrueCloudLab/frostfs-node/pkg/services/object/util"
)

// Service implements Put operation of Object service v2.
type Service struct {
	*cfg
}

// Option represents Service constructor option.
type Option func(*cfg)

type cfg struct {
	svc        *putsvc.Service
	keyStorage *util.KeyStorage
}

// NewService constructs Service instance from provided options.
func NewService(opts ...Option) *Service {
	c := new(cfg)

	for i := range opts {
		opts[i](c)
	}

	return &Service{
		cfg: c,
	}
}

// Put calls internal service and returns v2 object streamer.
func (s *Service) Put(ctx context.Context) (object.PutObjectStream, error) {
	stream, err := s.svc.Put(ctx)
	if err != nil {
		return nil, fmt.Errorf("(%T) could not open object put stream: %w", s, err)
	}

	return &streamer{
		stream:     stream,
		keyStorage: s.keyStorage,
	}, nil
}

func WithInternalService(v *putsvc.Service) Option {
	return func(c *cfg) {
		c.svc = v
	}
}

func WithKeyStorage(ks *util.KeyStorage) Option {
	return func(c *cfg) {
		c.keyStorage = ks
	}
}
