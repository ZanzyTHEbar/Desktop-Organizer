# Desktop Cleaner

This is a simple tool that helps you to clean your desktop (or any directory) by moving specific groups of file types to a designated folder.

Currently, this tool does not support moving files between mount points/file systems. This is because the tool uses the `rename` function from the `std::fs` module, which does not support moving files between mount points/file systems.

This tool also does not support a target folder that is not a subdirectory of the source folder.

You pass in a directory, and the tool will organize that directory by moving files of specific types to a designated folder within that directory.

You can configure the file types and the folders in the config file.

if you do not specify a path, the tool will use your desktop, if a desktop cannot be found, it will use your home.

You can always use the `.` path to specify the current directory.

## Development

- [x] Add support for config file
- [x] Add support for custom file types
- [x] Add support for custom folders
- [x] Automatically create config file if it does not exist
- [x] Add colors to the output
- [ ] Add support for subdirectories
- [ ] Add installers
- [ ] Sign releases

## Usage

```bash
desktop_cleaner [OPTIONS] [PATH]
```

### Options

```bash
--help, -h      Prints help information
--version, -v   Prints version information
```

### Arguments

```bash
PATH            The path to the folder you want to clean
```

## Installation

You can install from the releases or build from source.

> [!IMPORTANT]\
> MacOS ARM64 users will have to build from source, as I do not have a new Mx Mac to build the executable on, yet.
> If you have a Mac and would like to help, please open an issue, I will be very grateful.
> Otherwise, if you feel like supporting me, check out my Github Sponsors Page, I will be very grateful too :smile:

### Releases

> [!NOTE]\
> Releases are only available for Linux and Windows at the moment.

> [!WARNING]\
> Releases are not signed, so you will get a warning from your OS when you try to download and run the executable, this is normal and you can safely ignore it.

1. Download the latest release from [here](https://github.com/ZanzyTHEbar/Desktop-Cleaner/releases)
2. Place the executable somewhere on your computer
3. Add the location of the executable to your PATH

### From source

1. Clone the repository
2. Make sure rust is installed (if not, follow the instructions [here](https://www.rust-lang.org/tools/install))

```bash
cargo build --release
```

Or to build and install

```bash
cargo build --release && install --path <path>
```

To test the executable

```bash
cargo run <args for executable>
```

Example:

```bash
cargo run -- -h
```

## List of file types

Please see the [config file](/docs/.desktop_cleaner.toml) for an example.

The list of file types and their associated folders is only limited by your imagination (and the rules of your OS).

## License

[MIT](/LICENSE)
