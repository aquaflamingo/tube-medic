1. Defeat TLS Fingerprinting (The Most Crucial Step)
   This is almost certainly why your basic Go program is getting blocked. WAFs inspect the TLS "Client Hello" packet (often using a method called JA3 fingerprinting). Go’s standard net/http library announces its supported cipher suites and extensions in a very specific order that screams, "I am a Go script, not a web browser."

To fix this, you need to spoof the TLS fingerprint of a real browser:

Use uTLS: Instead of crypto/tls, use the github.com/refraction-networking/utls package. It allows you to mimic the exact cryptographic handshake of Chrome, Firefox, or Safari.

Use tls-client: An even easier wrapper is github.com/bogdanfinn/tls-client. It handles both TLS and HTTP/2 fingerprinting out of the box, allowing you to easily say, "Make this request look exactly like Chrome 120."

2. Perfect Your HTTP Headers
   You can't just slap a standard Chrome User-Agent on a request and expect it to work anymore. Bot protections look for the precise combination of headers that modern browsers send.

Include Sec- Headers: Modern WAFs heavily weigh Sec-Fetch-Dest, Sec-Fetch-Mode, Sec-Fetch-Site, and Sec-CH-UA headers.

Header Order Matters: Go's net/http standardizes header capitalization and sometimes orders them alphabetically. Real browsers send headers in a specific order. Libraries like tls-client help preserve browser-specific header ordering.

Accept-Language: Always include realistic locale headers (e.g., en-US,en;q=0.9), as bots often forget them.

3. Implement a Strict Cookie Jar
   If you are making multiple requests, you need to maintain session state.

When you visit eBay in a browser, the first request often returns a set of tracking/session cookies, and the bot protection system evaluates your trust score over time.

Ensure your Go http.Client is configured with an http.CookieJar. If you make 50 requests and none of them pass back the cookies provided in the previous responses, the WAF will instantly flag you as a bot and issue a 403.

4. Manage HTTP/2 Fingerprinting
   Just like TLS, HTTP/2 connections have fingerprints (based on settings frames, window updates, and stream priorities). Go's default HTTP/2 implementation is highly identifiable.

Downgrade to HTTP/1.1: Sometimes, forcing your client to downgrade to HTTP/1.1 (by customizing the transport) can bypass HTTP/2-specific bot checks, though this is becoming less effective as WAFs evolve.

Use a spoofing library: Again, tls-client handles HTTP/2 frame spoofing to match Chrome/Firefox.

5. Transition to a Headless Browser (If API calls fail)
   If eBay's bot protection is actively serving JavaScript challenges (which execute in the browser to collect mouse movements, canvas fingerprints, and hardware data before granting access), purely HTTP-based requests will never succeed.

If you hit a hard wall with standard HTTP requests, you may need to use a library like github.com/chromedp/chromedp or playwright-go.

These tools actually spin up a real instance of Chromium locally, execute the JavaScript challenges, and pass the bot checks. You can then extract the resulting cookies and pass them back into your lightweight Go HTTP client for faster subsequent scraping.
