use crate::prelude::*;
use directories::ProjectDirs;
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
            return Ok(Self {
                file_types: data.file_types.clone(),
            });
        }
        Ok(Self::new().expect("Failed to load DesktopCleaner Config file"))
    }
}
