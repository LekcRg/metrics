package grpcapi

import (
	"context"

	"github.com/LekcRg/metrics/internal/logger"
	"github.com/LekcRg/metrics/internal/models"
	"github.com/LekcRg/metrics/internal/server/storage"
	pb "github.com/LekcRg/metrics/proto"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MetricService interface {
	UpdateMany(ctx context.Context, list []models.Metrics) error
}

type server struct {
	pb.UnimplementedMetricsServer
	service MetricService
}

func NewServer(s MetricService) *grpc.Server {
	grpcServer := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(
			logger.InterceptorLogger,
		),
	)

	metricsHandler := &server{
		service: s,
	}

	pb.RegisterMetricsServer(grpcServer, metricsHandler)

	return grpcServer
}

func (s *server) UpdateMetrics(
	ctx context.Context, in *pb.UpdateMetricsRequest,
) (*pb.UpdateMetricsResponse, error) {
	// err = validateSHA256(w, r, body, key)
	// if err != nil {
	// 	logger.Log.Error("Error while validating SHA256 hash ", zap.Error(err))
	// 	return
	// }

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
		logger.Log.Error(err.Error())
		return nil, status.Error(codes.DataLoss, "error from service")
	}

	res := &pb.UpdateMetricsResponse{}
	return res, nil
}
