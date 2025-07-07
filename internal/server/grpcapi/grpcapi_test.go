package grpcapi

import (
	"context"
	"errors"
	"testing"

	"github.com/LekcRg/metrics/internal/config"
	"github.com/LekcRg/metrics/internal/models"
	"github.com/LekcRg/metrics/internal/server/storage"
	pb "github.com/LekcRg/metrics/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type mockMetricService struct {
	errToReturn     error
	receivedMetrics []models.Metrics
}

func (m *mockMetricService) UpdateMany(ctx context.Context, list []models.Metrics) error {
	m.receivedMetrics = list
	return m.errToReturn
}

func TestUpdateMetrics_HappyPath(t *testing.T) {
	mockService := &mockMetricService{}
	serverConfig := config.ServerConfig{}

	grpcServer := &server{
		service: mockService,
		config:  serverConfig,
	}

	request := &pb.UpdateMetricsRequest{
		Metrics: []*pb.Metric{
			{
				Id:    "TestGauge",
				MType: pb.Metric_GAUGE,
				Value: floatPtr(123.45),
			},
			{
				Id:    "TestCounter",
				MType: pb.Metric_COUNTER,
				Delta: intPtr(10),
			},
		},
	}

	response, err := grpcServer.UpdateMetrics(context.Background(), request)

	require.NoError(t, err)
	require.NotNil(t, response)

	expectedMetrics := []models.Metrics{
		{
			ID:    "TestGauge",
			MType: "gauge",
			Value: gaugePtr(123.45),
		},
		{
			ID:    "TestCounter",
			MType: "counter",
			Delta: counterPtr(10),
		},
	}
	assert.Equal(t, expectedMetrics, mockService.receivedMetrics)
}

func TestUpdateMetrics_ServiceError(t *testing.T) {
	mockService := &mockMetricService{
		errToReturn: errors.New("database is down"),
	}
	serverConfig := config.ServerConfig{}

	grpcServer := &server{
		service: mockService,
		config:  serverConfig,
	}

	request := &pb.UpdateMetricsRequest{
		Metrics: []*pb.Metric{
			{
				Id:    "any",
				MType: pb.Metric_GAUGE,
				Value: floatPtr(1),
			},
		},
	}

	response, err := grpcServer.UpdateMetrics(context.Background(), request)

	require.Error(t, err)
	assert.Nil(t, response)

	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Internal, st.Code())
}

func TestUpdateMetrics_InvalidHMAC(t *testing.T) {
	mockService := &mockMetricService{}
	serverConfig := config.ServerConfig{
		CommonConfig: config.CommonConfig{
			Key: "my-secret-key",
		},
	}

	grpcServer := &server{
		service: mockService,
		config:  serverConfig,
	}

	request := &pb.UpdateMetricsRequest{
		Metrics: []*pb.Metric{
			{
				Id:    "any",
				MType: pb.Metric_GAUGE,
				Value: floatPtr(1),
			},
		},
	}

	md := metadata.New(map[string]string{"HashSHA256": "invalid-hash"})
	ctx := metadata.NewIncomingContext(context.Background(), md)

	response, err := grpcServer.UpdateMetrics(ctx, request)

	require.Error(t, err)
	assert.Nil(t, response)

	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.PermissionDenied, st.Code())

	assert.Empty(t, mockService.receivedMetrics)
}

func gaugePtr(v storage.Gauge) *storage.Gauge {
	return &v
}

func intPtr(v int64) *int64 {
	return &v
}

func counterPtr(v storage.Counter) *storage.Counter {
	return &v
}
func floatPtr(v float64) *float64 {
	return &v
}
