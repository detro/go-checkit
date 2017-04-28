# (Go) check it

`checkit` is a command line tool (written in Go, obviously), that runs a sequence of HTTP GET against a given
HTTP(S) endpoint, for a configurable duration and frequency. It supports HTTP 1 and 2, over IPv4 and IPv6,
because Go `net/http` does.

## How does it "check"

For every HTTP GET we _currently_ measure:
- time to establish the connection
- time to send request, wait, receive the response
- round trip time from the moment the request is submitted to the moment it comes back
- the "total" HTTP GET execution as seen by the application

**NOTE:** The word _currently_ above stands to signify that we could evolve further to instrument the `net/http` library further,
and gather more precise metrics (like _DNS resolution time_ or separating _send,wait,receive_ times). Those are
currently left as TODOs.

### Connection reuse?

All HTTP connection reuse is disabled. On purpose.
The purpose of this tool is not to optimize for end speed: is to collect accurate data. And so connection reuse would
definite skew the results if left in place.

For details, take a look at the `KeepAlive` options in `transport.go` module.

## How do we measure

We track the time measurements using the Histogram implementation by the [go-metrics](https://github.com/rcrowley/go-metrics)
project. This allows to produce as final result:

* Min, Mean and Max
* 75 and 99 Percentiles

Allowing to spot outliers and ensure that the final numbers are not skewed in any way.

## Usage

Just use `-h` flag to print it out:

```
$ checkit -h
Usage of ./bin/checkit:
  -duration float
    	How long to run the check for (in minutes) (default 5)
  -frequency int
    	Frequency at which to run the checks (in milliseconds (default 500)
  -h	Print usage
  -url string
    	URL to check (default "https://gitlab.com")
```

## Example output

Launch against the default URL, run for 1 minute at a 500 milliseconds frequency:

```
$ ./bin/checkit -duration=1 -frequency=500
2017/04/28 12:18:28 Checking URL 'https://gitlab.com'
2017/04/28 12:18:28 Running for 1.0 minute(s) at a 500 millisecond(s) frequency
2017/04/28 12:18:28 
2017/04/28 12:18:28 Beginning checks at 2017-04-28 12:18:28.511723766 +0100 BST
..........................................
2017/04/28 12:19:29 Ending checks at 2017-04-28 12:19:29.662597002 +0100 BST
2017/04/28 12:19:29 
2017/04/28 12:19:29 *** RESULTS (in milliseconds) ***
2017/04/28 12:19:29 Performed 42 checks
2017/04/28 12:19:29 Establish connection  	 Min 88 	 Mean 90.05 	 Max 103 	 P75 91.00 	 P99 103.00 (ms)
2017/04/28 12:19:29 Send, wait, receive   	 Min 444 	 Mean 451.74 	 Max 516 	 P75 452.00 	 P99 516.00 (ms)
2017/04/28 12:19:29 Round trip            	 Min 533 	 Mean 542.38 	 Max 605 	 P75 544.00 	 P99 605.00 (ms)
2017/04/28 12:19:29 HTTP GET (i.e. total) 	 Min 921 	 Mean 952.00 	 Max 1123 	 P75 952.75 	 P99 1123.00 (ms)
```

Launch against `https://youtube.com`, run for 2 minutes at a 300 milliseconds frequency:

```
checkit -url https://youtube.com -duration 2 -frequency 300
2017/04/28 13:12:50 Checking URL 'https://youtube.com'
2017/04/28 13:12:50 Running for 2.0 minute(s) at a 300 millisecond(s) frequency
2017/04/28 13:12:50 
2017/04/28 13:12:50 Beginning checks at 2017-04-28 13:12:50.108452272 +0100 BST
....................................................................................................................................................................................................................................................................................
2017/04/28 13:14:50 Ending checks at 2017-04-28 13:14:50.430113974 +0100 BST
2017/04/28 13:14:50 
2017/04/28 13:14:50 *** RESULTS (in milliseconds) ***
2017/04/28 13:14:50 Performed 276 checks
2017/04/28 13:14:50 Establish connection  	 Min 10 	 Mean 12.52 	 Max 23 	 P75 13.00 	 P99 22.23 (ms)
2017/04/28 13:14:50 Send, wait, receive   	 Min 51 	 Mean 61.44 	 Max 85 	 P75 64.00 	 P99 81.92 (ms)
2017/04/28 13:14:50 Round trip            	 Min 63 	 Mean 74.58 	 Max 100 	 P75 78.00 	 P99 93.69 (ms)
2017/04/28 13:14:50 HTTP GET (i.e. total) 	 Min 115 	 Mean 132.41 	 Max 244 	 P75 136.00 	 P99 156.46 (ms)
```

## License

**MIT**. Please see `LICENSE` file on top of the repository.