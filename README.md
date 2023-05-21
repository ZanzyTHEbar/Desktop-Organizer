# Desktop Cleaner

This is a simple python script that helps you to clean your desktop by moving specific groups of files to a designated folder.

## Usage

### Releases

> **Note**: Releases are only available for Linux and Windows at the moment.

> **Warning**: Releases are not signed, so you will get a warning from your OS when you try to download and run the executable, this is normal and you can safely ignore it.

1. Download the latest release from [here](https://github.com/ZanzyTHEbar/Desktop-Cleaner/releases)
2. Run the executable

### From source

1. Clone the repository
2. Run the script

```bash
python3 src/desktop-cleaner.py
```

## Build from source

1. Clone the repository
2. Install pyinstaller (if not already installed)

```bash
pip3 install pyinstaller
```

3. Build the executable

```bash
pyinstaller --onefile src/desktop-cleaner.py
```

## List of file types

> **Note**: I am working on setting up a config file to make it easier to add/remove file types and change their associated folders.

```bash
"CODE": [".c", ".h", ".py", ".rs", ".go", ".js", "ts", ".jsx", "tsx", ".html", ".css", ".php", ".java", ".cpp", ".cs", ".vb", ".sql", ".pl", ".swift", ".kt", ".r", ".m", ".asm"],
"MARKUP": [".json", ".xml", ".yml", ".yaml", ".ini", ".toml", ".cfg", ".conf", ".log", ".md"],
"NOTES": [".md", ".rtf", ".txt"],
"DOCS": [".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx"],
"EXE": [".exe", ".appimage", ".msi"],
"VIDS": [".mp4", ".mov", ".avi", ".mkv"],
"COMPRESSED": [".zip", ".rar", ".tar", ".gz", ".7z"],
"SCRIPTS": [".sh", ".bat"],
"INSTALLERS": [".deb", ".rpm"],
"BOOKS": [".epub", ".mobi"],
"MUSIC": [".mp3", ".wav", ".ogg", ".flac"],
"PDFS": [".pdf"],
"PICS": [".bmp", ".gif", ".jpg", ".jpeg", ".svg", ".png"],
```

## License

[MIT](/LICENSE)
