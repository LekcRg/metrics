package req

import (
	"context"
	"errors"

	"github.com/LekcRg/metrics/internal/logger"
	"github.com/LekcRg/metrics/internal/models"
	pb "github.com/LekcRg/metrics/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GRPCClient struct {
	conn   *grpc.ClientConn
	client pb.MetricsClient
}

func NewGRPCClient(addr string) *GRPCClient {
	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		logger.Log.Fatal("GRPC Connect error", zap.Error(err))
	}

	client := pb.NewMetricsClient(conn)

	return &GRPCClient{
		conn:   conn,
		client: client,
	}
}

func (g *GRPCClient) GRPCRequest(ctx context.Context, metrics []models.Metrics) error {
	if len(metrics) == 0 || metrics == nil {
		return errors.New("empty metrics list")
	}

	list := make([]*pb.Metric, 0, len(metrics))
	for _, m := range metrics {
		mtype := pb.Metric_GAUGE
		if m.MType == "counter" {
			mtype = pb.Metric_COUNTER
		}
		list = append(list, &pb.Metric{
			Id:    m.ID,
			MType: mtype,
			Delta: (*int64)(m.Delta),
			Value: (*float64)(m.Value),
		})
	}

	_, err := g.client.UpdateMetrics(ctx, &pb.UpdateMetricsRequest{
		Metrics: list,
	})
	return err
}

func (g *GRPCClient) Shutdown() {
	if err := g.conn.Close(); err != nil {
		logger.Log.Error("GRPC conn close error", zap.Error(err))
	}
}
