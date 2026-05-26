# Tube Medic Web

A modern web experience for Tube Medic, allowing creators to scan their YouTube channel for broken links through a simple, interactive dashboard.

## Features

- **HTMX Powered:** Fast, dynamic UI updates without complex JavaScript frameworks.
- **Revenue Intelligence:** Automatically identifies and prioritizes high-value affiliate and product links.
- **Concurrent Scanning:** Leverages Go's concurrency to scan dozens of videos and hundreds of links in seconds.
- **Modern UI:** Built with Tailwind CSS for a clean, professional aesthetic.

## Prerequisites

- **Go 1.26+**
- **YouTube Data API v3 Key:** You'll need an API key from the Google Cloud Console.

## Setup

1. **Environment Variables:**
   Create a `.env` file in the root or set the following variable:
   ```bash
   YT_API_KEY=your_youtube_api_key_here
   ```

2. **Run with Makefile:**
   From the root of the repository, you can use the workspace-level commands:
   ```bash
   make run-web
   ```

## Development

The project uses Go Workspaces (`go.work`) to share logic between the CLI tool and the web interface. 

- `web/cmd/server/`: The application entry point.
- `web/internal/handlers/`: Request handlers and business logic coordination.
- `web/templates/`: HTML templates (Layout and Content).
- `tube-medic-mvp/internal/`: Shared core logic for YouTube scraping and link checking.
