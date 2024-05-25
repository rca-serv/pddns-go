package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"
)

const (
	defaultTTL = 300
)

type Record struct {
	Content  string `json:"content"`
	Name     string `json:"name"`
	TTL      int    `json:"ttl"`
	Type     string `json:"type"`
	Disabled bool   `json:"disabled"`
}

type RRSet struct {
	Name       string   `json:"name"`
	Type       string   `json:"type"`
	TTL        int      `json:"ttl"`
	ChangeType string   `json:"changetype"`
	Records    []Record `json:"records"`
}

type DNSUpdate struct {
	RRSets []RRSet `json:"rrsets"`
}

func getLocalIPAddress(interfaceName string) (string, error) {
	iface, err := net.InterfaceByName(interfaceName)
	if err != nil {
		return "", fmt.Errorf("failed to get interface: %v", err)
	}

	addrs, err := iface.Addrs()
	if err != nil {
		return "", fmt.Errorf("failed to get addresses: %v", err)
	}

	for _, addr := range addrs {
		switch v := addr.(type) {
		case *net.IPNet:
			if v.IP.To4() != nil {
				return v.IP.String(), nil
			}
		}
	}

	return "", fmt.Errorf("no valid IPv4 address found for interface %s", interfaceName)
}

func main() {
	// Get environment variables
	apiKey := os.Getenv("PDNS_API_KEY")
	ownName := os.Getenv("PDNS_OWN_NAME")
	pdnsServer := os.Getenv("PDNS_SERVER")
	interfaceName := os.Getenv("PDNS_INTERFACE")
	ttlStr := os.Getenv("PDNS_TTL")
	zoneName := os.Getenv("PDNS_ZONE")

	if apiKey == "" || ownName == "" || pdnsServer == "" || interfaceName == "" || zoneName == "" {
		fmt.Println("Please set the environment variables PDNS_API_KEY, PDNS_OWN_NAME, PDNS_SERVER, PDNS_INTERFACE, and PDNS_ZONE")
		return
	}

	// Parse TTL or use default
	ttl := defaultTTL
	if ttlStr != "" {
		ttlParsed, err := strconv.Atoi(ttlStr)
		if err == nil {
			ttl = ttlParsed
		}
	}

	for {

		// Get local IP address
		localIP, err := getLocalIPAddress(interfaceName)
		if err != nil {
			fmt.Printf("Error getting local IP address: %v\n", err)
			return
		}

		// Create record
		record := Record{
			Content:  localIP,
			Name:     fmt.Sprintf("%s.%s.", ownName, zoneName),
			TTL:      ttl,
			Type:     "A",
			Disabled: false,
		}

		rrset := RRSet{
			Name:       record.Name,
			Type:       record.Type,
			TTL:        record.TTL,
			ChangeType: "REPLACE",
			Records:    []Record{record},
		}

		update := DNSUpdate{
			RRSets: []RRSet{rrset},
		}

		// Encode to JSON
		body, err := json.Marshal(update)
		if err != nil {
			fmt.Printf("Error encoding JSON: %v\n", err)
			return
		}

		// Send request to PowerDNS server
		powerDNSAPIURL := fmt.Sprintf("http://%s/api/v1/servers/localhost/zones/%s", pdnsServer, zoneName)
		req, err := http.NewRequest("PATCH", powerDNSAPIURL, bytes.NewBuffer(body))
		if err != nil {
			fmt.Printf("Error creating request: %v\n", err)
			return
		}

		// Set headers
		req.Header.Set("X-API-Key", apiKey)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("Error sending request: %v\n", err)
			return
		}
		defer resp.Body.Close()

		// Check response
		if resp.StatusCode != http.StatusNoContent {
			fmt.Printf("Request failed: status code %d\n", resp.StatusCode)
		} else {
			fmt.Println("Resource record updated successfully!")
		}


        // Sleep for 30 minutes
        // TODO: 時間を上書きできるようにする
        time.Sleep(30 * time.Minute)
        fmt.Println("Successfully updated DNS record!")
        fmt.Println("IP: ", localIP)
        fmt.Println("Sleeping for 30 minutes...")

	}
}
