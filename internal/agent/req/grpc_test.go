package req

import (
	"context"
	"crypto/x509"
	"net"
	"path/filepath"
	"testing"

	"github.com/LekcRg/metrics/internal/config"
	"github.com/LekcRg/metrics/internal/crypto"
	"github.com/LekcRg/metrics/internal/models"
	"github.com/LekcRg/metrics/internal/server/storage"
	pb "github.com/LekcRg/metrics/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/proto"
)

// Мок-сервер
type mockServer struct {
	pb.UnimplementedMetricsServer
	recieved    *pb.UpdateMetricsRequest
	recievedCtx context.Context
}

func (s *mockServer) UpdateMetrics(
	ctx context.Context, in *pb.UpdateMetricsRequest,
) (*pb.UpdateMetricsResponse, error) {
	s.recieved = in
	s.recievedCtx = ctx
	return &pb.UpdateMetricsResponse{}, nil
}

func getConn(t *testing.T) (*grpc.ClientConn, *mockServer) {
	lis := bufconn.Listen(1024 * 1024)

	s := grpc.NewServer()
	srv := &mockServer{}
	pb.RegisterMetricsServer(s, srv)

	go func() {
		if err := s.Serve(lis); err != nil {
			t.Errorf("Server exited with error: %v", err)
		}
	}()

	conn, err := grpc.NewClient(
		"passthrough://bufnet",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
			return lis.Dial()
		}),
	)

	require.NoError(t, err)
	return conn, srv
}

func checkList(t *testing.T, want, list []*pb.Metric) {
	for i, m := range list {
		assert.Equal(t, want[i].Id, m.Id)
		assert.Equal(t, want[i].MType, m.MType)
		assert.Equal(t, want[i].Delta, m.Delta)
		assert.Equal(t, want[i].Value, m.Value)
	}
}

var (
	counterMetric = models.Metrics{
		ID:    "counter",
		MType: "counter",
		Delta: counterPtr(1),
	}
	counterMetricPb = pb.Metric{
		Id:    "counter",
		MType: pb.Metric_COUNTER,
		Delta: intPtr(1),
	}
	gaugeMetric = models.Metrics{
		ID:    "gauge",
		MType: "gauge",
		Value: gaugePtr(1),
	}
	gaugeMetricPb = pb.Metric{
		Id:    "gauge",
		MType: pb.Metric_GAUGE,
		Value: floatPtr(1),
	}
	list = []models.Metrics{
		counterMetric,
		gaugeMetric,
	}
	wantList = []*pb.Metric{
		&counterMetricPb,
		&gaugeMetricPb,
	}
)

func TestGRPCRequest(t *testing.T) {
	type test struct {
		name         string
		sshKey       string
		pub          string
		metrics      []models.Metrics
		wantList     []*pb.Metric
		wantErr      bool
		notCheckList bool
	}

	tests := []test{
		{
			name: "Correct req",
			metrics: []models.Metrics{
				counterMetric,
			},
			wantList: []*pb.Metric{
				&counterMetricPb,
			},
			wantErr: false,
		},
		{
			name:         "Empty list",
			metrics:      []models.Metrics{},
			wantErr:      true,
			notCheckList: true,
		},
	}

	conn, srv := getConn(t)

	defer conn.Close()

	for _, tt := range tests {
		func() {
			cl := NewGRPCClientWithConn(conn, config.AgentConfig{})

			err := cl.GRPCRequest(context.Background(), tt.metrics)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if !tt.notCheckList {
				// assert.Equal(t, tt.wantList, srv.recieved.Metrics)
				checkList(t, tt.wantList, srv.recieved.Metrics)
			}
		}()
	}
}

func TestGRPCRequestRSA(t *testing.T) {
	const (
		pathToPriv = "../../testdata/keys/priv.pem"
		pathToPub  = "../../testdata/keys/pub.pem"
	)

	privPath, err := filepath.Abs(pathToPriv)
	t.Log(privPath)
	require.NoError(t, err)
	pemPriv, err := crypto.ParsePEMFile(privPath)
	require.NoError(t, err)
	priv, err := x509.ParsePKCS1PrivateKey(pemPriv)
	require.NoError(t, err)

	pubPath, err := filepath.Abs(pathToPub)
	require.NoError(t, err)
	pemPub, err := crypto.ParsePEMFile(pubPath)
	require.NoError(t, err)
	pub, err := x509.ParsePKCS1PublicKey(pemPub)
	require.NoError(t, err)

	list := []models.Metrics{
		counterMetric,
		gaugeMetric,
	}
	wantList := []*pb.Metric{
		&counterMetricPb,
		&gaugeMetricPb,
	}

	conn, srv := getConn(t)
	defer conn.Close()
	cl := NewGRPCClientWithConn(conn, config.AgentConfig{
		PublicKey: pub,
	})

	err = cl.GRPCRequest(context.Background(), list)
	require.NoError(t, err)

	decrypted, err := crypto.DecryptRSA(srv.recieved.Encrypted, priv)
	require.NoError(t, err)

	in := &pb.UpdateMetricsRequest{}
	err = proto.Unmarshal(decrypted, in)
	require.NoError(t, err)

	checkList(t, wantList, srv.recieved.Metrics)
}

func TestGRPCRequestHMAC(t *testing.T) {
	conn, srv := getConn(t)
	key := "secret-key"
	cl := NewGRPCClientWithConn(conn, config.AgentConfig{
		CommonConfig: config.CommonConfig{
			Key: key,
		},
	})

	ctx := context.Background()
	err := cl.GRPCRequest(ctx, list)
	require.NoError(t, err)

	wantReq := &pb.UpdateMetricsRequest{
		Metrics: wantList,
	}
	b, err := proto.Marshal(wantReq)
	require.NoError(t, err)
	wantHmac := crypto.GenerateHMAC(b, key)

	err = cl.GRPCRequest(context.Background(), list)
	require.NoError(t, err)

	checkList(t, wantList, srv.recieved.Metrics)

	md, ok := metadata.FromIncomingContext(srv.recievedCtx)
	headerHash := ""
	if ok {
		headerHash = md.Get("HashSHA256")[0]
	}

	assert.Equal(t, wantHmac, headerHash)
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
