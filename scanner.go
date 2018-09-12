package main

import "net"
import "fmt"
import "time"
import "sync"
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
	var waiter sync.WaitGroup

	waiter.Add(1)

	/*
	 * Create a single runner that will collect statistics.
	 */
	go runner(generator, LocalhostIP(), results, executive)

	go func() {
		/*
		 * Let us run for a period of time.
		 */
		time.Sleep(5*time.Minute)
		close(executive)

		/*
		 * Cool off time.
		 */
		time.Sleep(5*time.Second)
		waiter.Done()
	}()

	waiter.Wait()

	scan_stats := <-results
	fmt.Printf("Success: %v, failure: %v\n", scan_stats.Success, scan_stats.Failure)
}
