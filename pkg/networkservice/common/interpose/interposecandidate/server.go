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
}

// NewServer - creates a NetworkServiceServer that tracks locally registered CrossConnect Endpoints and on Request
//				selects a candidate cross connect nse based on the ns labels of the request and the stored cross connects
//              nse's labels
func NewServer(registryServer *registry.NetworkServiceEndpointRegistryServer) networkservice.NetworkServiceServer {
	rv := &interposeCandidateServer{
		interposeData: interposedata.NewMap(),
	}
	*registryServer = interposedata.NewNetworkServiceEndpointDataRegistryServer(rv.interposeData)
	return rv
}

func (l *interposeCandidateServer) Request(ctx context.Context, request *networkservice.NetworkServiceRequest) (result *networkservice.Connection, err error) {
	conn := request.GetConnection()
	ind := conn.GetPath().GetIndex() // It is designed to be used inside Endpoint, so current index is Endpoint already
	connID := conn.GetId()

	logrus.Infof("EZOLLUG: interposeCandidateServer: Request: ind=%v connD=%v", ind, connID)

	if len(conn.GetPath().GetPathSegments()) == 0 || ind <= 0 {
		return nil, errors.Errorf("path segment doesn't have a client or cross connect nse identity")
	}

	// Iterate over all cross connect NSEs to check one with passed state.
	crossCTX := ctx
	l.interposeData.Range(func(theKey string, crossNSE *registry.NetworkServiceEndpoint) bool {
		if ok := matchEndpoint3(request.GetConnection().GetLabels(), crossNSE); ok {
			logrus.Infof("EZOLLUG: interposeCandidateServer: Request: MATCH!!! theKey=%v crossNSE=%v reqLabels=%v", theKey, crossNSE, request.GetConnection().GetLabels())
			// save candidate crossNSE.name into context for interpose server to read (it will try to contact the url matching the candidate)
			crossCTX = WithCandidate(ctx, crossNSE)
			c, _ := Candidate(crossCTX)
			logrus.Infof("EZOLLUG: interposeCandidateServer: from ctx %v", c)
			// Found matching interpose nse, stop iterating.
			return false
		}

		logrus.Infof("EZOLLUG: interposeCandidateServer: Request: No match! theKey=%v crossNSE=%v reqLabels=%v", theKey, crossNSE, request.GetConnection().GetLabels())

		return true
	})

	return next.Server(crossCTX).Request(crossCTX, request)
}

func (l *interposeCandidateServer) Close(ctx context.Context, conn *networkservice.Connection) (*empty.Empty, error) {
	return next.Server(ctx).Close(ctx, conn)
}
