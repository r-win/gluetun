package common

import (
	"context"
	"errors"
	"net/netip"

	"github.com/qdm12/gluetun/internal/models"
	"github.com/qdm12/gluetun/internal/publicip/types"
	"github.com/qdm12/gluetun/internal/updater/resolver"
)

var (
	ErrNotEnoughServers    = errors.New("not enough servers found")
	ErrHTTPStatusCodeNotOK = errors.New("HTTP status code not OK")
)

type Fetcher interface {
	FetchServers(ctx context.Context, minServers int) (servers []models.Server, err error)
}

type ParallelResolver interface {
	Resolve(ctx context.Context, settings resolver.ParallelSettings) (
		hostToIPs map[string][]netip.Addr, warnings []string, err error)
}

type Unzipper interface {
	FetchAndExtract(ctx context.Context, url string) (
		contents map[string][]byte, err error)
}

type Warner interface {
	Warn(s string)
}

type IPFetcher interface {
	FetchInfo(ctx context.Context, logger types.Logger, ip netip.Addr) (result models.PublicIP, err error)
}
