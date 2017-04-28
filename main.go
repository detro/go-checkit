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
  durationPtr := flag.Float64("duration", DEFAULT_DURATION_MIN, "How long to run the check for (in minutes)")
  frequencyPtr := flag.Int64("frequency", DEFAULT_FREQUENCY_MS, "Frequency at which to run the checks (in milliseconds")
  flag.Parse()

  // Confirm given setup
  log.Printf("Checking URL '%s'\n", *urlPtr);
  log.Printf("Running for %.1f minute(s) at a %d millisecond(s) frequency\n", *durationPtr, *frequencyPtr)
  log.Println()

  // NOTE: As per documentation of https://golang.org/pkg/net/http/#Transport, Transport should be re-used as needed.
  timedTransport := transport.NewTimedTransport()
  client := &http.Client{Transport: timedTransport}

  // Initialize histograms for the durations we are going to measure (grossly approximating 1 check per second, multiplied by 10)
  approxChecks := int(60 * 10 * *durationPtr)
  totalHG := metrics.NewHistogram(metrics.NewUniformSample(approxChecks))
  connectDurationHG := metrics.NewHistogram(metrics.NewUniformSample(approxChecks))
  sendWaitReceiveDurationHG := metrics.NewHistogram(metrics.NewUniformSample(approxChecks))
  roundTripDurationHG := metrics.NewHistogram(metrics.NewUniformSample(approxChecks))

  // Start the checks
  startTime := time.Now()
  log.Println("Beginning checks at", startTime)
  for time.Since(startTime).Minutes() < *durationPtr {
    fmt.Print(".")

    // Perform and time the call to "client.Get"s
    totalPerformGetDuration := performGetTimed(client, *urlPtr)

    // Update all the histograms
    totalHG.Update(durationToMs(totalPerformGetDuration))
    connectDurationHG.Update(durationToMs(timedTransport.ConnectDuration()))
    sendWaitReceiveDurationHG.Update(durationToMs(timedTransport.SendWaitReceiveDuration()))
    roundTripDurationHG.Update(durationToMs(timedTransport.RoundTripDuration()))

    // Sleeping current go-routine for the given "frequency"
    // NOTE: We are expressing "frequency" as a duration but it should really be a count of how
    // many checks we want to do in a given time interval.
    // Our approach is less "scientifically exact" but does the job for now.
    time.Sleep(msToDuration(*frequencyPtr))
  }
  fmt.Println()
  log.Println("Ending checks at", time.Now())

  // Print-out results nicely
  log.Println()
  log.Println("*** RESULTS (in milliseconds) ***")
  log.Printf("Performed %d checks", totalHG.Count())
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
  log.Printf("%s \t Min %d \t Mean %.2f \t Max %d \t P75 %.2f \t P99 %.2f (ms)\n", histogramName, histogram.Min(), histogram.Mean(), histogram.Max(), histogram.Percentile(0.75), histogram.Percentile(0.99))
}

func durationToMs(duration time.Duration) int64 {
  return duration.Nanoseconds() / 1000000
}

func msToDuration(ms int64) time.Duration {
  return time.Duration(ms * 1000000)
}