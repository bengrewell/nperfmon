package speeder

import (
	"github.com/BGrewell/go-iperf"
	"strings"
	"time"
)

type Option func(*Speeder)

func WithStreams(streams int) Option {
	return func(s *Speeder) {
		s.streams = streams
	}
}

func WithDuration(duration int) Option {
	return func(s *Speeder) {
		s.duration = duration
	}
}

func WithReportInterval(reportInterval int) Option {
	return func(s *Speeder) {
		s.reportInterval = reportInterval
	}
}

func WithIncludeServer(includeServer bool) Option {
	return func(s *Speeder) {
		s.includeServer = includeServer
	}
}

func WithJSON(json bool) Option {
	return func(s *Speeder) {
		s.json = json
	}
}

func WithProtocol(protocol string) Option {
	return func(s *Speeder) {
		s.protocol = protocol
	}
}

func WithReverse(reverse bool) Option {
	return func(s *Speeder) {
		s.reverse = reverse
	}
}

//func WithBidirectional(bidirectional bool) Option {
//	return func(s *Speeder) {
//		s.bidirectional = bidirectional
//	}
//}

func NewSpeeder(target string, testInterval int, options ...Option) *Speeder {
	s := &Speeder{
		Target:         target,
		IntervalSecs:   testInterval,
		protocol:       "tcp",
		streams:        1,
		duration:       10,
		reportInterval: 1,
		reverse:        false,
		//bidirectional:  false,
		includeServer: false,
		json:          true,
	}

	for _, option := range options {
		option(s)
	}

	return s
}

type Speeder struct {
	// Target is the iperf server to test against
	Target string `json:"target" yaml:"target" xml:"target"`
	// IntervalSecs is the interval between tests
	IntervalSecs int `json:"interval_secs" yaml:"interval_secs" xml:"interval_secs"`
	protocol     string
	streams      int
	duration     int
	reverse      bool
	//bidirectional  bool
	reportInterval int
	includeServer  bool
	json           bool
	running        bool
	client         *iperf.Client
	callback       func(report *iperf.TestReport, err error)
}

func (s *Speeder) Start(callback func(report *iperf.TestReport, err error)) error {

	s.callback = callback
	proto := iperf.PROTO_TCP
	if strings.ToLower(s.protocol) == "udp" {
		proto = iperf.PROTO_UDP
	}

	s.client = iperf.NewClient(s.Target)
	s.client.SetProto(iperf.Protocol(proto))
	s.client.SetStreams(s.streams)
	s.client.SetTimeSec(s.duration)
	s.client.SetReverse(s.reverse)
	s.client.SetInterval(s.reportInterval)
	s.client.SetIncludeServer(s.includeServer)
	s.client.SetJSON(s.json)

	s.running = true
	go s.run()

	return nil
}

func (s *Speeder) run() {
	for s.running {
		nextStartTime := time.Now().Add(time.Duration(s.IntervalSecs) * time.Second)
		err := s.client.Start()

		<-s.client.Done

		if err != nil {
			if s.callback != nil {
				s.callback(nil, err)
			}
		} else {
			s.callback(s.client.Report(), nil)
		}
		// Sleep until the next test
		time.Sleep(nextStartTime.Sub(time.Now()))
	}
}

func (s *Speeder) Stop() error {
	s.running = false

	return nil
}
