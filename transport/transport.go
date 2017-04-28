package transport

import (
  "net"
  "net/http"
  "time"
)

// A Transport that provides instrumented RoundTripper and Dialer
// to time various steps of communication in an HTTP request
type TimedTransport struct {
  roundTripper http.RoundTripper
  dialer       *net.Dialer

  dialStart      time.Time
  dialEnd        time.Time
  roundTripStart time.Time
  roundTripEnd   time.Time
}

func NewTimedTransport() *TimedTransport {
  newTT := &TimedTransport{ }

  newTT.dialer = &net.Dialer{
    Timeout:   2 * time.Minute, //< seams reasonable: hanging forever wouldn't help
    KeepAlive: 0,               //< we are interested in timing/performance: no connection reuse
    DualStack: true,            //< support both IPv4 and IPv6 addresses
  }

  newTT.roundTripper = &http.Transport{
    Proxy:               http.ProxyFromEnvironment,
    Dial:                newTT.Dial,
    TLSHandshakeTimeout: 30 * time.Second,  //< seams reasonable for this scenario
    DisableKeepAlives: true,                //< again, we are interested in timing/performance: no connection reuse
  }

  return newTT
}

func (newTT *TimedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
  newTT.roundTripStart = time.Now()

  res, err := newTT.roundTripper.RoundTrip(req)

  newTT.roundTripEnd = time.Now()
  return res, err
}

func (newTT *TimedTransport) Dial(network, address string) (net.Conn, error) {
  newTT.dialStart = time.Now()

  con, err := newTT.dialer.Dial(network, address)

  newTT.dialEnd = time.Now()
  return con, err
}

// NOTE: Sequence of events will be:
//
//   roundTripStart -> dialStart -> [resolve address] -> dialEnd -> [send request] -> [receive response] -> roundTripEnd
//
// Current implementation instruments only the "dialing" and the "roundtrip": with further overriding it would
// be possible to capture/hook-into more key points in the timeline, to provide more accurate and complete timing.
func (newTT *TimedTransport) ConnectDuration() time.Duration {
  return newTT.dialEnd.Sub(newTT.dialStart)
}

func (newTT *TimedTransport) SendWaitReceiveDuration() time.Duration {
  return newTT.roundTripEnd.Sub(newTT.dialEnd)
}

func (newTT *TimedTransport) RoundTripDuration() time.Duration {
  return newTT.roundTripEnd.Sub(newTT.roundTripStart)
}
