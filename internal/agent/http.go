package agent

import (
	"bytes"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/vleukhin/prom-light/internal/crypt"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/vleukhin/prom-light/internal/config"
	"github.com/vleukhin/prom-light/internal/metrics"
)

type httpClient struct {
	serverAddr string
	client     http.Client
	IP         net.IP
	key        *rsa.PublicKey
}

func NewHTTPClient(serverAddr string, IP net.IP, timeout time.Duration, key *rsa.PublicKey) *httpClient {
	client := http.Client{}
	client.Timeout = timeout
	return &httpClient{
		serverAddr: serverAddr,
		client:     client,
		IP:         IP,
		key:        key,
	}
}

// SendMetricToServer отправляет запрос на сервер метрик
func (c *httpClient) SendMetricToServer(m metrics.Metric) error {
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}

	return c.sendRequest("/update", data)
}

// SendBatchMetricsToServer отправляет batch запрос на сервер метрик
func (c *httpClient) SendBatchMetricsToServer(m metrics.Metrics) error {
	data, err := c.encrypt(m)
	if err != nil {
		return err
	}

	return c.sendRequest("/updates", data)
}

// sendRequest отправляет запрос на сервер метрик
func (c *httpClient) sendRequest(endpoint string, data []byte) error {
	r, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://%s%s/", c.serverAddr, endpoint), bytes.NewBuffer(data))
	r.Header.Set(config.XRealIPHeader, c.IP.String())
	if err != nil {
		return err
	}
	resp, err := c.client.Do(r)
	if err != nil {
		return err
	}
	err = resp.Body.Close()
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return errors.New("bad response while batch reporting: " + strconv.Itoa(resp.StatusCode))
	}

	return nil
}

// encrypt encrypts metrics with public key
func (c *httpClient) encrypt(m metrics.Metrics) ([]byte, error) {
	data, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	if c.key == nil {
		return data, nil
	}

	return crypt.EncryptOAEP(c.key, data, nil)
}
