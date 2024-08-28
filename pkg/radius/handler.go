package radius

import (
	"fmt"
	"log"

	"layeh.com/radius"
)

func HandleAccessRequest(c *Client, packet *radius.Packet) {
	log.Printf("Handling Access-Request for user: %s", fmt.Sprintf("%v", packet))
	// Add your handling logic here
}
