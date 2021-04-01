package interposecandidate

import (
	"context"

	"github.com/networkservicemesh/api/pkg/api/registry"
)

const (
	candidatesKey contextKeyType = "InterposeCandidates"
)

type contextKeyType string

// NetworkServiceInterposeCandidate -
//    Interpose NSE candidate to handle a service request
type NetworkServiceInterposeCandidate struct {
	interposeName string
}

// WithCandidate -
//    Wraps 'parent' in a new Context that has the Candidate as NetworkServiceEndpoint name
func WithCandidate(parent context.Context, candidate *registry.NetworkServiceEndpoint) context.Context {
	if parent == nil {
		panic("cannot create context from nil parent")
	}
	return context.WithValue(parent, candidatesKey, &NetworkServiceInterposeCandidate{
		interposeName: candidate.Name,
	})
}

// Candidate -
//   Returns the Candidate as NetworkServiceEndpoint name
func Candidate(ctx context.Context) (string, bool) {
	if rv, ok := ctx.Value(candidatesKey).(*NetworkServiceInterposeCandidate); ok {
		return rv.interposeName, true
	}
	return "", false
}
