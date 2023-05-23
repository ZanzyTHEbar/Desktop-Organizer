use crate::prelude::*;
use std::fs::read_dir;

#[derive(Debug)]
pub struct DirEntry {
    pub path: String,
    pub file_name: String,
    pub file_type: String,
    pub is_dir: bool,
}

pub struct DirEntries {
    pub dir_entries: Vec<DirEntry>,
}

impl DirEntries {
    fn new() -> Self {
        Self {
            dir_entries: Vec::new(),
        }
    }

    pub fn default() -> Self {
        Self::new()
    }
}

impl TryFrom<W<&DirEntries>> for String {
    type Error = Error;
    fn try_from(dir_entries: W<&DirEntries>) -> Result<Self> {
        let mut dir_entries_str = String::new();
        for dir_entry in &dir_entries.0.dir_entries {
            dir_entries_str.push_str(&f!(
                "Path: {}\nFile Name: {}\nFile Type: {}\nIs Dir: {}\n",
                dir_entry.path,
                dir_entry.file_name,
                dir_entry.file_type,
                dir_entry.is_dir
            ));
        }
        Ok(dir_entries_str)
    }
}

impl TryFrom<W<&DirEntry>> for String {
    type Error = Error;

    fn try_from(dir_entry: W<&DirEntry>) -> Result<Self> {
        let dir_entry_str = f!(
            "Path: {}\nFile Name: {}\nFile Type: {}\nIs Dir: {}\n",
            dir_entry.0.path,
            dir_entry.0.file_name,
            dir_entry.0.file_type,
            dir_entry.0.is_dir
        );
        Ok(dir_entry_str)
    }
}

impl DirEntry {
    fn new(path: String, file_name: String, file_type: String, is_dir: bool) -> Self {
        Self {
            path,
            file_name,
            file_type,
            is_dir,
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
            let dir_entry = Self::new(entry, file_name.to_string(), file_type, is_dir);
            dir_entry.print_dir_entry()?;
        }
        Ok(())
    }

    pub fn print_dir_entry(&self) -> Result<()> {
        let dir_entry: String = W(self).try_into()?;
        println!("{}", dir_entry);
        Ok(())
    }
}
