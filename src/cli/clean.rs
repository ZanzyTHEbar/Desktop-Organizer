use crate::config;
use crate::handle_dir;
use crate::prelude::*;
use directories::UserDirs;
use std::path::PathBuf;

pub struct CleanHandler {
    pub directory: Option<PathBuf>,
    pub recursive: Option<bool>,
    pub hidden: Option<bool>,
}

impl CleanHandler {
    pub fn new(directory: Option<PathBuf>, recursive: Option<bool>, hidden: Option<bool>) -> Self {
        Self {
            directory,
            recursive,
            hidden,
        }
    }

    pub fn try_parse_directory(src: Option<PathBuf>) -> Result<Option<PathBuf>> {
        dc_stdout!(f!("{}", Color::new("Parsing directory...").bold().green()).as_str());

        let src = match src {
            Some(src) => src,
            None => {
                if let Some(user_dirs) = UserDirs::new() {
                    let path = user_dirs.desktop_dir().unwrap_or_else(|| {
                        warn!("Failed to get desktop dir, using the home dir instead");
                        user_dirs.home_dir()
                    });
                    debug!("Location Dir: {:?}", path);
                    return Ok(Some(path.to_path_buf()));
                }
                return Err(Error::OperationCancelled(f!(
                    "{}",
                    Color::new("Failed to get desktop dir, using the home dir instead")
                        .bold()
                        .red()
                )));
            }
        };

        if !src.exists() {
            return Err(Error::OperationCancelled(f!(
                "{}",
                Color::new("Could not find the path specified").bold().red()
            )));
        }

        Ok(Some(src))
    }

    pub fn run(&self) -> Result<()> {
        // read config
        // Linux	$XDG_CONFIG_HOME/_project_path_ or $HOME/.config/_project_path_	/home/alice/.config/barapp
        // macOS	$HOME/Library/Application Support/_project_path_	/Users/Alice/Library/Application Support/com.Foo-Corp.Bar-App
        // Windows	{FOLDERID_RoamingAppData}\_project_path_\config	C:\Users\Alice\AppData\Roaming\Foo Corp\Bar App\config
        let config = config::DesktopCleanerConfig::init()?;
        let debug_level = config.map_debug_level();

        // Setup logger
        DesktopCleanerLogger::init(debug_level).unwrap();

        debug!("Config: {:?}", config.file_types);
        debug!("Debug Level: {:?}", debug_level);

        // get user dirs
        // Linux	XDG_DESKTOP_DIR	/home/alice/Desktop
        // macOS	$HOME/Desktop	/Users/Alice/Desktop
        // Windows  {FOLDERID_Desktop}	C:\Users\Alice\Desktop
        if debug_level != log::LevelFilter::Off {
            debug!("Debugging enabled");
        }

        // get the appropriate src from the user- if none provided use the default for their desktop
        let path = Self::try_parse_directory(self.directory.clone())?;
        let path = path.as_ref().unwrap();

        // get dir entries
        let mut dir_entries = handle_dir::DirEntries::default();
        handle_dir::DirEntry::get_dirs(path.as_path(), &mut dir_entries)?;
        debug!("---------------------------------");
        dir_entries.print_dir_entries()?;
        debug!("---------------------------------");

        // move files
        let mut files_moved = 0;

        // iterate over each dir entry and compare the file type to the config file types
        for dir_entry in dir_entries.dir_entries.unwrap() {
            if !dir_entry.is_dir {
                for (key, value) in config.file_types.as_ref().unwrap().iter() {
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
                        dc_stdout!(f!("Moved {} to {}", dir_entry.file_name, key).as_str());
                        // if the file is a symlink, skip it
                        // if the file is a hidden file, skip it
                        // if the file is a src, recursively call the function if that config option is enabled
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
            0 => dc_stdout!("Nothing to Do"),
            1 => dc_stdout!("Moved 1 file"),
            _ => dc_stdout!(f!("Moved {} files", files_moved).as_str()),
        }

        Ok(())
    }
}
