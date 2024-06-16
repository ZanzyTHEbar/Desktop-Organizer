# Project Port to Go

## Overview

`Desktop Cleaner` is a simple CLI tool that helps you to organize a directory by organizing and moving files/folder to a specified locations.

## Main Features

- [ ] Read a config file to get the rules to organize the files/folders
  - Config file is an `ini` file
  - Config file maps the file/folder type to the destination folder name
    - Folder name can be a path to the destination folder
    - Example: `.pdf` files should be moved to `PDF` folder
- [ ] Support custom File Types
  - User can specify the file types to be organized
  - User can specify the destination folder for the file types
- [ ] Support custom Folders
- [ ] Automatically create config file if it does not exist
- [ ] Add colors to the output
- [ ] Add support for subdirectories
- [ ] Add installer and updater
- [ ] Sign releases

### Advanced Features

- [ ] Impl option to skip hidden files/folders
- [ ] Impl skipping symlinks
- [ ] Impl searching subdirectories
- [ ] Impl creating sub directories
- [ ] Impl support for other mount points and targeting a specific destination
- [ ] Impl support for multiple source directories
- [ ] Impl support for multiple destination directories
- [ ] Impl LLM cmd assistant
  - LLM to convert query into action plan.
  - Action plan that will walk the directory (and sub-directories if chosen) and then batch analyze files for context.
    - Summarize context and find relationships.
    - Group files into relevant directories based on relationship.
  - Create temporary config file
  - Execute desktop cleaner using this config file.

### Options

```bash
--help, -h      Prints help information
--version, -v   Prints version information
```

### Arguments

```bash
PATH            The path to the folder you want to clean
```
