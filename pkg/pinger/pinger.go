package pinger

import (
	"fmt"
	"github.com/bgrewell/nperfmon/pkg/utils"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"net"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

type PingState string

const (
	PingStateOK   PingState = "ok"
	PingStateLost PingState = "lost"
	PingStateErr  PingState = "error"
)

type PingResult struct {
	Timestamp time.Time `json:"timestamp" yaml:"timestamp" xml:"timestamp"`
	SeqNum    int       `json:"seq_num" yaml:"seq_num" xml:"seq_num"`
	State     PingState `json:"state" yaml:"state" xml:"state"`
	Latency   float64   `json:"latency" yaml:"latency" xml:"latency"`
	Error     error     `json:"error,omitempty" yaml:"error,omitempty" xml:"error"`
}

type PingResults struct {
	Timestamp time.Time    `json:"timestamp" yaml:"timestamp" xml:"timestamp"`
	Results   []PingResult `json:"results" yaml:"results" xml:"results"`
}

// Pinger is a struct which encapsulates the latency measurement functions.
type Pinger struct {
	Target            string  `json:"target" yaml:"target" xml:"target"`
	Protocol          string  `json:"protocol" yaml:"protocol" xml:"protocol"`
	IntervalSecs      float64 `json:"interval_secs" yaml:"interval_secs" xml:"interval_secs"`
	Samples           int     `json:"samples" yaml:"samples" xml:"samples"`
	SampleSpacingSecs float64 `json:"sample_spacing_secs" yaml:"sample_spacing_secs" xml:"sample_spacing_secs"`
	WindowSecs        float64 `json:"window_secs" yaml:"window_secs" xml:"window_secs"`
	resultsBuffer     utils.CircularBuffer[PingResults]
	addr              *net.IPAddr
	conn              *net.IPConn
	seq               atomic.Int32
	callback          func(PingResults)
	running           bool
}

func (p *Pinger) Start(callback func(PingResults)) (err error) {
	p.callback = callback
	p.resultsBuffer = *utils.NewCircularBuffer[PingResults](int(p.WindowSecs/p.IntervalSecs) + 1)

	p.addr, err = net.ResolveIPAddr("ip4", p.Target)
	if err != nil {
		return fmt.Errorf("Failed to resolve address: %v\n", err)
	}

	p.conn, err = net.DialIP("ip4:icmp", nil, p.addr)
	if err != nil {
		return fmt.Errorf("Failed to dial: %v\n", err)
	}

	p.running = true
	go p.run()

	return nil
}

func (p *Pinger) run() {
	for p.running {
		start := time.Now()
		results := make([]PingResult, p.Samples)
		resultsChan := make(chan PingResult, p.Samples)
		wg := &sync.WaitGroup{}

		// Start sending pings
		for i := 0; i < p.Samples; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				result, _ := p.sendPing()
				resultsChan <- *result
			}(i)
			time.Sleep(time.Duration(p.SampleSpacingSecs * float64(time.Second)))
		}

		// Collect results
		go func() {
			wg.Wait()
			close(resultsChan)
		}()

		for result := range resultsChan {
			results[(result.SeqNum-1)%p.Samples] = result
		}

		prs := PingResults{
			Timestamp: start,
			Results:   results,
		}

		if p.callback != nil {
			p.callback(prs)
		}

		time.Sleep(time.Duration(p.IntervalSecs) * time.Second)
	}
}

func (p *Pinger) Stop() error {
	p.running = false
	return p.conn.Close()
}

func (p *Pinger) sendPing() (*PingResult, error) {
	seq := int(p.seq.Add(1))

	icmpMessage := icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: &icmp.Echo{
			ID:   os.Getpid() & 0xffff,
			Seq:  seq,
			Data: []byte("PING"),
		},
	}

	messageBytes, err := icmpMessage.Marshal(nil)
	if err != nil {
		ferr := fmt.Errorf("Failed to marshal ICMP message: %v\n", err)
		return &PingResult{
			SeqNum:  seq,
			State:   PingStateErr,
			Latency: 0,
			Error:   ferr,
		}, ferr
	}

	startTime := time.Now()
	if _, err := p.conn.Write(messageBytes); err != nil {
		ferr := fmt.Errorf("Failed to write ICMP message: %v\n", err)
		return &PingResult{
			Timestamp: startTime,
			SeqNum:    seq,
			State:     PingStateErr,
			Latency:   0,
			Error:     ferr,
		}, ferr
	}

	receiveBuffer := make([]byte, 1500)
	if err := p.conn.SetReadDeadline(time.Now().Add(3 * time.Second)); err != nil {
		ferr := fmt.Errorf("Failed to set read deadline: %v\n", err)
		return &PingResult{
			Timestamp: startTime,
			SeqNum:    seq,
			State:     PingStateErr,
			Latency:   0,
			Error:     ferr,
		}, ferr
	}

	n, _, err := p.conn.ReadFrom(receiveBuffer)
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return &PingResult{
				Timestamp: startTime,
				SeqNum:    seq,
				State:     PingStateLost,
				Latency:   0,
				Error:     nil,
			}, nil
		} else {
			ferr := fmt.Errorf("Failed to read from connection: %v\n", err)
			return &PingResult{
				Timestamp: startTime,
				SeqNum:    seq,
				State:     PingStateErr,
				Latency:   0,
				Error:     ferr,
			}, ferr
		}
	}
	endTime := time.Now()

	receivedMessage, err := icmp.ParseMessage(ipv4.ICMPTypeEcho.Protocol(), receiveBuffer[:n])
	if err != nil {
		ferr := fmt.Errorf("Failed to parse ICMP message: %v\n", err)
		return &PingResult{
			Timestamp: startTime,
			SeqNum:    seq,
			State:     PingStateErr,
			Latency:   0,
			Error:     ferr,
		}, ferr
	}

	switch receivedMessage.Type {
	case ipv4.ICMPTypeEchoReply:
		return &PingResult{
			Timestamp: startTime,
			SeqNum:    receivedMessage.Body.(*icmp.Echo).Seq,
			State:     PingStateOK,
			Latency:   endTime.Sub(startTime).Seconds(),
		}, nil
	default:
		ferr := fmt.Errorf("Received unexpected ICMP message type: %v\n", receivedMessage.Type)
		return &PingResult{
			Timestamp: startTime,
			SeqNum:    seq,
			State:     PingStateErr,
			Latency:   0,
			Error:     ferr,
		}, ferr
	}
}
