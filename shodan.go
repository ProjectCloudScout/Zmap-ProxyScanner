package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
)

const maxShodanResults = 100000

type shodanMatch struct {
	IP        int64    `json:"ip"`
	Port      int      `json:"port"`
	Product   string   `json:"product"`
	Hostnames []string `json:"hostnames"`
}

type proxyCandidate struct {
	Address  string
	Protocol string
	Port     int
}

func buildProxyURL(protocol, address string, port int) string {
	protocol = strings.ToLower(protocol)
	if protocol == "socks4" || protocol == "socks5" {
		return fmt.Sprintf("%s://%s:%d", protocol, address, port)
	}
	return fmt.Sprintf("http://%s:%d", address, port)
}

func candidateFromMatch(match shodanMatch) proxyCandidate {
	protocol := "http"
	product := strings.ToLower(match.Product)
	switch {
	case strings.Contains(product, "socks5"):
		protocol = "socks5"
	case strings.Contains(product, "socks4"):
		protocol = "socks4"
	case strings.Contains(product, "proxy"):
		protocol = "http"
	case strings.Contains(product, "http"):
		protocol = "http"
	}

	return proxyCandidate{
		Address:  ipFromInt(match.IP),
		Protocol: protocol,
		Port:     match.Port,
	}
}

func ipFromInt(ip int64) string {
	if ip == 0 {
		return "0.0.0.0"
	}
	return fmt.Sprintf("%d.%d.%d.%d", (ip>>24)&0xff, (ip>>16)&0xff, (ip>>8)&0xff, ip&0xff)
}

func buildShodanQuery(port int, protocol, userQuery string) string {
	if port <= 0 {
		port = 80
	}

	if strings.TrimSpace(userQuery) != "" {
		return strings.TrimSpace(userQuery)
	}

	parts := make([]string, 0, 3)
	parts = append(parts, "proxy")
	if strings.TrimSpace(protocol) != "" {
		parts = append(parts, strings.TrimSpace(protocol))
	}
	parts = append(parts, fmt.Sprintf("port:%d", port))
	return strings.Join(parts, " ")
}

func shouldContinueShodanPagination(collected, total, limit int) bool {
	if limit <= 0 {
		return false
	}
	if total <= 0 {
		return false
	}
	return collected < total && collected < limit
}

func discoverFromShodan(apiKey string, port int, protocol, userQuery string, limit int, onCandidate func(proxyCandidate)) ([]proxyCandidate, error) {
	if strings.TrimSpace(apiKey) == "" {
		return nil, fmt.Errorf("SHODAN_API_KEY is empty; export it first: export SHODAN_API_KEY='your_key'")
	}

	query := buildShodanQuery(port, protocol, userQuery)

	if limit <= 0 {
		limit = maxShodanResults
	}

	var allCandidates []proxyCandidate
	page := 1
	for page <= 1000 {
		endpoint := fmt.Sprintf("https://api.shodan.io/shodan/host/search?key=%s&query=%s&page=%d",
			url.QueryEscape(apiKey), url.QueryEscape(query), page)

		resp, err := http.Get(endpoint)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode == http.StatusUnauthorized {
			resp.Body.Close()
			return nil, fmt.Errorf("shodan rejected the API key (401 Unauthorized); verify SHODAN_API_KEY is valid and exported")
		}
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, fmt.Errorf("shodan request failed: %s", resp.Status)
		}

		var payload struct {
			Matches []shodanMatch `json:"matches"`
			Total   int           `json:"total"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
			resp.Body.Close()
			return nil, err
		}
		resp.Body.Close()

		for _, match := range payload.Matches {
			candidate := candidateFromMatch(match)
			allCandidates = append(allCandidates, candidate)
			if onCandidate != nil {
				onCandidate(candidate)
			}
		}
		if len(payload.Matches) == 0 || !shouldContinueShodanPagination(len(allCandidates), payload.Total, limit) {
			break
		}
		page++
	}

	return allCandidates, nil
}

func writeCandidates(path string, candidates []proxyCandidate) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	for _, candidate := range candidates {
		if _, err := f.WriteString(fmt.Sprintf("%s\n", buildProxyURL(candidate.Protocol, candidate.Address, candidate.Port))); err != nil {
			return err
		}
	}
	return nil
}
