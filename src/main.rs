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

    // read config
    let config = config::DesktopCleanerConfig::init()?;
    let debug_level = config.map_debug_level();

    // Setup logger
    logger::DesktopCleanerLogger::init(debug_level).unwrap();

    debug!("Config: {:?}", config.file_types);
    debug!("Debug Level: {:?}", debug_level);

    let mut directory = cli_args.directory.unwrap_or_else(|| {
        debug!("Failed to get directory, using the default for your OS");
        String::from("")
    });

    let mut recursive = cli_args.recursive.unwrap_or_else(|| {
        debug!("Failed to get recursive, using the default for your OS");
        false
    });

    let mut hidden = cli_args.hidden.unwrap_or_else(|| {
        debug!("Failed to get hidden, using the default for your OS");
        false
    });

    for (key, value) in config.file_types.iter() {
        debug!("Key: {}, Value: {:?}", key, value);
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
    if debug_level != log::LevelFilter::Off {
        debug!("Debugging enabled");
    }

    if directory.is_empty() {
        if let Some(user_dirs) = UserDirs::new() {
            let path = user_dirs.desktop_dir().unwrap_or_else(|| {
                warn!("Failed to get desktop dir, using the home dir instead");
                user_dirs.home_dir()
            });
            debug!("Location Dir: {:?}", path);
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
    let mut files_moved = 0;

    // iterate over each dir entry and compare the file type to the config file types
    for dir_entry in dir_entries.dir_entries.unwrap() {
        if !dir_entry.is_dir {
            for (key, value) in config.file_types.iter() {
                // prefix the file type with a period
                let mut file_type = String::from(".");
                file_type.push_str(dir_entry.file_type.as_str());
                if value.contains(&file_type) {
                    info!("File Type: {}", file_type);
                    let mut new_path = path.join(key);
                    std::fs::create_dir_all(&new_path)?;
                    // append the file name to the new path
                    new_path = new_path.join(dir_entry.file_name.clone());
                    info!("New Path: {:?}", new_path);
                    // rename the file
                    std::fs::rename(&dir_entry.path, &new_path)?;
                    files_moved += 1;
                    println!("Moved {} to {}", dir_entry.file_name, key);
                    // if the file is a symlink, skip it
                    // if the file is a hidden file, skip it
                    // if the file is a directory, recursively call the function if that config option is enabled
                }
            }
        } else {
            /* if recursive {
                debug!("Recursive enabled");
                // recursively call the function if that config option is enabled
                let path = dir_entry.path.clone();
                let mut dir_entries = handle_dir::DirEntries::default();
                handle_dir::DirEntry::get_dirs(&path, &mut dir_entries)?;
                debug!("---------------------------------");
                dir_entries.print_dir_entries()?;
                debug!("---------------------------------");
            } */
        }
    }

    match files_moved {
        0 => println!("No files moved"),
        1 => println!("Moved 1 file"),
        _ => println!("Moved {} files", files_moved),
    }

    Ok(())
}
