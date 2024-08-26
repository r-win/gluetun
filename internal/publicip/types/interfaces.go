package types 

import (
	"context"
	"net/netip"

	"github.com/qdm12/gluetun/internal/models"
)

type Fetcher interface {
	FetchInfo(ctx context.Context, logger Logger, ip netip.Addr) (
		result models.PublicIP, err error)
}

type Logger interface {
	Info(s string)
	Warn(s string)
	Error(s string)
}
