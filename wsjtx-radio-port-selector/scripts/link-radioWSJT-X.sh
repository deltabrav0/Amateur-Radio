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

# ======================================================================
#   NEW PORT DETECTION (rigctl validates the actual working radio port)
# ======================================================================

RIGCTL=${RIGCTL:-"rigctl"}
RIGCTLOPT=${RIGCTLOPT:-"/opt/homebrew/bin/rigctl"}
SLAB="SLAB"

# Hamlib ID for your radio (verified by your rigctl tests)
HAMLIBID=1042

# Locate rigctl
RIGCTLCMD=$(which "$RIGCTL" 2>/dev/null || true)
if [[ -z "$RIGCTLCMD" && -f "$RIGCTLOPT" ]]; then
    RIGCTLCMD="$RIGCTLOPT"
fi

if [[ -z "$RIGCTLCMD" ]]; then
    echo "Error: hamlib rigctl not installed. Install via brew and try again."
    exit 1
fi

echo "Using rigctl: $RIGCTLCMD"

shopt -s nullglob

PORT=""
echo "Probing Silicon Labs ports..."

for f in /dev/cu.*$SLAB* /dev/tty.*$SLAB*; do
    echo " - Testing $f ..."
    if "$RIGCTLCMD" -m "$HAMLIBID" -r "$f" _ >/dev/null 2>&1; then
        echo " --> Success: rigctl can communicate via $f"
        PORT="$f"
        break
    fi
done

if [[ -z "$PORT" ]]; then
    echo "No functioning SLAB USB ports detected. Removing stale symlink."
    rm -f "$LINK"
    exit 1
fi

echo "Detected working CAT port: $PORT"

# ======================================================================
#   Create symlink & update WSJT-X.ini
# ======================================================================

ln -sf "$PORT" "$LINK"
echo "Updated symlink: $LINK → $PORT"

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
    restart_wsjt
else
    echo "WSJT-X.ini not found, skipping update and restart."
fi
