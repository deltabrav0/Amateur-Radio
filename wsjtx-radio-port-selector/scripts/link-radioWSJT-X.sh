#!/bin/bash

osascript <<EOF
tell application "wsjtx-improved" to quit
EOF

set -euo pipefail

TARGET="/usr/local/serial"
LINK="$TARGET/radio"
INI="$HOME/Library/Preferences/WSJT-X.ini"
APP="/Applications/wsjtx-improved.app"
APP_NAME="wsjtx-improved"

mkdir -p "$TARGET"

# ----- Function: Restart WSJT-X Improved safely -----
restart_wsjt() {
    APP="/Applications/wsjtx-improved.app"
    EXEC_NAME="wsjtx"
    APPLESCRIPT_NAME="wsjtx-improved"

    if [[ ! -d "$APP" ]]; then
        echo "Application not found at $APP — skipping restart."
        return
    fi

    echo "Restarting wsjtx-improved..."

    # Quit if running
    if pgrep -x "$EXEC_NAME" >/dev/null; then
        echo "wsjtx-improved is running. Attempting to quit..."

        osascript <<EOF
tell application "$APPLESCRIPT_NAME" to quit
EOF

        # Wait for clean termination
        while pgrep -x "$EXEC_NAME" >/dev/null; do
            sleep 0.3
        done
    fi

    echo "Launching wsjtx-improved..."
    open "$APP"
}

# ----- Detect newest Silicon Labs port -----
PORT=$(ls -t /dev/cu.SLAB_USBtoUART* 2>/dev/null | head -n 1 || true)

if [[ -n "${PORT:-}" ]]; then
    echo "Found port: $PORT"
    ln -sf "$PORT" "$LINK"
    echo "Updated symlink: $LINK → $PORT"

    # Update WSJT-X.ini if it exists
    if [[ -f "$INI" ]]; then
        echo "Updating CATSerialPort in $INI"

        if grep -q '^CATSerialPort=' "$INI"; then
            sed -i '' "s|^CATSerialPort=.*|CATSerialPort=$LINK|" "$INI"
        else
            if grep -q '^\[Rig\]' "$INI"; then
                awk -v v="CATSerialPort=$LINK" '
                    $0=="[Rig]" { print; print v; next }
                    { print }
                ' "$INI" > "$INI.tmp" && mv "$INI.tmp" "$INI"
            else
                echo "CATSerialPort=$LINK" >> "$INI"
            fi
        fi

        echo "WSJT-X CATSerialPort updated → $LINK"

        # Restart wsjtx-improved
        restart_wsjt
    else
        echo "WSJT-X.ini not found, skipping update and restart."
    fi
else
    echo "No SLAB USB devices found. Removing stale symlink."
    rm -f "$LINK"
fi
