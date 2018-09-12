package ipgenerator

import "sync"
import "math/rand"
import "net"
import "time"

type IPGenerator interface {
	IP(seed net.IPAddr, success bool) net.IPAddr
}

type RandomIPGenerator struct {
	random *rand.Rand
	m sync.Mutex
}

func NewRandomIPGenerator() RandomIPGenerator {
	generator := RandomIPGenerator{}
	generator.random = rand.New(rand.NewSource(time.Now().Unix()))
	return generator
}

func (s RandomIPGenerator) IP(seed net.IPAddr, success bool) net.IPAddr {
	var value uint8
	var octets [4]uint8
	s.m.Lock()
	for i:=0;i<4;i++ {
		value = 0
		for {
			value = uint8(rand.Uint32() % 255)
			if value != 0 {
				break
			}
		}
		octets[i] = value
	}
	s.m.Unlock()

	result := net.IPAddr{net.IPv4(octets[0],octets[1], octets[2], octets[3]),""}

	return result
}


