use crate::prelude::*;
use std::fs::read_dir;

#[derive(Debug)]
pub struct DirEntry {
    pub path: String,
    pub file_name: String,
    pub file_type: String,
    pub is_dir: bool,
    pub dir_entries: Option<Vec<DirEntry>>,
}

impl DirEntry {
    pub fn new(
        path: String,
        file_name: String,
        file_type: String,
        is_dir: bool,
        dir_entries: Option<Vec<DirEntry>>,
    ) -> Self {
        Self {
            path,
            file_name,
            file_type,
            is_dir,
            dir_entries,
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

            let entry = entry
                .path()
                .to_str()
                .map(String::from)
                .ok_or_else(|| Error::Generic(f!("Invalid Path {entry:?}")));

            let dir_entries = if is_dir { Some(vec![]) } else { None };

            let dir_entry = Self::new(
                entry?,
                file_name.to_string(),
                file_type,
                is_dir,
                dir_entries,
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
        println!("Dir Entries: {:?}", self.dir_entries);
    }
}

/*
let mut dir_entries_vec = Vec::new();
        for entry in dir_entries.unwrap() {
            let path = entry.path().to_str().unwrap().to_string();
            let file_name = entry.file_name().to_str().unwrap().to_string();
            let file_type = match entry.path().extension() {
                Some(ext) => ext.to_str().unwrap().to_string(),
                None => "".to_string(),
            };
            let is_dir = entry.path().is_dir();
            let dir_entries = None;
            dir_entries_vec.push(DirEntry::new(
                path,
                file_name,
                file_type,
                is_dir,
                dir_entries,
            ));
        }
        Ok(Self::new(
            path,
            file_name,
            file_type,
            is_dir,
            Some(dir_entries_vec),
        ))


 */
