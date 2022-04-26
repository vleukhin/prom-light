package storage

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"os"
	"sync"
	"time"

	"github.com/vleukhin/prom-light/internal/metrics"
)

type fileStorage struct {
	memoryStorage
	mutex       sync.Mutex
	fileName    string
	syncMode    bool
	memStorage  *memoryStorage
	storeTicker *time.Ticker
}

func NewFileStorage(fileName string, storeInterval time.Duration, restore bool) (*fileStorage, error) {
	storage := fileStorage{
		fileName:   fileName,
		memStorage: NewMemoryStorage(),
		mutex:      sync.Mutex{},
		syncMode:   true,
	}

	// Try to open file here just to check that everything is ok,
	// because if something's wrong it's better to know about that now,
	// on start up, than later when we already have collected data
	f, err := storage.openFile()
	if err != nil {
		return nil, err
	}
	if err = f.Close(); err != nil {
		return nil, err
	}

	if restore {
		if err := storage.RestoreData(); err != nil {
			return nil, err
		}
	}

	if storeInterval != 0 {
		storage.syncMode = false
		storage.storeTicker = time.NewTicker(storeInterval)
		go func() {
			for {
				<-storage.storeTicker.C
				err := storage.StoreData()
				if err != nil {
					log.Println("Failed to store data to file: " + err.Error())
				}
			}
		}()
	}

	return &storage, nil
}

func (s *fileStorage) openFile() (*os.File, error) {
	f, err := os.OpenFile(s.fileName, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Println("Failed to open file ")
	}

	return f, err
}

func (s *fileStorage) StoreData() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	f, err := s.openFile()
	if err != nil {
		return err
	}

	data, err := s.memStorage.GetAllMetrics(context.Background(), false)
	if err != nil {
		return err
	}

	err = json.NewEncoder(f).Encode(data)
	if err != nil {
		return err
	} else {
		log.Println("Data stored to file successfully")
	}

	return nil
}

func (s *fileStorage) RestoreData() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	var data []metrics.Metric

	f, err := s.openFile()
	if err != nil {
		return err
	}
	err = json.NewDecoder(f).Decode(&data)
	if err != nil && err != io.EOF {
		log.Println("Failed to restore data")
		return err
	}

	for _, m := range data {
		if m.IsCounter() {
			err := s.memStorage.IncCounter(context.Background(), m.Name, *m.Delta)
			if err != nil {
				return err
			}
		} else {
			err := s.memStorage.SetGauge(context.Background(), m.Name, *m.Value)
			if err != nil {
				return err
			}
		}
	}

	log.Println("Data restored from file successfully")

	return nil
}

func (s *fileStorage) ShutDown(_ context.Context) error {
	if err := s.StoreData(); err != nil {
		return err
	}

	if !s.syncMode {
		s.storeTicker.Stop()
	}

	return nil
}

func (s *fileStorage) SetGauge(ctx context.Context, metricName string, value metrics.Gauge) error {
	if err := s.memStorage.SetGauge(ctx, metricName, value); err != nil {
		return err
	}
	if s.syncMode {
		if err := s.StoreData(); err != nil {
			return err
		}
	}
	return nil
}

func (s *fileStorage) IncCounter(ctx context.Context, metricName string, value metrics.Counter) error {
	if err := s.memStorage.IncCounter(ctx, metricName, value); err != nil {
		return err
	}
	if s.syncMode {
		if err := s.StoreData(); err != nil {
			return err
		}
	}
	return nil
}

func (s *fileStorage) GetGauge(ctx context.Context, metricName string) (metrics.Gauge, error) {
	return s.memStorage.GetGauge(ctx, metricName)
}

func (s *fileStorage) GetCounter(ctx context.Context, metricName string) (metrics.Counter, error) {
	return s.memStorage.GetCounter(ctx, metricName)
}

func (s *fileStorage) GetAllMetrics(ctx context.Context, resetCounters bool) ([]metrics.Metric, error) {
	return s.memStorage.GetAllMetrics(ctx, resetCounters)
}

func (s *fileStorage) Ping(context.Context) error {
	f, err := s.openFile()
	if err != nil {
		return err
	}

	return f.Close()
}
