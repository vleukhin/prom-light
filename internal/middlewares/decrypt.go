package middlewares

import (
	"crypto/rsa"
	"io"
	"net/http"
	"strings"

	"github.com/rs/zerolog/log"

	"github.com/vleukhin/prom-light/internal/crypt"
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

		if err := m.decryptRequestBody(r); err != nil {
			log.Error().Err(err)
			w.WriteHeader(http.StatusInternalServerError)
		}

		next.ServeHTTP(w, r)
	})
}

func (m Decrypt) decryptRequestBody(r *http.Request) error {
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Error().Err(err).Msg("Failed to close request body")
		}
	}(r.Body)
	original, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}

	decrypted, err := crypt.DecryptOAEP(m.key, original, nil)
	if err != nil {
		return err
	}

	r.Body = io.NopCloser(strings.NewReader(string(decrypted)))
	r.ContentLength = int64(len(decrypted))

	return nil
}
