package grpcapi

import (
	"context"

	"github.com/LekcRg/metrics/internal/config"
	"github.com/LekcRg/metrics/internal/crypto"
	"github.com/LekcRg/metrics/internal/logger"
	"github.com/LekcRg/metrics/internal/models"
	"github.com/LekcRg/metrics/internal/server/storage"
	pb "github.com/LekcRg/metrics/proto"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

type MetricService interface {
	UpdateMany(ctx context.Context, list []models.Metrics) error
}

type server struct {
	pb.UnimplementedMetricsServer
	service MetricService
	config  config.ServerConfig
}

func NewServer(s MetricService, cfg config.ServerConfig) *grpc.Server {
	grpcServer := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(
			logger.InterceptorLogger,
		),
	)

	metricsHandler := &server{
		service: s,
		config:  cfg,
	}

	pb.RegisterMetricsServer(grpcServer, metricsHandler)

	return grpcServer
}

func (s *server) UpdateMetrics(
	ctx context.Context, in *pb.UpdateMetricsRequest,
) (*pb.UpdateMetricsResponse, error) {
	if s.config.Key != "" {
		md, ok := metadata.FromIncomingContext(ctx)
		headerHash := ""
		if ok {
			headerHash = md.Get("HashSHA256")[0]
		}

		if headerHash == "" {
			return nil, status.Error(codes.PermissionDenied, "Empty HashSHA256")
		}

		ctx = metadata.NewOutgoingContext(ctx, md)

		inBytes, err := proto.Marshal(in)
		if err != nil {
			return nil, status.Error(codes.Internal, "Internal server error")
		}

		hash := crypto.GenerateHMAC(inBytes, s.config.Key)
		if hash != headerHash {
			return nil, status.Error(codes.PermissionDenied, "Hash is not correct")
		}
	}

	list := make([]models.Metrics, 0, len(in.Metrics))

	for _, m := range in.Metrics {
		mtype := "gauge"
		if m.MType == pb.Metric_COUNTER {
			mtype = "counter"
		}
		list = append(list, models.Metrics{
			Delta: (*storage.Counter)(m.Delta),
			Value: (*storage.Gauge)(m.Value),
			MType: mtype,
			ID:    m.Id,
		})
	}

	err := s.service.UpdateMany(ctx, list)
	if err != nil {
		logger.Log.Error("Error from UpdateMany service", zap.Error(err))
		return nil, status.Error(codes.Internal, "error from service")
	}

	res := &pb.UpdateMetricsResponse{}
	return res, nil
}
