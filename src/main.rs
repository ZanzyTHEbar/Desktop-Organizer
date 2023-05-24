#[macro_use]
extern crate log;
use crate::prelude::*;
use clap::Parser;
use directories::UserDirs;

mod args;
mod config;
mod error;
mod handle_dir;
mod logger;
mod prelude;
mod utils;

/**  
 *  Logic Loop
 *  1. Get config
 *  2. Get dir entries by detecting the OS and then getting the directory to Desktop or home if config setting enabled
 *     iterating through the input path and creating a DirEntry struct for each entry
 *  3. Iterate through each DirEntry and move the file to the appropriate folder
 *   - If the file is a symlink, skip it
 *   - If the file is a hidden file, skip it
 *   - if the appropriate folder does not exist, create it
 *   - If the file is a directory, recursively call the function if that config option is enabled
 *   - If the file is not a directory, move the file to the appropriate folder
 *  4. Print the number of files moved
 *  5. Save the config if the config has been modified
 */
fn main() -> Result<()> {
    // get args
    let cli_args = args::DesktopCleanerArgs::parse();

    let mut directory = cli_args.directory.unwrap_or_else(|| {
        println!("Failed to get directory, using the default for your OS");
        String::from("")
    });

    // read config
    let config = config::DesktopCleanerConfig::init()?;
    println!("Config: {:?}", config.file_types);
    let debug_level = config.map_debug_level();
    println!("Debug Level: {:?}", debug_level);

    // Setup logger
    logger::DesktopCleanerLogger::init(debug_level).unwrap();

    for (key, value) in config.file_types.iter() {
        println!("Key: {}, Value: {:?}", key, value);
    }

    // get folders
    let folders = config.file_types.keys();
    for folder in folders {
        debug!("Folder: {}", folder);
    }

    // get user dirs
    // Linux	XDG_DESKTOP_DIR	/home/alice/Desktop
    // macOS	$HOME/Desktop	/Users/Alice/Desktop
    // Windows  {FOLDERID_Desktop}	C:\Users\Alice\Desktop
    /* if args.subcmd {
        debug!("Debugging enabled");
    } */

    if directory.is_empty() {
        if let Some(user_dirs) = UserDirs::new() {
            let path = user_dirs.desktop_dir().unwrap_or_else(|| {
                println!("Failed to get desktop dir, using the home dir instead");
                user_dirs.home_dir()
            });
            println!("Location Dir: {:?}", path);
            directory = path.to_str().unwrap().to_string();
        }
    }
    // get the appropriate directory from the user- if none provided use the default for their desktop
    let path = std::path::PathBuf::from(directory);

    // get dir entries
    let mut dir_entries = handle_dir::DirEntries::default();
    handle_dir::DirEntry::get_dirs(&path, &mut dir_entries)?;
    debug!("---------------------------------");
    dir_entries.print_dir_entries()?;
    debug!("---------------------------------");

    // move files

    // print number of files moved

    // TODO: create global files moved counter
    // TODO: write main logic loop

    Ok(())
}
