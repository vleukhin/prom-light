package middlewares

import (
	"crypto/rsa"
	"net/http"
)

type Decrypt struct {
	key *rsa.PrivateKey
}

func NewDecryptMiddleware(key *rsa.PrivateKey) Decrypt {
	return Decrypt{
		key: key,
	}
}

func (m Decrypt) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if m.key == nil {
			next.ServeHTTP(w, r)
			return
		}

		// decrypt
		next.ServeHTTP(w, r)
	})
}
