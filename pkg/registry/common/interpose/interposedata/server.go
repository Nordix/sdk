package interposedata

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"

	"github.com/networkservicemesh/api/pkg/api/registry"

	"github.com/networkservicemesh/sdk/pkg/registry/common/interpose"
	"github.com/networkservicemesh/sdk/pkg/registry/core/next"

	"github.com/sirupsen/logrus"
)

type interposeDataRegistryServer struct {
	interposeData *Map
}

// NewNetworkServiceEndpointDataRegistryServer - creates a NetworkServiceRegistryServer that registers local Cross connect Endpoints
//				including all their register information and adds them to Map
func NewNetworkServiceEndpointDataRegistryServer(interposeData *Map) registry.NetworkServiceEndpointRegistryServer {
	return &interposeDataRegistryServer{
		interposeData: interposeData,
	}
}

func (s *interposeDataRegistryServer) Register(ctx context.Context, nse *registry.NetworkServiceEndpoint) (*registry.NetworkServiceEndpoint, error) {
	logrus.Infof("interposeDataRegistryServer: Register: nseName=%v, nse=%v", nse.Name, nse)
	r, err := next.NetworkServiceEndpointRegistryServer(ctx).Register(ctx, nse)
	if err != nil {
		return nil, err
	}

	logrus.Infof("interposeDataRegistryServer: Register: result nseName=%v, nse=%v", r.Name, r)
	if interpose.Is(nse.Name) {
		logrus.Infof("interposeDataRegistryServer: Register: storing nse=%v", r)
		// its key must be kept in sync with interposeURLs' key!!!
		s.interposeData.Store(nse.Name, r.Clone())
	}

	return r, nil
}

func (s *interposeDataRegistryServer) Find(query *registry.NetworkServiceEndpointQuery, server registry.NetworkServiceEndpointRegistry_FindServer) error {
	logrus.Infof("interposeDataRegistryServer: Find: query=%v, server=%v", query, server)
	// No need to modify find logic.
	return next.NetworkServiceEndpointRegistryServer(server.Context()).Find(query, server)
}

func (s *interposeDataRegistryServer) Unregister(ctx context.Context, nse *registry.NetworkServiceEndpoint) (*empty.Empty, error) {
	logrus.Infof("interposeDataRegistryServer: Unregister: nse=%v", nse)
	if interpose.Is(nse.Name) {
		logrus.Infof("interposeDataRegistryServer: Unregister: remove nse.Name=%v", nse.Name)
		s.interposeData.Delete(nse.Name)
	}

	return next.NetworkServiceEndpointRegistryServer(ctx).Unregister(ctx, nse)
}

var _ registry.NetworkServiceEndpointRegistryServer = (*interposeDataRegistryServer)(nil)
