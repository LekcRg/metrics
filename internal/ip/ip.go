package ip

import (
	"net"
	"net/http"
	"net/netip"

	"github.com/LekcRg/metrics/internal/logger"
	"go.uber.org/zap"
)

func GetOutboundIP() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String(), nil
}

func FilterMiddleware(network *netip.Prefix) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if network == nil {
				next.ServeHTTP(w, r)
				return
			}

			ipStr := r.Header.Get("X-Real-IP")
			ip, err := netip.ParseAddr(ipStr)
			if err != nil {
				logger.Log.Error("parse ip err", zap.Error(err))
				http.Error(w, "403 Forbidden", http.StatusForbidden)
				return
			}

			isContain := network.Contains(ip)
			if !isContain {
				http.Error(w, "403 Forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
