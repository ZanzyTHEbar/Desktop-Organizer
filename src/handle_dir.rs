use crate::prelude::*;
use std::fs::read_dir;

#[derive(Debug)]
pub struct DirEntry {
    pub path: String,
    pub file_name: String,
    pub file_type: String,
    pub is_dir: bool,
}

impl DirEntry {
    pub fn new(
        path: String,
        file_name: String,
        file_type: String,
        is_dir: bool,
    ) -> Self {
        Self {
            path,
            file_name,
            file_type,
            is_dir,
            //dir_entries,
        }
    }

    pub fn get_dir_entry(input_path: String) -> Result<()> {
        for entry in read_dir(input_path)?.filter_map(|e| e.ok()) {
            let path = entry.path().clone();
            let file_name = path
                .file_name()
                .unwrap()
                .to_str()
                .ok_or_else(|| Error::Generic(f!("Invalid Path {path:?}")))?;

            let file_type = match path.extension() {
                Some(ext) => ext.to_str().unwrap().to_string(),
                None => "".to_string(),
            };

            let is_dir = path.is_dir();
            let entry: String = W(&entry).try_into()?;
            let dir_entry = Self::new(
                entry,
                file_name.to_string(),
                file_type,
                is_dir,
            );

            dir_entry.print_dir_entry();
        }
        Ok(())
    }

    pub fn print_dir_entry(&self) {
        println!("Path: {}", self.path);
        println!("File Name: {}", self.file_name);
        println!("File Type: {}", self.file_type);
        println!("Is Dir: {}", self.is_dir);
    }
}
