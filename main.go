package main

// Golang library dependencies
import (
  "net/http"
  "flag"
  "log"
  "time"
  "fmt"
)

// Third party dependencies
import (
  metrics "github.com/rcrowley/go-metrics"
)

// Internal dependencies
import (
  "checkit/transport"
)

const DEFAULT_DURATION_MIN = 5
const DEFAULT_URL = "https://gitlab.com"
const DEFAULT_FREQUENCY_MS = 500

func main() {
  // Parse command line arguments/flags
  urlPtr := flag.String("url", DEFAULT_URL, "URL to check")
  durationPtr := flag.Int("duration", DEFAULT_DURATION_MIN, "How long to run the check for (in minutes)")
  frequencyPtr := flag.Int("frequency", DEFAULT_FREQUENCY_MS, "Frequency at which to run the checks (in milliseconds")
  flag.Parse()

  // Confirm given setup
  log.Printf("* Checking URL '%s'\n", *urlPtr);
  log.Printf("* Running for %d minute(s) at a %d millisecond(s) frequency\n", *durationPtr, *frequencyPtr)
  log.Println()

  // NOTE: As per documentation of https://golang.org/pkg/net/http/#Transport, Transport should be re-used as needed.
  timedTransport := transport.NewTimedTransport()
  client := &http.Client{Transport: timedTransport}

  // Initialize histograms for the durations we are going to measure
  totalHG := metrics.NewHistogram(metrics.NewUniformSample(10000))
  connectDurationHG := metrics.NewHistogram(metrics.NewUniformSample(10000))
  sendWaitReceiveDurationHG := metrics.NewHistogram(metrics.NewUniformSample(10000))
  roundTripDurationHG := metrics.NewHistogram(metrics.NewUniformSample(10000))

  for i := 0; i < DEFAULT_DURATION_MIN; i++ {
    fmt.Print(".")

    totalPerformGetDuration := performGetTimed(client, *urlPtr)

    totalHG.Update(durationToMs(totalPerformGetDuration))
    connectDurationHG.Update(durationToMs(timedTransport.ConnectDuration()))
    sendWaitReceiveDurationHG.Update(durationToMs(timedTransport.SendWaitReceiveDuration()))
    roundTripDurationHG.Update(durationToMs(timedTransport.RoundTripDuration()))
  }

  fmt.Print("\n")
  log.Println("*** RESULTS (in milliseconds) ***")
  printlnHistogram("Establish connection ", connectDurationHG)
  printlnHistogram("Send, wait, receive  ", sendWaitReceiveDurationHG)
  printlnHistogram("Round trip           ", roundTripDurationHG)
  printlnHistogram("HTTP GET (i.e. total)", totalHG)
}

func performGetTimed(client *http.Client, url string) time.Duration {
  start := time.Now()

  resp, err := client.Get(url)
  if err != nil {
    log.Fatalf("Unable to perform HTTP GET against %s: %s", url, err)
  }
  defer resp.Body.Close()

  return time.Now().Sub(start)
}

func printlnHistogram(histogramName string, histogram metrics.Histogram) {
  log.Printf("* %s \t Min %d \t Mean %.2f \t Max %d \t P75 %.2f \t P99 %.2f (ms)\n", histogramName, histogram.Min(), histogram.Mean(), histogram.Max(), histogram.Percentile(0.75), histogram.Percentile(0.99))
}

func durationToMs(duration time.Duration) int64 {
  return duration.Nanoseconds() / 1000000
}