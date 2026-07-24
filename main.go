/*
   (c) Yariya
*/

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"time"
)

var port = flag.Int("p", 80, "proxy port")
var output = flag.String("o", "output.txt", "output file")
var configFile = flag.String("cfg", "config.json", "configuration file")

var input = flag.String("in", "", "input file to check")
var fetch = flag.String("url", "", "url proxy fetch")
var shodanMode = flag.Bool("shodan", false, "discover candidates from Shodan")
var shodanPort = flag.Int("shodan-port", 80, "port to search for with Shodan")
var shodanProtocol = flag.String("shodan-protocol", "", "protocol filter: http, socks4, socks5")
var shodanOutput = flag.String("shodan-output", "shodan-proxies.txt", "file to write discovered proxies to")
var shodanLimit = flag.Int("shodan-limit", 100000, "maximum number of Shodan hits to collect")
var liveInterval = flag.Duration("live-interval", 2*time.Second, "refresh interval for the live status line")
var query = flag.String("q", "", "query string for Shodan discovery")
var live = flag.Bool("l", false, "print live output as results arrive")

const wt = 3

type Api struct {
	Status string `json:"Status"`
	Reason string `json:"Reason"`
}

type Config struct {
	CheckSite   string `json:"check-site"`
	ProxyType   string `json:"proxy-type"`
	HttpThreads int    `json:"http_threads"`
	Headers     struct {
		UserAgent string `json:"user-agent"`
		Accept    string `json:"accept"`
	} `json:"headers"`
	PrintIps struct {
		Enabled       bool `json:"enabled"`
		DisplayIpInfo bool `json:"display-ip-info"`
	} `json:"print_ips"`
	Timeout struct {
		HttpTimeout   int `json:"http_timeout"`
		Socks4Timeout int `json:"socks4_timeout"`
		Socks5Timeout int `json:"socks5_timeout"`
	} `json:"timeout"`
}

var config Config

type liveStatus struct {
	hits   int
	latest string
}

func startLiveStatusPrinter(interval time.Duration, status *liveStatus, done <-chan struct{}) {
	spinner := []string{"|", "/", "-", "\\"}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	idx := 0
	for {
		select {
		case <-done:
			fmt.Printf("\r\033[2K[shadowmpp] live done hits=%d latest=%s\n", status.hits, status.latest)
			return
		case <-ticker.C:
			fmt.Printf("\r\033[2K[shadowmpp] live hits=%d latest=%s %s", status.hits, status.latest, spinner[idx%len(spinner)])
			idx++
		}
	}
}

func resolveQueryFromArgs(args []string, explicitQuery string) string {
	if strings.TrimSpace(explicitQuery) != "" {
		return explicitQuery
	}

	parts := make([]string, 0, len(args))
	for _, arg := range args {
		if strings.HasPrefix(arg, "-") {
			break
		}
		if strings.TrimSpace(arg) == "" {
			continue
		}
		parts = append(parts, arg)
	}
	if len(parts) == 0 {
		return ""
	}
	return strings.TrimSpace(strings.Join(parts, " "))
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Zmap ProxyScanner\n")
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [flags]\n", os.Args[0])
		flag.PrintDefaults()
	}

	if len(os.Args) > 1 && (os.Args[1] == "-h" || os.Args[1] == "--help") {
		flag.Usage()
		return
	}
	resolvedQuery := resolveQueryFromArgs(os.Args[1:], *query)
	flag.Parse()
	rand.Seed(time.Now().UnixNano())

	cfgBytes, err := os.ReadFile(*configFile)
	if err != nil {
		log.Println("error while opening config file")
		return
	}
	err = json.Unmarshal(cfgBytes, &config)
	if err != nil {
		fmt.Println("error while parsing config json")
		return
	}

	if *shodanMode || resolvedQuery != "" {
		apiKey := os.Getenv("SHODAN_API_KEY")
		proto := *shodanProtocol
		if resolvedQuery != "" {
			proto = ""
		}
		var emit func(proxyCandidate)
		status := &liveStatus{}
		statusDone := make(chan struct{})
		if *live {
			go startLiveStatusPrinter(*liveInterval, status, statusDone)
			emit = func(candidate proxyCandidate) {
				status.hits++
				status.latest = fmt.Sprintf("%s:%d", candidate.Address, candidate.Port)
				PrintProxyWithProtocol(candidate.Address, candidate.Port, candidate.Protocol)
			}
		}
		candidates, err := discoverFromShodan(apiKey, *shodanPort, proto, resolvedQuery, *shodanLimit, emit)
		if err != nil {
			log.Fatalf("shodan discovery failed: %v", err)
		}
		if err := writeCandidates(*shodanOutput, candidates); err != nil {
			log.Fatalf("writing shodan candidates failed: %v", err)
		}
		fmt.Printf("\033[35m[shadowmpp]\033[39m query=%q\n", resolvedQuery)
		fmt.Printf("\033[35m[shadowmpp]\033[39m discovered=%d\n", len(candidates))
		if *live {
			fmt.Printf("\033[32m[+]\033[39m live output enabled\n")
		}
		if *live {
			close(statusDone)
		} else {
			for i, candidate := range candidates {
				if i >= 5 {
					break
				}
				PrintProxyWithProtocol(candidate.Address, candidate.Port, candidate.Protocol)
			}
		}
		fmt.Printf("\033[36m[save]\033[39m %s\n", *shodanOutput)
		return
	}

	_ = os.Remove(*output)

	exporter = &Exporter{
		out: *output,
	}

	fmt.Println("=== Zmap ProxyScanner ===")
	fmt.Printf("mode=scan port=%d output=%s\n", *port, *output)
	fmt.Println("starting workers...")
	go exporter.create()
	go Queue()
	go Scanner()
	for x := 0; x < wt; x++ {
		go Proxies.WorkerThread()
	}
	go Stater()
	time.Sleep(time.Second)
	fmt.Println("live updates: every 1s")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, os.Kill, os.Interrupt)
	<-sc
	PrintSummary()
	exporter.Close()
}
