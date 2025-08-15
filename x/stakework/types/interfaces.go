package types

import (
	eventtypes "github.com/hetu-project/hetu/v1/x/event/types"
)

// EventKeeper event module interface - used to get data from event module
// Deprecated: Use eventtypes.EventKeeper directly instead
type EventKeeper = eventtypes.EventKeeper
