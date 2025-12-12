# LoTW Prometheus Exporter & Dashboard

A comprehensive monitoring stack for Amateur Radio [Logbook of the World (LoTW)](https://lotw.arrl.org/) activity. This project fetches your QSO logs, parses the ADIF data, and visualizes it in a beautiful pre-configured Grafana dashboard.

![Dashboard Preview](https://i.imgur.com/placeholder-dashboard.png) *Note: Replace with actual screenshot*

## Features

- **Automated Fetching**: Downloads your latest QSLs and QSOs from LoTW API every hour.
- **Full History**: Automatically backfills your entire log history from 1900-present on the first run.
- **Rich Dashboard**:
  - **Totals**: Heads-up display of Total QSOs and Confirmed QSLs.
  - **Band & Mode Breakdown**: Interactive Pie Charts showing your most active bands and operating modes.
  - **Daily History**: A "Recent Activity" stacked bar chart (Last 30 Days) to track your daily logging cadence.
- **Cross-Platform**: Runs on **Windows**, **macOS**, and **Linux** (including Raspberry Pi) via Docker.

## Prerequisites

- **Docker Desktop** (or Docker Engine + Compose plugin).
  - [Download for Mac](https://docs.docker.com/desktop/install/mac-install/)
  - [Download for Windows](https://docs.docker.com/desktop/install/windows-install/)
  - [Download for Linux](https://docs.docker.com/engine/install/)
- An active [ARRL LoTW Account](https://lotw.arrl.org/lotw-help/getting-started/).

## Installation

### 1. Download
Clone this repository or download the ZIP file.

```bash
git clone https://github.com/yourusername/lotw-exporter.git
cd lotw-exporter/deploy
```

### 2. Configure Credentials
Create a `.env` file to store your LoTW username and password. We have provided an example file for you.

**Mac / Linux:**
```bash
cp .env.example .env
nano .env
```

**Windows (PowerShell):**
```powershell
Copy-Item .env.example .env
notepad .env
```

**Edit the `.env` file**:
```ini
LOTW_USERNAME=YourCallsign
LOTW_PASSWORD=YourLotwPassword
FETCH_INTERVAL=1h
```

### 3. Run
Start the entire stack (Exporter, Prometheus, Grafana) with a single command.

```bash
docker compose up -d
```
*Wait a few seconds for the containers to start.*

## Accessing the Dashboard

1. Open your browser to [http://localhost:3000](http://localhost:3000).
2. Login with default credentials:
   - **User**: `admin`
   - **Password**: `admin`
3. Click on the **"LoTW Stats"** dashboard.

## Operation

- **Initial Load**: On the first start, the exporter effectively performs a "Full Import", downloading all your records from LoTW. This ensures your stats are complete from day one.
- **Recurring**: It will check for new records every hour (configurable via `FETCH_INTERVAL`).
- **Data Persistence**: Data is stored in Docker volumes (`prometheus_data`, `grafana_data`), so your history is safe even if you restart containers.

## Troubleshooting

- **No Data?**
  - Check the logs: `docker logs lotw_exporter`
  - Ensure your username/password are correct in `.env`.
  - Verify LoTW is online.
- **Windows Users**: Ensure Docker Desktop is running. You should see the whale icon in your system tray.

## License
MIT
