package sumologic

import (
	"fmt"
	"out_sumologic/pkg/fluentbit"
	"out_sumologic/pkg/fluentbit/logger"
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
	log      *logrus.Entry
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
			s.log.Warnf("failed to parse the record %v", err)
			return nil, err
		}
		batch.entries = append(batch.entries, &Entry{
			record:    jsonRecord,
			timestamp: timestamp,
		})
	}

	batch.config, err = parseConfig(s.config, tag)
	if err != nil {
		s.log.Warnf("failed to parse the config %v", err)
		return nil, err
	}

	return batch, nil
}

func (s *SumoLogic) SendBatch(batch *Batch) error {
	// queue batch for upload
	select {
	case s.uploader.batches <- batch:
		s.log.Trace("batch queued successfully")
		return nil
	default:
		s.log.Warnf("failed to queue the batch")
		return fmt.Errorf("channel closed, unable to queue the batch")
	}
}

func (s *SumoLogic) Start() {
	defer s.uploader.wg.Done()
	for batch := range s.uploader.batches {
		// upload batch to sumologic
		s.log.Debugf("attempting upload to sumologic with config: %v", batch.config)
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
		s.log.Errorf("failed to upload batch to sumologic %v, retries exhausted!", err)
	} else {
		s.log.Debug("successfully uploaded batch to sumologic")
	}
}

func (s *SumoLogic) retryer(attempts uint64, f func() error) error {
	err := f()
	if err != nil && attempts > 0 {
		s.log.Debugf("failed with err: %v, %d retries left", err, attempts)
		time.Sleep(5 * time.Second)
		return s.retryer(attempts-1, f)
	} else {
		return err
	}
}

func (s *SumoLogic) Stop() {
	s.log.Info("exiting gracefully...")
	close(s.uploader.batches)
	s.uploader.wg.Wait()
	s.log.Info("exited")
}

func Initalize(plugin unsafe.Pointer, id int) (*SumoLogic, error) {
	var err error
	s := new(SumoLogic)

	s.config, err = loadConfig(plugin)
	if err != nil {
		return nil, fmt.Errorf("unable to load the plugin config\n %w", err)
	}

	s.log = logger.New(
		fluentbit.GetOutputInstanceName(PLUGIN_NAME, id),
		s.config.level,
	)

	s.log.Info("initializing...")
	s.log.Debug(s.config)

	s.uploader = &uploader{
		batches: make(chan *Batch, 1000),
	}
	s.uploader.wg.Add(1)
	go s.Start()

	return s, nil
}
