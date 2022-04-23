package storage

import (
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
				storage.StoreData()
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

func (s *fileStorage) StoreData() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	f, err := s.openFile()
	if err != nil {
		return
	}

	err = json.NewEncoder(f).Encode(s.memStorage.GetAllMetrics(false))
	if err != nil {
		log.Println("Failed to store data to file: " + err.Error())
	} else {
		log.Println("Data stored to file successfully")
	}
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
			s.memStorage.IncCounter(m.Name, *m.Delta)
		} else {
			s.memStorage.SetGauge(m.Name, *m.Value)
		}
	}

	log.Println("Data restored from file successfully")

	return nil
}

func (s *fileStorage) ShutDown() error {
	s.StoreData()

	if !s.syncMode {
		s.storeTicker.Stop()
	}

	return nil
}

func (s *fileStorage) SetGauge(metricName string, value metrics.Gauge) {
	s.memStorage.SetGauge(metricName, value)
	if s.syncMode {
		s.StoreData()
	}
}

func (s *fileStorage) IncCounter(metricName string, value metrics.Counter) {
	s.memStorage.IncCounter(metricName, value)
	if s.syncMode {
		s.StoreData()
	}
}

func (s *fileStorage) GetGauge(name string) (metrics.Gauge, error) {
	return s.memStorage.GetGauge(name)
}

func (s *fileStorage) GetCounter(name string) (metrics.Counter, error) {
	return s.memStorage.GetCounter(name)
}

func (s *fileStorage) GetAllMetrics(resetCounters bool) []metrics.Metric {
	return s.memStorage.GetAllMetrics(resetCounters)
}

func (s *fileStorage) Ping() error {
	f, err := s.openFile()
	if err != nil {
		return err
	}

	return f.Close()
}
