package config

import (
	"encoding/json"
	"errors"
	"net"
	"os"
	"reflect"
	"time"

	"github.com/spf13/pflag"

	"github.com/caarlos0/env/v6"
)

type Duration struct {
	time.Duration
}

const XRealIPHeader = "X-Real-IP"

const (
	ProtocolHTTP = "http"
	ProtocolGRPC = "grpc"
)

func (d *Duration) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	switch value := v.(type) {
	case float64:
		d.Duration = time.Duration(value)
		return nil
	case string:
		var err error
		d.Duration, err = time.ParseDuration(value)
		if err != nil {
			return err
		}
		return nil
	default:
		return errors.New("invalid duration")
	}
}

func fillConfigFromFile(cfg interface{}) error {
	cfile := os.Getenv("CONFIG")
	if cfile == "" {
		pflag.StringVar(&cfile, "c", "", "Path to JSON config file")
	}

	if cfile == "" {
		return nil
	}

	data, err := os.ReadFile(cfile)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, cfg)
}

func parseFuncs() map[reflect.Type]env.ParserFunc {
	return map[reflect.Type]env.ParserFunc{
		reflect.TypeOf(Duration{}): func(v string) (interface{}, error) {
			d, err := time.ParseDuration(v)
			if err != nil {
				return nil, err
			}

			return Duration{d}, nil
		},
		reflect.TypeOf(net.IPNet{}): func(v string) (interface{}, error) {
			_, subnet, err := net.ParseCIDR(v)
			if err != nil {
				return nil, err
			}
			return *subnet, nil
		},
	}
}
