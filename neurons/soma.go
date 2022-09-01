package neurons

import (
	"fmt"
	"strings"
	"sync"
)

type Client struct {
	Mu       sync.Mutex
	Uid      string
	Hostname string
	Username string
	Response chan []byte
	Jobs     chan string
	Transfer chan string
	Pulse    chan struct{}
}

type ClientList struct {
	Mu      sync.Mutex
	List    map[string]*Client // string key will be the uid sha256 hash
	Current string
}

var Clients = ClientList{
	List: make(map[string]*Client),
}

const padLength = 66
const numberOfFields = 3

var frame string

func init() {
	var tmpFrame strings.Builder
	padding := strings.Repeat("-", padLength)
	for i := 0; i < numberOfFields; i++ {
		tmpFrame.WriteString("+")
		tmpFrame.WriteString(padding)
	}
	tmpFrame.WriteString("+")
	frame = tmpFrame.String()
}

func leftPad(s string, padStr string, pLen int) string {
	return strings.Repeat(padStr, pLen) + s
}

func rightPad(s string, padStr string, pLen int) string {
	return s + strings.Repeat(padStr, pLen)
}

// ref: https://stackoverflow.com/a/47823574
func center(s string, w int) string {
	return fmt.Sprintf("%*s", -w, fmt.Sprintf("%*s", (w+len(s))/2, s))
}

func (c *ClientList) Display() {
	fmt.Println(frame)
	if len(c.List) == 0 {
		fmt.Println("|", center("No clients connected.", (padLength*numberOfFields)+2), "|")
	} else {
		for uid, client := range c.List {
			fmt.Println("|", center(uid, padLength), "|", center(client.Hostname, padLength), "|", center(client.Username, padLength), "|")
		}
	}
	fmt.Println(frame)
}

func (c *ClientList) Size() int {
	return len(c.List)
}

func (c *ClientList) DisplayCurrent() {
	if c.HasCurrent() {
		fmt.Println(c.Current)
	} else {
		fmt.Println("No current client selected.")
	}
}

func (c *ClientList) GetCurrent() *Client {
	return c.List[c.Current]
}

func (c *ClientList) HasCurrent() bool {
	return c.Current != ""
}

func (c *ClientList) Cleanup(uid string) {
	fmt.Printf("\nClient dead: %s\n", uid)
	c.Mu.Lock()
	if c.Current == uid {
		c.Current = ""
	}
	delete(c.List, uid)
	c.Mu.Unlock()
}

func (c *ClientList) Add(uid string, name string) {
	fmt.Printf("\nClient connected: %s\n", uid)
	c.Mu.Lock()
	c.List[uid] = &Client{
		Uid:      uid,
		Hostname: strings.Split(name, ":")[0],
		Username: strings.Split(name, ":")[1],
		Response: make(chan []byte),
		Jobs:     make(chan string, 1),
		Transfer: make(chan string, 1),
		Pulse:    make(chan struct{}, 1),
	}
	c.Mu.Unlock()
}
