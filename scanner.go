package main

import "net"
import "fmt"
import "time"
import "math/rand"

type scan_stats struct {
	Success int
	Failure int
}

var debug int = 1

/*
 * Generate random bytes. Do this until someone closes our
 * executive.
 */
func rand_gen(random chan uint8, executive chan struct{}) {
	var value uint8 = 0
	rand := rand.New(rand.NewSource(time.Now().Unix()))
	for {
		value = uint8(rand.Uint32() % 255)
		select {
			case random<-value:
				break
			case <-executive:
				if debug > 0 {
					fmt.Printf("This channel was closed.\n")
				}
				close(random)
				return
			default:
		}
	}
}

/*
 * In a loop, get a random IP address and then try to connect.
 * Do this until next_ip fails to return an IP address.
 */
func runner(random chan uint8, seed net.IPAddr, results chan scan_stats) {
	success := 0
	failure := 0
	for {
		var new_ip net.IPAddr
		var okay bool
		if new_ip, okay= next_ip(random, seed, true); !okay {
			fmt.Printf("Failed to get another IP address. Quitting.\n")
			results <- scan_stats{success, failure}
			return
		}
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
 * Generate a /random/ IP address. This is a very brain
 * dead way to do it. Simply get four random bytes and
 * do not accept 0s.
 */
func next_ip(random chan uint8, previous net.IPAddr, success bool) (net.IPAddr, bool) {
	var octets [4]uint8
	for i:=0; i<4; i++ {
		var okay bool
		var octet uint8
		for {
			octet, okay = <-random
			if !okay {
				/*
				 * If we did not get a random value okay, then we should quit.
				 */
				return net.IPAddr{net.ParseIP("0.0.0.0"), ""}, false
			}
			/*
			 * We may have gotten 0 and we don't want that.
			 */
			if octet == 0 {
				continue;
			} else {
				break
			}
		}
		octets[i] = octet
	}
	return net.IPAddr{net.IPv4(octets[0],octets[1], octets[2], octets[3]),""}, true
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
	random := make(chan uint8)
	executive := make(chan struct{})
	results := make(chan scan_stats)

	/*
	 * Create a random byte generator.
	 */
	go rand_gen(random, executive)

	/*
	 * Create a single runner that will collect statistics.
	 */
	go runner(random, net.IPAddr{net.ParseIP("0.0.0.0"), ""}, results);

	go func() {
		/*
		 * Let us run for a period of time.
		 */
		time.Sleep(2*time.Minute)
		close(executive)

		/*
		 * Cool off time.
		 */
		time.Sleep(5*time.Second)
	}()

	scan_stats := <-results
	fmt.Printf("Success: %v, failure: %v\n", scan_stats.Success, scan_stats.Failure)
}
