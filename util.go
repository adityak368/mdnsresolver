package mdns

import (
	"errors"
	"strings"

	"google.golang.org/grpc/resolver"
)

type targetInfo struct {
	instanceName string
	serviceName  string
	domain       string
	target       resolver.Target
}

func parseResolverTarget(target resolver.Target) (*targetInfo, error) {
	// mdns://serviceName/serviceInstanceName.domain
	endpoint := target.Endpoint
	authority := target.Authority

	if authority == "" {
		return nil, errors.New("Could not parse target. Invalid Authority")
	}
	if endpoint == "" {
		return nil, errors.New("Could not parse target. Invalid Endpoint")
	}

	ti := &targetInfo{target: target}
	parts := strings.Split(endpoint, ".")
	if len(parts) == 2 {
		ti.instanceName = parts[0]
		ti.serviceName = authority
		ti.domain = parts[1]
		return ti, nil
	}

	return nil, errors.New("Could not parse target. Invalid Endpoint. Could not find domain and serviceInstanceName")
}
