#[macro_use]
extern crate log;
use crate::prelude::*;
use clap::Parser;

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
    // Setup logger
    //let args = args::DesktopCleanerArgs::parse();
    let config = config::DesktopCleanerConfig::init()?;
    println!("Config: {:?}", config.file_types);
    let debug_level = config.map_debug_level();
    println!("Debug Level: {:?}", debug_level);
    logger::DesktopCleanerLogger::init(debug_level).unwrap();
    // get args
    //debug!("Args: {:?}", args);
    // read config

    // get folders
    let folders = config.file_types.keys();

    for folder in folders {
        debug!("Folder: {}", folder);
    }

    // get the appropriate directory from the user- if none provided use the default for their desktop
    let path = std::path::PathBuf::from("./");

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
