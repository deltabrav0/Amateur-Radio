#!/bin/bash

set -e

echo "Installing WSJT-X Radio Port Selector..."

# Install script
install -m 755 ./scripts/link-radioWSJT-X.sh /usr/local/bin/

# Install the application
cp -R ./app/WSJT-X-fixed.app /Applications/

echo "Installation complete. You can now launch WSJT-X-fixed from Spotlight."
