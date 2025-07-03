package ip

import (
	"net"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"testing"
)

func TestGetOutboundIP(t *testing.T) {
	ip, err := GetOutboundIP()
	if err != nil {
		t.Fatalf("GetOutboundIP() failed: %v", err)
	}

	if ip == "" {
		t.Error("GetOutboundIP() returned empty string")
	}

	// Проверяем, что возвращаемое значение является валидным IP
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		t.Errorf("GetOutboundIP() returned invalid IP: %s", ip)
	}

	// Проверяем, что IP не является loopback
	if parsedIP.IsLoopback() {
		t.Errorf("GetOutboundIP() returned loopback IP: %s", ip)
	}
}

func TestFilterMiddleware(t *testing.T) {
	network := netip.MustParsePrefix("192.168.1.0/24")
	ipv6Network := netip.MustParsePrefix("2001:db8::/32")

	tests := []struct {
		name           string
		network        *netip.Prefix
		xRealIP        string
		expectedStatus int
	}{
		{
			name:           "nil network - should allow all",
			network:        nil,
			xRealIP:        "192.168.1.1",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "allowed IPv4 - in network",
			network:        &network,
			xRealIP:        "192.168.1.100",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "forbidden IPv4 - not in network",
			network:        &network,
			xRealIP:        "10.0.0.1",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "allowed IPv6 - in network",
			network:        &ipv6Network,
			xRealIP:        "2001:db8::1",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "forbidden IPv6 - not in network",
			network:        &ipv6Network,
			xRealIP:        "2001:db9::1",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "edge case - network boundary IP",
			network:        &network,
			xRealIP:        "192.168.1.0",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "edge case - network broadcast IP",
			network:        &network,
			xRealIP:        "192.168.1.255",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid IP - should return 403",
			network:        &network,
			xRealIP:        "invalid-ip",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "empty IP - should return 403",
			network:        &network,
			xRealIP:        "",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "malformed IP - should return 403",
			network:        &network,
			xRealIP:        "192.168.1.256",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "missing X-Real-IP header",
			network:        &network,
			xRealIP:        "", // не устанавливаем заголовок
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("OK"))
			})

			middleware := FilterMiddleware(tt.network)
			wrappedHandler := middleware(handler)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.xRealIP != "" || tt.name != "missing X-Real-IP header" {
				req.Header.Set("X-Real-IP", tt.xRealIP)
			}

			rr := httptest.NewRecorder()
			wrappedHandler.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}
		})
	}
}
