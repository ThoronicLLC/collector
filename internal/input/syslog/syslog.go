package syslog

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ThoronicLLC/collector/pkg/core"
	log "github.com/sirupsen/logrus"
	"gopkg.in/mcuadros/go-syslog.v2"
	"sync"
	"time"
)

var InputName = "syslog"

type Config struct {
	Address        string `json:"address"`
	Port           int    `json:"port"`
	Protocol       string `json:"protocol"`
	Format         string `json:"format"`
	FlushFrequency int    `json:"flush_frequency"`
}

type syslogInput struct {
	Config     []byte
	ctx        context.Context
	cancelFunc context.CancelFunc
	server     *syslog.Server
	logChannel syslog.LogPartsChannel
}

func Handler() core.InputHandler {
	return func(config []byte) core.Input {
		// Setup context
		ctx, cancelFn := context.WithCancel(context.Background())

		return &syslogInput{
			Config:     config,
			ctx:        ctx,
			cancelFunc: cancelFn,
			server:     syslog.NewServer(),
			logChannel: make(syslog.LogPartsChannel),
		}
	}
}

func (s *syslogInput) Run(errorHandler core.ErrorHandler, state core.State, processPipe chan<- core.PipelineResults) {
	// Setup default config
	conf := Config{
		Address:        "0.0.0.0",
		Port:           1514,
		Protocol:       "udp",
		Format:         "raw",
		FlushFrequency: 500,
	}

	// Marshal settings into config
	err := json.Unmarshal(s.Config, &conf)
	if err != nil {
		errorHandler(true, fmt.Errorf("issue unmarshalling config: %s", err))
		return
	}

	// Setup local variables
	tmpWriter, err := core.NewTmpWriter()
	if err != nil {
		errorHandler(true, err)
		return
	}

	// Setup log channel and handler
	handler := syslog.NewChannelHandler(s.logChannel)

	// Set syslog format and handler
	switch conf.Format {
	case "automatic":
		s.server.SetFormat(syslog.Automatic)
	case "RFC3164":
		s.server.SetFormat(syslog.RFC3164)
	case "RFC5424":
		s.server.SetFormat(syslog.RFC5424)
	case "RFC6587":
		s.server.SetFormat(syslog.RFC6587)
	case "raw":
		s.server.SetFormat(noFormat)
	default:
		s.server.SetFormat(noFormat)
	}
	s.server.SetHandler(handler)

	addressAndPort := fmt.Sprintf("%s:%d", conf.Address, conf.Port)

	// Setup TCP listener
	if conf.Protocol == "tcp" || conf.Protocol == "both" {
		log.Debugf("syslog server listening on %s/%s", addressAndPort, "TCP")
		if err := s.server.ListenTCP(addressAndPort); err != nil {
			errorHandler(true, fmt.Errorf("unable to start TCP listener on %s", addressAndPort))
			return
		}
	}

	// Setup UDP listener
	if conf.Protocol == "udp" || conf.Protocol == "both" {
		log.Debugf("syslog server listening on %s/%s", addressAndPort, "UDP")
		if err := s.server.ListenUDP(addressAndPort); err != nil {
			errorHandler(true, fmt.Errorf("unable to start UDP listener on %s", addressAndPort))
			return
		}
	}

	// Boot up server
	if err = s.server.Boot(); err != nil {
		errorHandler(true, fmt.Errorf("unable to boot syslog service: %v", err))
		return
	}

	// Setup wait group
	var wg sync.WaitGroup

	// Start timed process sync go routine
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-s.ctx.Done():
				err := s.flush(tmpWriter, processPipe)
				if err != nil {
					errorHandler(false, err)
				}
				return
			case <-time.After(time.Duration(conf.FlushFrequency) * time.Second):
				err := s.flush(tmpWriter, processPipe)
				if err != nil {
					errorHandler(true, err)
					s.Stop()
					continue
				}
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-s.ctx.Done():
				return
			case logParts, ok := <-s.logChannel:
				if !ok {
					return
				}

				// Get data from content of message
				if contentVal, contentExists := logParts["content"]; contentExists {
					if stringContentVal, ok := contentVal.(string); ok {
						_, err := tmpWriter.Write([]byte(stringContentVal))
						if err != nil {
							errorHandler(false, fmt.Errorf("issue writing log: %s", err))
						}
					}
				} else if messageVal, messageExists := logParts["message"]; messageExists {
					if stringMessageVal, ok := messageVal.(string); ok {
						_, err := tmpWriter.Write([]byte(stringMessageVal))
						if err != nil {
							errorHandler(false, fmt.Errorf("issue writing log: %s", err))
						}
					}
				}
			}
		}
	}()

	wg.Wait()
}

func (s *syslogInput) Stop() {
	s.cancelFunc()
	_ = s.server.Kill()
	close(s.logChannel)
}

func (s *syslogInput) flush(writer *core.TmpWriter, processPipe chan<- core.PipelineResults) error {
	// Rotate the temp writer
	count, fileName, rErr := writer.Rotate()
	if rErr != nil {
		return fmt.Errorf("issue rotating temp file: %s", rErr)
	}

	// Only send on if there are results
	if count > 0 {
		processPipe <- core.PipelineResults{
			FilePath:    fileName,
			ResultCount: count,
			State:       nil,
			RetryCount:  0,
		}
	}

	return nil
}
