package main

import "net"
import "fmt"
import "time"
import "ipgenerator"

var debug int = 1

type scan_stats struct {
	Success int
	Failure int
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
	for {
		select {
			case <-controller:
				if debug > 0 {
					fmt.Printf("Stopping this runner.")
				}
				results<-scan_stats{success, failure}
				return
			default:
				break
		}
		new_ip := generator.IP(seed, true)
		if connect(new_ip, 80) {
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
func connect(target net.IPAddr, port uint16) bool {
	address := target.IP.String() + ":" + fmt.Sprint(port)
	var conn net.Conn
	var err error

	if conn, err = net.DialTimeout("tcp", address, 1*time.Second); err != nil {
		return false
	}
	conn.Close()
	return true
}

func main() {
	executive := make(chan runner_controller)
	results := make(chan scan_stats)
	generator := ipgenerator.NewRandomIPGenerator()
	total_runners := 2

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
		time.Sleep(10*time.Second)
		close(executive)
	}()

	successes := 0
	failures := 0
	for i:=0; i<total_runners; i++ {
		scan_stats := <-results
		successes += scan_stats.Success
		failures += scan_stats.Failure
		fmt.Printf("Success: %v, failure: %v\n", scan_stats.Success, scan_stats.Failure)
	}
	fmt.Printf("Total Successes: %v, Total Failures: %v\n", successes, failures)
}
