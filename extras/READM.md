# Extras

This project was prototyped as a python script, you can find the source code for that in this folder.

I am keeping here for posterity, but I will not be updating it anymore.

It does not contain all of the features of the final version, but it is still functional.

## Desktop Cleaner (Python)

### From source

1. Copy the Script
2. Run the script

```bash
python3 src/desktop-cleaner.py
```

### Build from source

1. Copy the Script and the `.spec` file
2. Install `pyinstaller` (if not already installed)

```bash
pip3 install pyinstaller
```

3. Build the executable

```bash
pyinstaller --onefile src/desktop-cleaner.py
```

## List of file types

> **Note**: These can be changed in the python script itself, the config file to make it easier to add/remove file types and change their associated folders is only supported in the rust version on main.

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

[MIT](../LICENSE)
