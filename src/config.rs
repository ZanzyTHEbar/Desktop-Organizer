use crate::prelude::*;
use std::collections::HashMap;

const CONFIG_FILE_NAME: &str = "config.toml";

#[derive(Debug)]
pub struct Config {
    pub config: HashMap<String, String>,
}

impl Config {
    pub fn new(config: HashMap<String, String>) -> Self {
        Self { config }
    }

    pub fn get_config() -> Result<Self> {
        let config = std::fs::read_to_string(CONFIG_FILE_NAME)?;
        let config: HashMap<String, String> = toml::from_str(&config).unwrap();
        Ok(Self::new(config))
    }

    pub fn get_config_value(&self, key: &str) -> Option<&String> {
        self.config.get(key)
    }

    pub fn set_config_value(&mut self, key: &str, value: &str) {
        self.config.insert(key.to_string(), value.to_string());
    }

    pub fn save_config(&self) -> Result<()> {
        let config = toml::to_string(&self.config).unwrap();
        std::fs::write(CONFIG_FILE_NAME, config)?;
        Ok(())
    }

    pub fn print_config(&self) {
        println!("{:#?}", self.config);
    }

    pub fn print_config_value(&self, key: &str) {
        println!("{:#?}", self.config.get(key));
    }

    pub fn print_config_value_str(&self, key: &str) {
        println!("{:#?}", self.config.get(key).unwrap());
    }

    pub fn print_config_value_bool(&self, key: &str) {
        println!("{:#?}", self.config.get(key).unwrap().parse::<bool>());
    }

    pub fn print_config_value_i32(&self, key: &str) {
        println!("{:#?}", self.config.get(key).unwrap().parse::<i32>());
    }

    pub fn print_config_value_u32(&self, key: &str) {
        println!("{:#?}", self.config.get(key).unwrap().parse::<u32>());
    }

    pub fn print_config_value_i64(&self, key: &str) {
        println!("{:#?}", self.config.get(key).unwrap().parse::<i64>());
    }
}
