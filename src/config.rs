use crate::prelude::*;
use directories::ProjectDirs;
use log::warn;
use serde::Deserialize;
use std::collections::HashMap;

#[derive(Debug, Deserialize)]
pub struct Config {
    file_types: HashMap<String, Vec<String>>,
}

impl Config {
    fn new() -> Result<Self> {
        Ok(Self {
            file_types: HashMap::new(),
        })
    }

    pub fn init() -> Result<Self> {
        const CONFIG_FILE_NAME: &str = ".desktop_cleaner.toml";
        if let Some(project_dirs) =
            ProjectDirs::from("com", "prometheon_technologies", "desktop_cleaner")
        {
            let config_path = project_dirs.config_dir().join(CONFIG_FILE_NAME);
            //dbg!(config_path);
            let config_file = std::fs::read_to_string(config_path);
            let config: Config = match config_file {
                Ok(file) => toml::from_str(&file),
                Err(_) => {
                    warn!("Failed to read config file, creating default config file");
                    Ok(Config {
                        file_types: HashMap::new(),
                    })
                }
            }
            .map_err(|e| Error::Generic(e.to_string()))?;
        }

        Ok(Self::new().expect("Failed to create default config file"))
    }
}
