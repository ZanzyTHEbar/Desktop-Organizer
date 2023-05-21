""" Clean up your Desktop Automagically """

import pathlib
import os
from sys import platform

# Find the path to the Desktop
if platform == "linux" or platform == "linux2" or platform == "darwin":
    desktop_path = os.path.expanduser("~/Desktop")
    # check if the path exists
    if not os.path.exists(desktop_path):
        # if not, use home directory
        print("[!] Desktop directory not found, using home directory.")
        desktop_path = os.path.expanduser("~/") 
    desktop = pathlib.Path(desktop_path)
elif platform == "win32":
    desktop_path = os.path.expanduser("~\\Desktop")
    if not os.path.exists(desktop_path):
        # if not, find in onedrive
        print("[!] Local Desktop not found, using OneDrive Desktop directory.")
        desktop_path = os.path.expanduser("~\\OneDrive\\Desktop")
    desktop = pathlib.Path(desktop_path)


def main():
    """ Main function """
    # create dict of file types and their respective folders
    file_types = {
        "CODE": [".c", ".h", ".py"],
        "NOTES": [".md", ".rtf", ".txt"],
        "PDFS": [".pdf"],
        "PICS": [".bmp", ".gif", ".jpg", ".jpeg", ".svg", ".png"],
    }
    files_moved = 0
    for each in desktop.iterdir():
        if each.is_file():
            for key, value in file_types.items():
                if each.suffix in value:
                    new_dir = desktop.joinpath(key)
                    new_dir.mkdir(exist_ok=True)
                    new_path = desktop / key / each.name
                    each.rename(new_path)
                    files_moved += 1
                    print(f"Moved {each.name} to {new_path}")
    if files_moved > 0:
        print(f"[+] Successfully moved {files_moved} files!")
    else:
        print("[!] No files were moved.")

main()
