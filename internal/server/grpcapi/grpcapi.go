package grpcapi

import (
	"context"

	"github.com/LekcRg/metrics/internal/config"
	"github.com/LekcRg/metrics/internal/crypto"
	"github.com/LekcRg/metrics/internal/logger"
	"github.com/LekcRg/metrics/internal/models"
	"github.com/LekcRg/metrics/internal/server/storage"
	pb "github.com/LekcRg/metrics/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
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
	err := crypto.GetAndValidHMAC(ctx, s.config.Key, in)
	if err != nil {
		return nil, err
	}

	if s.config.PrivateKey != nil {
		var b []byte
		b, err = crypto.DecryptRSA(in.Encrypted, s.config.PrivateKey)
		if err != nil {
			return nil, status.Error(codes.PermissionDenied, "Permission denied")
		}

		err = proto.Unmarshal(b, in)
		if err != nil {
			return nil, status.Error(codes.Internal, "Internal server error")
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

	err = s.service.UpdateMany(ctx, list)
	if err != nil {
		logger.Log.Error("Error from UpdateMany service", zap.Error(err))
		return nil, status.Error(codes.Internal, "error from service")
	}

	res := &pb.UpdateMetricsResponse{}
	return res, nil
}
