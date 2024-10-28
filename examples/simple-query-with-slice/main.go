package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/zadoryzsolt/gosolar"
)

func main() {
	hostname := "localhost"
	username := "admin"
	password := ""

	// NewClient creates a client that will handle the connection to SolarWinds
	// along with the timeout and HTTP conversation.
	client := gosolar.NewClient(hostname, username, password, true)

	ctx := context.Background()

	// put the query into a string using a multi-line string assignment
	query := `
		SELECT
			Caption
			,IPAddress
		FROM Orion.Nodes
		WHERE Vendor = @vendor
		AND Status IN @statuses
	`

	// build a map that will hold the parameters for the query above including
	// a slice for the IN portion of the query
	parameters := map[string]interface{}{
		"vendor":   "Cisco",
		"statuses": []int{1, 2, 3},
	}

	// run the query without with the parameters map above
	res, err := client.Query(ctx, query, parameters)
	if err != nil {
		log.Fatal(err)
	}

	// build a structure to unmarshal the results into
	var nodes []struct {
		Caption   string `json:"caption"`
		IPAddress string `json:"ipaddress"`
	}

	// let unmarshal do the work of unpacking the JSON
	if err := json.Unmarshal(res, &nodes); err != nil {
		log.Fatal(err)
	}

	// iterate over the resulting slice of node structures
	for _, n := range nodes {
		fmt.Printf("Working with node [%s] on IP address [%s]...\n", n.Caption, n.IPAddress)
	}
}
