package interposecandidate

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/pkg/errors"

	"github.com/networkservicemesh/api/pkg/api/networkservice"
	"github.com/networkservicemesh/api/pkg/api/registry"

	"github.com/networkservicemesh/sdk/pkg/networkservice/core/next"

	"github.com/networkservicemesh/sdk/pkg/registry/common/interpose/interposedata"
	"github.com/sirupsen/logrus"
)

type interposeCandidateServer struct {
	interposeData *interposedata.Map
	nseClient     registry.NetworkServiceEndpointRegistryClient
}

// NewServer - creates a NetworkServiceServer that tracks locally registered CrossConnect Endpoints and on Request
//				selects a candidate cross connect nse based on the ns labels of the request and the stored cross connects
//              nse's labels
func NewServer(registryServer *registry.NetworkServiceEndpointRegistryServer, nseClient registry.NetworkServiceEndpointRegistryClient) networkservice.NetworkServiceServer {
	rv := &interposeCandidateServer{
		interposeData: interposedata.NewMap(),
		nseClient:     nseClient,
	}
	*registryServer = interposedata.NewNetworkServiceEndpointDataRegistryServer(rv.interposeData)
	return rv
}

func (l *interposeCandidateServer) Request(ctx context.Context, request *networkservice.NetworkServiceRequest) (result *networkservice.Connection, err error) {
	conn := request.GetConnection()
	ind := conn.GetPath().GetIndex() // It is designed to be used inside Endpoint, so current index is Endpoint already
	connID := conn.GetId()

	logrus.Infof("interposeCandidateServer: Request: ind=%v connD=%v", ind, connID)

	if len(conn.GetPath().GetPathSegments()) == 0 || ind <= 0 {
		return nil, errors.Errorf("path segment doesn't have a client or cross connect nse identity")
	}

	crossCTX := ctx
	// Match function to be called for each registered cross NSE (using Range) to find a feasible one
	matchFunc := func(theKey string, crossNSE *registry.NetworkServiceEndpoint, i ...interface{}) bool {
		labels := make(map[string]string)
		if len(i) > 0 {
			// read the labels based on which the match making will happen
			if l, ok := i[0].(map[string]string); ok {
				labels = l
			}
		}
		if ok := matchEndpoint(labels, crossNSE); ok {
			logrus.Infof("interposeCandidateServer: Request: MATCH!!! theKey: \"%v\" crossNSE: %v", theKey, crossNSE)
			// save candidate crossNSE.name into context for interpose server to read (it will try to contact the url matching the candidate)
			crossCTX = WithCandidate(ctx, crossNSE)
			c, _ := Candidate(crossCTX)
			logrus.Infof("interposeCandidateServer: from ctx %v", c)
			// Found matching interpose nse, stop iterating.
			return false
		}

		logrus.Infof("interposeCandidateServer: Request: Non match! theKey: \"%v\" crossNSE: %v", theKey, crossNSE)

		// no matching interpose nse yet, keep iterating.
		return true
	}

	// Iterate over all cross connect NSEs to check one with passed state.
	// Try to pick a cross connect NSE based on labels of the request first
	if request.GetConnection().GetLabels() != nil {
		logrus.Infof("interposeCandidateServer: Request: Find cross NSE based on request labels=%v", request.GetConnection().GetLabels())
		l.interposeData.Range(matchFunc, request.GetConnection().GetLabels())
	} else {
		// If no labels in request, try to pick a cross connect NSE based on the labels of the NSE selected earlier
		// - look up NSE because of its labels based on its name
		// - only labels belonging to the appropriate NS are considered
		nseName := request.GetConnection().GetNetworkServiceEndpointName()
		nsName := request.GetConnection().GetNetworkService()
		if nseName != "" && nsName != "" {
			if nse, err := l.discoverNetworkServiceEndpoint(crossCTX, nseName); err == nil {
				logrus.Infof("interposeCandidateServer: Request: Find cross NSE for NS \"%v\" and NSE %v", nsName, nse)
				nseLabels := nse.GetNetworkServiceLabels()
				if val, ok := nseLabels[nsName]; ok && len(val.Labels) != 0 {
					l.interposeData.Range(matchFunc, val.Labels)
				}
			}
		}
	}

	return next.Server(crossCTX).Request(crossCTX, request)
}

func (l *interposeCandidateServer) Close(ctx context.Context, conn *networkservice.Connection) (*empty.Empty, error) {
	return next.Server(ctx).Close(ctx, conn)
}

func (l *interposeCandidateServer) discoverNetworkServiceEndpoint(ctx context.Context, nseName string) (*registry.NetworkServiceEndpoint, error) {
	query := &registry.NetworkServiceEndpointQuery{
		NetworkServiceEndpoint: &registry.NetworkServiceEndpoint{
			Name: nseName,
		},
	}

	nseStream, err := l.nseClient.Find(ctx, query)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	nseList := registry.ReadNetworkServiceEndpointList(nseStream)

	for _, nse := range nseList {
		if nse.Name == nseName {
			return nse, nil
		}
	}

	query.Watch = true

	nseStream, err = l.nseClient.Find(ctx, query)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	for {
		var nse *registry.NetworkServiceEndpoint
		if nse, err = nseStream.Recv(); err != nil {
			return nil, errors.WithStack(err)
		}

		if nse.Name == nseName {
			return nse, nil
		}
	}
}
