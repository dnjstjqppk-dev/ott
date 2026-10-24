package clients

import (
	billingv1 "github.com/netflix-clone/proto/billing/v1"
	catalogv1 "github.com/netflix-clone/proto/catalog/v1"
	iamv1 "github.com/netflix-clone/proto/iam/v1"
	playbackv1 "github.com/netflix-clone/proto/playback/v1"
	recov1 "github.com/netflix-clone/proto/reco/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ServiceClients struct {
	Iam      iamv1.IamServiceClient
	Catalog  catalogv1.CatalogServiceClient
	Billing  billingv1.BillingServiceClient
	Playback playbackv1.PlaybackServiceClient
	Reco     recov1.RecommendationServiceClient
	conns    []*grpc.ClientConn
}

type EndpointsConfig struct {
	IamAddr      string
	CatalogAddr  string
	BillingAddr  string
	PlaybackAddr string
	RecoAddr     string
}

func NewServiceClients(cfg EndpointsConfig) (*ServiceClients, error) {
	dial := func(addr string) (*grpc.ClientConn, error) {
		return grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}
	iamConn, err := dial(cfg.IamAddr)
	if err != nil {
		return nil, err
	}
	catConn, err := dial(cfg.CatalogAddr)
	if err != nil {
		return nil, err
	}
	bilConn, err := dial(cfg.BillingAddr)
	if err != nil {
		return nil, err
	}
	pbConn, err := dial(cfg.PlaybackAddr)
	if err != nil {
		return nil, err
	}
	rcConn, err := dial(cfg.RecoAddr)
	if err != nil {
		return nil, err
	}
	return &ServiceClients{
		Iam:      iamv1.NewIamServiceClient(iamConn),
		Catalog:  catalogv1.NewCatalogServiceClient(catConn),
		Billing:  billingv1.NewBillingServiceClient(bilConn),
		Playback: playbackv1.NewPlaybackServiceClient(pbConn),
		Reco:     recov1.NewRecommendationServiceClient(rcConn),
		conns:    []*grpc.ClientConn{iamConn, catConn, bilConn, pbConn, rcConn},
	}, nil
}

func (s *ServiceClients) CloseAll() {
	for _, c := range s.conns {
		_ = c.Close()
	}
}

