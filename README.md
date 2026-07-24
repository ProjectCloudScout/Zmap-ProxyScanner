# Zmap-ProxyScanner

A Thread Safe fast way to find proxies. Find 2000-5000 working http,socks4,socks5 proxies in one scan.

![379e86d10c8e05e9d21a20647d37c70ea0d5e976c72a44a2a5506c88d31e5cf3](https://user-images.githubusercontent.com/65712074/195901928-721235f2-163e-4266-ae4e-d7c76b2626d2.png)

# Config
  ```json
   {
    "check-site": "https://google.com",
    "proxy-type": "http",

    "http_threads": 2000,
    "headers": {
      "user-agent": "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/102.0.5005.115 Safari/537.36",
      "accept": "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8"
    },
    "print_ips": {
      "enabled": true,
      "display-ip-info": true
    },
    "timeout": {
      "http_timeout": 5,
      "socks4_timeout": 5,
      "socks5_timeout": 5
    }
  }
  ```
## Flag Args
  ```shell
-p <port> - Port you want to scan.
-o <proxies.txt> - Writes proxy hits to file.
-input <proxies.txt> - Loads the proxy list and checks it.
-url https://api.com/proxies - Loads the proxies from an api and checks it.
-shodan - Discover candidate proxies from Shodan.
-shodan-port <port> - Port to search for when using -shodan.
-shodan-protocol <http|socks4|socks5> - Protocol filter for Shodan discovery.
-shodan-output <file> - Output file for discovered candidates.
  ```


# Features
  * Scan entire world for http,socks4 and socks5 proxies.
  * Inbuilt file + (from url) proxy scanner
  * Display ISP, Country
  
# Example Run
  * Be Sure to use an Hosting that allows Portscan just like https://pfcloud.io
  > zmap -p 8080 | ./ZmapProxyScanner -p 8080

  * Shodan discovery example (requires SHODAN_API_KEY)
  > SHODAN_API_KEY=your-key ./ZmapProxyScanner -shodan -shodan-port 8080 -shodan-protocol http -shodan-output proxies.txt

# Sample Output
  ```text
  Shodan discovery complete
  Candidates discovered: 3
    - http://198.51.100.10:8080
    - http://203.0.113.20:8080
    - http://192.0.2.55:8080
  Saved to proxies.txt
  ```

# Verification / Live Stats
  During a scan, the binary prints live progress every second, for example:
  ```text
  Imported [12] IPs Checked [30] IPs (Success: 4, StatusCodeErr: 2, ProxyErr: 1, Timeout: 1) with 3 open http threads
  ```
  When you stop the run, it prints a final summary block so the results are easy to verify in the terminal.

# Build
  > Requires go v1.19+
  ```shell
  git clone https://github.com/Yariya/Zmap-ProxyScanner.git
  cd Zmap-ProxyScanner
  go build
  ```
