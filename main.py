""" Clean up your Desktop Automagically """

import pathlib
import os
from sys import platform

# Find the path to the Desktop
if platform == "linux" or platform == "linux2" or platform == "darwin":
    #desktop = pathlib.Path(r"~/")
    #desktop = os.path.join(os.path.join(os.path.expanduser('~')), 'Desktop')
    desktop = os.path.expanduser("~/Desktop")
    # check if the path exists
    if not os.path.exists(desktop):
        # if not, use home directory
        desktop = os.path.expanduser("~") 
#elif platform == "darwin":
    #desktop = pathlib.Path(r"~/Desktop")
    #desktop = os.path.join(os.path.join(os.path.expanduser('~')), 'Desktop') 
elif platform == "win32":
    os.path.expanduser("~\\Desktop")
    #desktop = os.path.join(os.path.join(os.environ['USERPROFILE']), 'Desktop')
    #desktop = pathlib.Path(r"C:\Users\MyName\Desktop")

# Create a new folder
for sub_dir in ["CODE", "NOTES", "PDFS", "PICS"]:
    new_dir = desktop.joinpath(sub_dir)
    new_dir.mkdir(exist_ok=True)

files_moved = 0

# Filter for screenshots only
for each in desktop.iterdir():
    if not each.is_file():
        # Skip directories
        continue

    extension = each.suffix.lower()

    # Create a new path for each file
    if extension in [".c", ".h", ".py"]:
        # put it in CODE
        new_path = desktop / "CODE" / each.name

    elif extension in [".md", ".rtf", ".txt"]:
        # put it in NOTES
        new_path = desktop / "NOTES" / each.name

    elif extension == ".pdf":
        # put it in PDFS
        new_path = desktop / "PDFS" / each.name

    elif extension in [".bmp", ".gif", ".jpg", ".jpeg", ".svg", ".png"]:
        # put it in PICS
        new_path = desktop / "PICS" / each.name

    else:
        continue

    # Move the screenshot there
    print(f"[!] Moving {each} to {new_path}...")
    each.rename(new_path)
    files_moved += 1

if files_moved > 0:
    print(f"[+] Successfully moved {files_moved} files!")
