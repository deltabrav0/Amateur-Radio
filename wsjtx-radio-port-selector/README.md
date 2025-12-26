# WSJT-X Radio Port Selector  
Automatically Fix WSJT-X CAT Port Detection on macOS

![badge](https://img.shields.io/badge/platform-macOS-blue)
![badge](https://img.shields.io/badge/status-stable-brightgreen)
![badge](https://img.shields.io/badge/license-MIT-lightgrey)

---

## 📡 Overview

On macOS, Silicon Labs USB-to-UART adapters frequently enumerate as:

```
/dev/cu.SLAB_USBtoUART
/dev/cu.SLAB_USBtoUART4
/dev/cu.SLAB_USBtoUART9
```

This creates a problem for rigs and digital‑mode software like WSJT‑X or wsjtx‑improved:  
**the port keeps changing**, and WSJT‑X stops communicating with the radio.

This tool resolves the problem automatically.

---

## ✨ What This Tool Does

Every time you run the launcher:

1. **Detects the newest Silicon Labs UART device**
2. **Creates a stable symlink:**

```
/usr/local/serial/radio → /dev/cu.SLAB_USBtoUART*
```

3. **Updates `CATSerialPort=` inside:**  
   `~/Library/Preferences/WSJT-X.ini`
4. **Gracefully quits and restarts `wsjtx-improved`**
5. Ships with a clickable macOS app:  
   **WSJT-X-fixed.app**  
   (Spotlight searchable, sits in `/Applications`)

---

## 🛠 Installation

Clone the repository:

```bash
git clone https://github.com/REPLACEME/wsjtx-radio-port-selector.git
cd wsjtx-radio-port-selector
```

Run the installer:

```bash
./install/install.sh
```

This installs:

- `/usr/local/bin/link-radioWSJT-X.sh`
- `/Applications/WSJT-X-fixed.app`

---

## ▶️ Usage

Launch via Spotlight:

```
WSJT-X-fixed
```

Or run manually:

```bash
link-radioWSJT-X.sh
```

---

## 📦 Repository Contents

```
scripts/
    link-radioWSJT-X.sh      # Main script that fixes the port

app/
    WSJT-X-fixed.app/        # macOS launcher

install/
    install.sh               # Installer
    uninstall.sh             # Clean removal
```

---

## 🔧 Requirements

- macOS 12+
- hamlib which provides rigctl.  can be installed with homebrew (brew install
  hamlib)
- `wsjtx-improved.app` installed in `/Applications`
- Silicon Labs USB UART driver

---

## 🧪 Testing

You can simulate a device reattach by running:

```bash
ls -ltr /dev/cu.SLAB_USBtoUART*
```

Then launch the app and verify WSJT-X switches to:

```
CATSerialPort=/usr/local/serial/radio
```

---

## 📝 License

This project is licensed under the MIT License — see `LICENSE` for details.

---

## 🙌 Credits

Developed by **Dan Butler (K5SUB)**  
Project assistance by **TARS**.

