# WSJT-X Radio Port Selector

# WSJT-X Radio Port Selector

This tool automatically:

- Detects the newest Silicon Labs USB UART device  
- Updates a stable symlink at `/usr/local/serial/radio`  
- Updates `CATSerialPort=` in `~/Library/Preferences/WSJT-X.ini`  
- Restarts `wsjtx-improved`  
- Provides a macOS launcher app (`WSJT-X-fixed.app`) you can run from Spotlight


Includes:

- Shell script (`link-radioWSJT-X.sh`)
- macOS launcher app (`Update Radio Port.app`)
- Optional installer script

## Requirements

- macOS
- wsjtx-improved installed in `/Applications`
- Silicon Labs USB UART driver

## Installation

Clone the repo:

```bash
git clone https://github.com/YOURNAME/wsjtx-radio-port-selector.git
cd wsjtx-radio-port-selector
./install/install.sh

## Usage
Launch the app
```WSJT-X-fixed

Or run manually 
link-radioWSJT-X.sh
