package sumologic

import (
	"fmt"
	"out_sumologic/pkg/fluentbit"
	"time"
	"unsafe"

	"github.com/fluent/fluent-bit-go/output"
	"github.com/sirupsen/logrus"
)

const (
	PLUGIN_NAME        string = "sumologic"
	PLUGIN_DESCRIPTION string = "Sends logs to Sumo Logic"
)

type Entry struct {
	record    string
	timestamp time.Time
}

type Batch struct {
	entries []*Entry
	config  *sumoLogicConfig
}

type SumoLogic struct {
	config   *sumoLogicConfig
	logger   *logrus.Entry
	uploader *uploader
}

func (s *SumoLogic) CreateBatch(data unsafe.Pointer, length int, tag string) (*Batch, error) {
	var err error
	batch := new(Batch)
	dec := output.NewDecoder(data, length)

	for {
		// Extract record
		ret, ts, record := output.GetRecord(dec)
		if ret != 0 {
			break
		}
		// Parse timestamp
		timestamp := fluentbit.GetTimeStamp(ts)
		// Convert record to JSON
		jsonRecord, err := fluentbit.CreateJSON(record, s.config.logKey)
		if err != nil {
			s.logger.Warnf("failed to parse the record %v", err)
			return nil, err
		}
		batch.entries = append(batch.entries, &Entry{
			record:    jsonRecord,
			timestamp: timestamp,
		})
	}

	batch.config, err = parseConfig(s.config, tag)
	if err != nil {
		s.logger.Warnf("failed to parse the config %v", err)
		return nil, err
	}

	return batch, nil
}

func (s *SumoLogic) SendBatch(batch *Batch) error {
	// queue batch for upload
	select {
	case s.uploader.batches <- batch:
		s.logger.Trace("batch queued successfully")
		return nil
	default:
		s.logger.Warnf("failed to queue the batch")
		return fmt.Errorf("channel closed, unable to queue the batch")
	}
}

func (s *SumoLogic) Start() {
	defer s.uploader.wg.Done()
	for batch := range s.uploader.batches {
		// upload batch to sumologic
		s.logger.Debugf("attempting upload to sumologic with config: %v", batch.config)
		s.uploader.wg.Add(1)
		go s.retryAndUpload(batch)
	}
}

func (s *SumoLogic) retryAndUpload(batch *Batch) {
	defer s.uploader.wg.Done()
	err := s.retryer(
		batch.config.maxRetries,
		func() error {
			return s.uploader.upload(batch)
		},
	)
	if err != nil {
		s.logger.Errorf("failed to upload batch to sumologic %v, retries exhausted!", err)
	} else {
		s.logger.Debug("successfully uploaded batch to sumologic")
	}
}

func (s *SumoLogic) retryer(attempts uint64, f func() error) error {
	err := f()
	if err != nil && attempts > 0 {
		s.logger.Debugf("failed with err: %v, %d retries left", err, attempts)
		time.Sleep(5 * time.Second)
		return s.retryer(attempts-1, f)
	} else {
		return err
	}
}

func (s *SumoLogic) Stop() {
	s.logger.Info("exiting gracefully...")
	close(s.uploader.batches)
	s.uploader.wg.Wait()
	s.logger.Info("exited")
}

func Initalize(plugin unsafe.Pointer, id int) (*SumoLogic, error) {
	var err error
	s := new(SumoLogic)

	s.config, err = loadConfig(plugin)
	if err != nil {
		return nil, fmt.Errorf("unable to load the plugin config\n %w", err)
	}

	s.logger = fluentbit.GetLogger(
		fluentbit.GetOutputInstanceName(PLUGIN_NAME, id),
		s.config.level,
	)

	s.logger.Info("initializing...")
	s.logger.Debug(s.config)

	s.uploader = &uploader{
		batches: make(chan *Batch, 1000),
	}
	s.uploader.wg.Add(1)
	go s.Start()

	return s, nil
}
