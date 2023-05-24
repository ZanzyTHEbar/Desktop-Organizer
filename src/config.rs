use crate::prelude::*;
use directories::ProjectDirs;
use log::{debug, info, warn};
use opzioni::Config;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;

#[derive(Debug, Deserialize, Serialize, Default)]
pub struct DesktopCleanerConfig {
    pub file_types: HashMap<String, Vec<String>>,
}

impl DesktopCleanerConfig {
    fn new() -> Result<Self> {
        Ok(Self {
            file_types: HashMap::new(),
        })
    }

    pub fn init() -> Result<Self> {
        const CONFIG_FILE: &str = ".desktop_cleaner.toml";
        if let Some(project_dirs) =
            ProjectDirs::from("com", "prometheon_technologies", "desktop_cleaner")
        {
            let config_path = project_dirs.config_dir().join(CONFIG_FILE);
            let config_file: Config<DesktopCleanerConfig> =
                Config::<DesktopCleanerConfig>::configure()
                    .load(&config_path)
                    .unwrap_or(Config::<DesktopCleanerConfig>::empty());
            let lock = config_file.get();
            let data = lock.read().unwrap();
            debug!("Config file: {:?}", config_path);
            //debug!("Config: {:?}", data.file_types);
            let file_types = data.file_types.clone();
            return Ok(Self { file_types });
            // save the config
            //config_file.save().unwrap();

            // read the config
            //let data = lock.read().unwrap();
            //debug!("Config: {:?}", data.file_types);

            //dbg!(DesktopCleanerConfig_path);
            /* let DesktopCleanerConfig_file = std::fs::read_to_string(DesktopCleanerConfig_path);
            let DesktopCleanerConfig: DesktopCleanerConfig = match DesktopCleanerConfig_file {
                Ok(file) => toml::from_str(&file),
                Err(_) => {
                    warn!("Failed to read DesktopCleanerConfig file, creating default DesktopCleanerConfig file");
                    Ok(DesktopCleanerConfig {
                        file_types: HashMap::new(),
                    })
                }
            }
            .map_err(|e| Error::Generic(e.to_string()))?; */
        }

        Ok(Self::new().expect("Failed to create default DesktopCleanerConfig file"))
    }
}
