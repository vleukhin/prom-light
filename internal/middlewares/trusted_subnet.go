package middlewares

import (
	"log"
	"net"
	"net/http"

	"github.com/vleukhin/prom-light/internal/config"
)

type TrustedIPs struct {
	CIDR net.IPNet
}

func NewTrustedIPsMiddleware(CIDR net.IPNet) TrustedIPs {
	return TrustedIPs{
		CIDR: CIDR,
	}
}

func (m TrustedIPs) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		IPRaw := r.Header.Get(config.XRealIPHeader)
		if IPRaw == "" {
			IPRaw = r.RemoteAddr
		}
		log.Println(IPRaw)
		if !m.CIDR.Contains(net.ParseIP(IPRaw)) {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
