# Disk Usage Analyzer (Windows)

A high-performance, concurrent disk usage analyzer written in Go. 
It scans a directory tree starting from a root path, calculates total disk usage, 
and displays the top N largest files. Designed specifically for Windows and Linux
operating systems. For windows this program can handles quirks like NTFS compression 
and system-protected directories.

## 🚀 Features

- Multi-threaded directory scanning using goroutines and semaphores
- Top N largest files output
- Skips common restricted directories (`PerfLogs`, `$Recycle.Bin`, symlinks)
- Correct file size accounting via Win32 system calls
- Depth-limited recursive walking
- Graceful error handling
- Minimal dependencies

`ConcurrentWalk` function automatically decides which operating system is being used

Windows solution has reliance on package golang.org/x/sys/windows

Linux will use standard library to return filesize

Does not follow symlinks or mount points

Designed for performance and visibility, not archival or deletion


## 🏗️ Build & Run

Requires **Go 1.20+** and **Windows / Linux OS**

```pwsh
go build -o main.exe
```

Then run the executable with:

```pwsh
.\main.exe -r C:\ -d -1 -n 20
```

flags:
- `r` root directory to start scanning from. Default: terminal working directory
- `d` depth to limit the number of subfolders to recursively search.
Set to negative one for no limit. Default depth set to zero (no recursion)
- `n` number of files to return. Default top 10 largest files

### Example

```pwsh
.\main.exe -r C:\ -d -1
```

**Output**
 
```pwsh
Scanning Directory: C:\ (depth: ∞)

Top 10 largest files:
 1. 42.66 GB   C:\Users\david\AppData\Local\Docker\wsl\disk\docker_data.vhdx
 2. 24.54 GB   C:\Users\david\AppData\Local\Packages\CanonicalGroupLimited.Ubuntu_79rhkp1fndgsc\LocalState\ext4.vhdx
 3. 15.84 GB   C:\pagefile.sys
 4. 6.62 GB    C:\Users\david\Documents\fooocus\Fooocus_win64_2-1-791\Fooocus\models\checkpoints\juggernautXL_v7Rundiffusion.safetensors
 5. 6.46 GB    C:\Users\david\Documents\fooocus\Fooocus_win64_2-1-791\Fooocus\models\checkpoints\bbbSDXL_bbbBetaV2.safetensors
 6. 6.34 GB    C:\hiberfil.sys
 7. 4.58 GB    C:\Users\david\.ollama\models\blobs\sha256-667b0c1932bc6ffc593ed1d03f895bf2dc8dc6df21db3042284a6f4416b06a29
 8. 4.00 GB    C:\Users\david\AppData\Local\NVIDIA\DXCache\fa7ba9db43ede110.nvph
 9. 2.44 GB    C:\Users\david\AppData\Local\pip\cache\http\4\5\3\7\b\4537bd14d5cc94e7b73fac5299ea971d99b17d3a6a4adbc8eace0df2
10. 2.32 GB    C:\Users\david\.ollama\models\blobs\sha256-3c168af1dea0a414299c7d9077e100ac763370e5a98b3c53801a958a47f0a5db

Scanned 330712 directories
Scanned 1780777 files
Total disk space scanned 399.95 GB
Scan Duration: 6.0580701s
```
