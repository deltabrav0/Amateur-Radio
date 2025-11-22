# Persistent Serial Port Naming on macOS for CP210x USB‑UART Adapters  
### A practical method for ham radio operators, microcontroller users, and developers

macOS does **not** provide persistent device naming for USB‑serial adapters (unlike Linux `udev`).  
Silicon Labs CP210x UART devices often appear randomly as:

```
/dev/cu.SLAB_USBtoUART
/dev/cu.SLAB_USBtoUART2
/dev/cu.SLAB_USBtoUART3
/dev/cu.SLAB_USBtoUART4
```

The number changes when:

- you reboot  
- unplug/replug  
- switch USB ports  
- or sometimes for no obvious reason  

This guide provides a **reliable, repeatable, stable** serial device:

```
/usr/local/serial/radio
```

This symlink always points to the **newest** `/dev/cu.SLAB_USBtoUART*` device created by macOS, which is consistently the *active* port.

---

# ✅ Why `/dev` cannot be used for symlinks

macOS **System Integrity Protection (SIP)** prevents creating files inside `/dev`.  
So instead, we place a persistent directory elsewhere and dynamically update a symlink that points **into** `/dev`.

---

# ## 1. Create a persistent directory for your stable port

```bash
sudo mkdir -p /usr/local/serial
sudo chmod 777 /usr/local/serial
```

This folder will hold the stable symlink:

```
/usr/local/serial/radio
```

---

# ## 2. Create the updater script

```bash
sudo vim /usr/local/bin/link-radio.sh
```

Paste in the script below:

```bash
#!/bin/bash

# Pick the most recently created Silicon Labs UART device
PORT=$(ls -t /dev/cu.SLAB_USBtoUART* 2>/dev/null | head -n 1)

if [[ -n "$PORT" ]]; then
    # Update the stable symlink
    ln -sf "$PORT" /usr/local/serial/radio
else
    # Remove symlink if device is not present
    rm -f /usr/local/serial/radio
fi
```

Make it executable:

```bash
sudo chmod +x /usr/local/bin/link-radio.sh
```

---

# ## 3. Test it manually

Run:

```bash
/usr/local/bin/link-radio.sh
```

Check:

```bash
ls -l /usr/local/serial/radio
```

Example:

```
/usr/local/serial/radio -> /dev/cu.SLAB_USBtoUART4
```

Now unplug your serial device and plug it back in.

Repeat:

```bash
/usr/local/bin/link-radio.sh
ls -l /usr/local/serial/radio
```

The symlink should now point to whichever `/dev/cu.SLAB_USBtoUART*` entry is newest.

To confirm which one is newest, run:

```bash
ls -ltr /dev/cu.SLAB_USBtoUART*
```

The bottom entry is the correct, active one.

---

# ## 4. Automate the process with launchd (updates every 5 seconds)

Create the launchd plist:

```bash
sudo vim /Library/LaunchDaemons/com.radio.linker.plist
```

Paste:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" 
"http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.radio.linker</string>

    <key>ProgramArguments</key>
    <array>
        <string>/usr/local/bin/link-radio.sh</string>
    </array>

    <!-- Run every 5 seconds -->
    <key>StartInterval</key>
    <integer>5</integer>

    <key>RunAtLoad</key>
    <true/>
</dict>
</plist>
```

Load it:

```bash
sudo launchctl unload /Library/LaunchDaemons/com.radio.linker.plist 2>/dev/null
sudo launchctl load /Library/LaunchDaemons/com.radio.linker.plist
```

Now the script runs every five seconds and automatically corrects the symlink whenever macOS re‑enumerates the device.

---

# ## 5. Use the stable serial port

Use **this** path in WSJT-X, CHIRP, fldigi, rigctld, Python, etc.:

```
/usr/local/serial/radio
```

No more tracking device numbers.  
No more guessing.  
No more renumbering surprises after reboots or cable swaps.

---

# ## 6. Optional: uninstall

```bash
sudo launchctl unload /Library/LaunchDaemons/com.radio.linker.plist
sudo rm /Library/LaunchDaemons/com.radio.linker.plist
sudo rm /usr/local/bin/link-radio.sh
sudo rm -r /usr/local/serial
```

---

# 🎯 Final Summary

This method gives you:

- A stable, permanent serial-device path  
- Automatic updates when devices are plugged/unplugged  
- Full compatibility with all ham radio software  
- Behavior resilient to macOS renumbering  
- No modification of SIP-protected areas  

Your new, persistent serial port is:

```
/usr/local/serial/radio
```

Share this file freely with other Mac-based radio operators and developers.
