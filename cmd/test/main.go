package main

import (
	"fmt"
	"github.com/BGrewell/go-iperf"
	"github.com/bgrewell/nperfmon/pkg/pinger"
	"github.com/bgrewell/nperfmon/pkg/speeder"
	"time"
)

func PrintPingResult(results pinger.PingResults) {
	// Print the results
	for _, result := range results.Results {
		fmt.Printf("Time: %v, SeqNum: %d, State: %s, Latency: %f\n", result.Timestamp.UnixMilli(), result.SeqNum, result.State, result.Latency)
	}
}

func PrintIperfResult(result *iperf.TestReport, err error) {
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf(result.String())

}

func main() {

	p := pinger.Pinger{}
	p.Target = "circuit.intel.com"
	p.Samples = 5
	p.SampleSpacingSecs = .8
	p.IntervalSecs = 5
	err := p.Start(PrintPingResult)
	if err != nil {
		panic(err)
	}

	s := speeder.NewSpeeder("127.0.0.1", 30,
		speeder.WithDuration(10),
		speeder.WithStreams(1),
	)

	err = s.Start(PrintIperfResult)
	if err != nil {
		panic(err)
	}

	time.Sleep(60 * time.Second)

	err = p.Stop()
	if err != nil {
		panic(err)
	}

	err = s.Stop()
	if err != nil {
		panic(err)
	}
}
