#[macro_use]
extern crate log;

use crate::prelude::*;

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
    logger::DesktopCleanerLogger::init(log::LevelFilter::max()).unwrap();

    info!("Hello, world!");
    handle_dir::DirEntry::get_dir_entry("./".to_string())?;

    // TODO: create static map of file extensions to folder names
    // TODO: create global files moved counter
    // TODO: write main logic loop

    Ok(())
}
