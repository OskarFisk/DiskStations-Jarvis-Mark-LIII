# MARK LIII Universal Browser

This folder contains the device-independent browser interface for MARK LIII.

## Use

Open `web/index.html` in a modern browser. Enter the URL of a running MARK LIII server, for example `http://192.168.1.50:8080`, then save it.

The interface is designed for Windows, Linux, macOS, ChromeOS, Android, iPhone/iPad and other modern browsers.

## Architecture

The browser is the client. MARK LIII remains the server/API. This keeps the Gemini API key off the client when the key is configured on the server.

For production/public hosting, serve the file over HTTPS and put the MARK LIII API behind HTTPS as well. Configure appropriate CORS/authentication on the server before exposing it to the Internet.
