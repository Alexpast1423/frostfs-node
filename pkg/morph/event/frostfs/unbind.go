package frostfs

import (
	"fmt"

	"github.com/TrueCloudLab/frostfs-node/pkg/morph/event"
	"github.com/nspcc-dev/neo-go/pkg/core/state"
)

type Unbind struct {
	bindCommon
}

func ParseUnbind(e *state.ContainedNotificationEvent) (event.Event, error) {
	var (
		ev  Unbind
		err error
	)

	params, err := event.ParseStackArray(e)
	if err != nil {
		return nil, fmt.Errorf("could not parse stack items from notify event: %w", err)
	}

	err = parseBind(&ev.bindCommon, params)
	if err != nil {
		return nil, err
	}

	ev.txHash = e.Container

	return ev, nil
}
