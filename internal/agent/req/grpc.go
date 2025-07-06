package req

import (
	"context"
	"errors"

	"github.com/LekcRg/metrics/internal/config"
	"github.com/LekcRg/metrics/internal/crypto"
	"github.com/LekcRg/metrics/internal/logger"
	"github.com/LekcRg/metrics/internal/models"
	pb "github.com/LekcRg/metrics/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
)

type GRPCClient struct {
	conn   *grpc.ClientConn
	client pb.MetricsClient
	config config.AgentConfig
}

func NewGRPCClient(cfg config.AgentConfig) *GRPCClient {
	conn, err := grpc.NewClient(
		cfg.Addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		logger.Log.Fatal("GRPC Connect error", zap.Error(err))
	}

	client := pb.NewMetricsClient(conn)

	return &GRPCClient{
		conn:   conn,
		client: client,
		config: cfg,
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

	req := &pb.UpdateMetricsRequest{
		Metrics: list,
	}

	if g.config.Key != "" {
		b, err := proto.Marshal(req)
		if err != nil {
			logger.Log.Error("Error while marshal UpdateMetricsRequest pb", zap.Error(err))
		}
		sha := crypto.GenerateHMAC(b, g.config.Key)

		md := metadata.New(map[string]string{"HashSHA256": sha})
		ctx = metadata.NewOutgoingContext(ctx, md)
	}

	_, err := g.client.UpdateMetrics(ctx, req)
	return err
}

func (g *GRPCClient) Shutdown() {
	if err := g.conn.Close(); err != nil {
		logger.Log.Error("GRPC conn close error", zap.Error(err))
	}
}
