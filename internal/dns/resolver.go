package dns

import (
	"context"
	"net"
)

type SystemResolver struct {
	*net.Resolver
}

func NewSystemResolver() *SystemResolver {
	return &SystemResolver{
		&net.Resolver{PreferGo: true},
	}
}

func (r *SystemResolver) Resolve(ctx context.Context, host string, _ []uint16) ([]net.IPAddr, error) {
	addrs, err := r.LookupIPAddr(ctx, host)
	if err != nil {
		return []net.IPAddr{}, err
	}
	return addrs, nil
}
