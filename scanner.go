package main

import "net"
import "fmt"
import "time"
import "ipgenerator"
import "github.com/gonum/stat"

var debug int = 1

type scan_stats struct {
	Success int
	Failure int
	AvgTime float64
	StdDevTime float64
}
type runner_controller struct {}

func LocalhostIP() net.IPAddr {
	return net.IPAddr{net.ParseIP("127.0.0.1"), ""}
}

/*
 * In a loop, get a random IP address and then try to connect.
 * Do this until next_ip fails to return an IP address.
 */
func runner(generator ipgenerator.IPGenerator, seed net.IPAddr, results chan scan_stats, controller chan runner_controller) {
	success := 0
	failure := 0
	times := []float64{}
	for {
		select {
			case <-controller:
				if debug > 0 {
					fmt.Printf("Stopping this runner.")
				}
				results<-scan_stats{success,
				                    failure,
						    stat.Mean(times, nil),
						    stat.StdDev(times, nil)}
				return
			default:
				break
		}
		new_ip := generator.IP(seed, true)
		connected, time := connect(new_ip, 445)
		times = append(times, float64(time.Nanoseconds()))
		if connected {
			if debug > 0 {
				fmt.Printf("Trying %v: Success.\n", new_ip);
			}
			success++
		} else {
			if debug > 0 {
				fmt.Printf("Trying %v: Failure.\n", new_ip);
			}
			failure++
		}
	}
}

/*
 * Attempt to connect to IP address target on port port
 * and return true/false based on whether the connection
 * was successful.
 */
func connect(target net.IPAddr, port uint16) (bool, time.Duration) {
	address := target.IP.String() + ":" + fmt.Sprint(port)
	var conn net.Conn
	var err error
	var start time.Time
	var elapsed time.Duration
	var success bool

	start = time.Now()
	conn, err = net.DialTimeout("tcp", address, 1*time.Second)
	elapsed = time.Now().Sub(start)

	if err != nil {
		success = false
	} else {
		success = true
		conn.Close()
	}

	return success, elapsed
}

func main() {
	executive := make(chan runner_controller)
	results := make(chan scan_stats)
	generator := ipgenerator.NewRandomIPGenerator()
	total_runners := 1

	/*
	 * Create runners that will collect statistics.
	 */
	for i:=0; i<total_runners; i++ {
		go runner(generator, LocalhostIP(), results, executive)
	}

	go func() {
		/*
		 * Let us run for a period of time.
		 */
		time.Sleep(60*time.Minute)
		close(executive)
	}()

	successes := 0
	failures := 0
	for i:=0; i<total_runners; i++ {
		scan_stats := <-results
		successes += scan_stats.Success
		failures += scan_stats.Failure
		fmt.Printf("Success: %v, failure: %v, average connect time: %v, std dev connect time: %v\n", scan_stats.Success, scan_stats.Failure, scan_stats.AvgTime, scan_stats.StdDevTime)
	}
	fmt.Printf("Total Successes: %v, Total Failures: %v\n", successes, failures)
}
