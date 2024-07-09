package wrapper

import (
	"fmt"
	"github.com/BGrewell/go-iperf"
	"github.com/bgrewell/nperfmon/pkg/pinger"
	"github.com/bgrewell/nperfmon/pkg/speeder"
)

func NewWrapper(target string, pingInterval int, speedInterval int) (wrapper *Wrapper, err error) {

	w := &Wrapper{}

	p := pinger.Pinger{}
	p.Target = target
	p.Samples = 3
	p.SampleSpacingSecs = .1
	p.IntervalSecs = float64(pingInterval)

	w.pinger = &p

	s := speeder.NewSpeeder(target, speedInterval,
		speeder.WithDuration(15),
		speeder.WithStreams(1),
	)

	w.speeder = s

	return w, nil
}

type Wrapper struct {
	pinger  *pinger.Pinger
	speeder *speeder.Speeder
}

func (w *Wrapper) Start() error {
	err := w.pinger.Start(w.ProcessPingResult)
	if err != nil {
		return err
	}

	//err = w.speeder.Start(w.ProcessIperfResult)
	if err != nil {
		return err
	}

	return nil
}

func (w *Wrapper) Stop() error {
	err := w.pinger.Stop()
	if err != nil {
		return err
	}

	err = w.speeder.Stop()
	if err != nil {
		return err
	}

	return nil
}

func (w *Wrapper) ProcessPingResult(results pinger.PingResults) {
	// Print the results
	for _, result := range results.Results {
		fmt.Printf("Time: %v, SeqNum: %d, State: %s, Latency: %f\n", result.Timestamp.UnixMilli(), result.SeqNum, result.State, result.Latency)
	}
}

func (w *Wrapper) ProcessIperfResult(result *iperf.TestReport, err error) {
	// Print the results
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf(result.String())
}
