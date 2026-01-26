# Running Primortal

## macOS

### First Time Setup

After downloading and extracting the game:

**Option 1: Use the helper script (recommended)**
```bash
./run_macos.sh
```

This script will automatically remove the quarantine attribute and launch the game.

**Option 2: Manual one-liner**
```bash
xattr -dr com.apple.quarantine primortal
./primortal
```

**Option 3: Right-click method**
1. Right-click `primortal`
2. Select "Open"
3. Click "Open" in the security dialog

You only need to do this once per download.

---

## Windows

### First Time Setup

After downloading and extracting the game:

**Option 1: Use the helper script (recommended)**

Double-click `run_windows.bat`

This script will automatically unblock the files and launch the game.

**Option 2: Manual PowerShell one-liner**

Open PowerShell in the game folder and run:
```powershell
Get-ChildItem -Recurse | Unblock-File
.\primortal.exe
```

**Option 3: Right-click method**
1. Right-click `primortal.exe`
2. Select "Properties"
3. Check "Unblock" at the bottom
4. Click "OK"
5. Run `primortal.exe`

You only need to do this once per download.

---

## Why These Steps Are Needed

Both macOS and Windows flag downloaded executables as potentially unsafe. Since this game is not code-signed with an Apple Developer ID or Windows certificate, you'll see security warnings.

The helper scripts bypass these warnings by removing the "downloaded from internet" flags that trigger the security prompts.

**This is safe for software you trust.** Only run these scripts on downloads from sources you trust.

---

## Troubleshooting

### macOS: "primortal cannot be opened because the developer cannot be verified"
- Use the helper script or the manual one-liner above
- Or: System Preferences → Security & Privacy → Click "Open Anyway"

### Windows: "Windows protected your PC" SmartScreen warning
- Use the helper script or click "More info" → "Run anyway"
- The manual unblock method also works

### Game won't launch
- Make sure you extracted the entire archive, not just the executable
- Check that you're running from the extracted folder
- On macOS, ensure the binary has execute permissions: `chmod +x primortal`
