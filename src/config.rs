use crate::prelude::*;
use directories::ProjectDirs;
use opzioni::Config;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::fs::OpenOptions;
use std::io::prelude::*;

const QUALIFIERS: [&str; 3] = ["com", "prometheon_technologies", "desktop_cleaner"];
const CONFIG_FILE: &str = ".desktop_cleaner.toml";

#[derive(Debug, Deserialize, Serialize, Default)]
pub struct DesktopCleanerConfig {
    pub file_types: Option<HashMap<String, Vec<String>>>,
    pub debug: Option<HashMap<String, String>>,
}

impl DesktopCleanerConfig {
    fn new() -> Result<Self> {
        Ok(Self {
            file_types: Some(HashMap::new()),
            debug: Some(HashMap::new()),
        })
    }

    pub fn init() -> Result<Self> {
        if let Some(project_dirs) = ProjectDirs::from(QUALIFIERS[0], QUALIFIERS[1], QUALIFIERS[2]) {
            Self::create_config_dir()?;
            let config_path = project_dirs.config_dir().join(CONFIG_FILE);
            let config_file: Config<DesktopCleanerConfig> =
                Config::<DesktopCleanerConfig>::configure()
                    .load(&config_path)
                    .unwrap_or(Config::<DesktopCleanerConfig>::empty());
            let lock = config_file.get();
            let data = lock.read().unwrap();
            return Ok(Self {
                file_types: data.file_types.clone(),
                debug: data.debug.clone(),
            });
        }
        Ok(Self::new().expect("Failed to load DesktopCleaner Config file"))
    }

    pub fn map_debug_level(&self) -> log::LevelFilter {
        let debug_level = match self
            .debug
            .as_ref()
            .unwrap()
            .get("level")
            .map(|level| level.as_str())
            .unwrap_or("info")
        {
            "trace" => log::LevelFilter::Trace,
            "debug" => log::LevelFilter::Debug,
            "info" => log::LevelFilter::Info,
            "warn" => log::LevelFilter::Warn,
            "error" => log::LevelFilter::Error,
            "off" => log::LevelFilter::Off,
            _ => log::LevelFilter::Debug,
        };
        debug_level
    }

    fn create_config_dir() -> Result<()> {
        if let Some(project_dirs) = ProjectDirs::from(QUALIFIERS[0], QUALIFIERS[1], QUALIFIERS[2]) {
            let config_path = project_dirs.config_dir();
            if !config_path.exists() {
                eprintln!(
                    "Config Directory Doesn't Exist - Creating it: {:?}",
                    config_path
                );
                std::fs::create_dir_all(config_path)?;

                let config_path = project_dirs.config_dir().join(CONFIG_FILE);

                eprintln!("Config File Path: {:?}", config_path);

                if config_path.exists() {
                    return Ok(());
                }

                eprintln!("Config File Doesn't Exist - Creating it: {:?}", config_path);

                let mut file = OpenOptions::new()
                    .write(true)
                    .create(true)
                    .open(&config_path)?;

                eprintln!("Config File Being Generated: {:?}", config_path);

                let config = indoc::indoc! {"
                [file_types]
                
                [debug]
                "};

                if let Err(e) = writeln!(file, "{}", config) {
                    return Err(e.into());
                }

                eprintln!(
                    "Created config file at: {:?}",
                    config_path.join(CONFIG_FILE)
                );

                return Ok(());
            }
            let config_path = project_dirs.config_dir().join(CONFIG_FILE);
            eprintln!("Config Exists: {:?}", config_path);
        }
        Ok(())
    }
}
