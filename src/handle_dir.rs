use crate::prelude::*;
use std::fs::read_dir;

#[derive(Debug)]
pub struct DirEntry {
    pub path: String,
    pub file_name: String,
    pub file_type: String,
    pub is_dir: bool,
}

#[derive(Debug, Default)]
pub struct DirEntries {
    pub dir_entries: Option<Vec<DirEntry>>,
}

impl DirEntries {
    fn new() -> Self {
        Self {
            dir_entries: Some(Vec::new()),
        }
    }

    pub fn default() -> Self {
        Self::new()
    }

    pub fn print_dir_entries(&self) -> Result<()> {
        let dir_entries_str = String::try_from(W(self))?;
        println!("{}", dir_entries_str);
        Ok(())
    }
}

impl TryFrom<W<&DirEntries>> for String {
    type Error = Error;
    fn try_from(dir_entries: W<&DirEntries>) -> Result<Self> {
        let mut dir_entries_str = String::new();
        for dir_entry in &dir_entries.0.dir_entries {
            for entry in dir_entry {
                dir_entries_str.push_str(&f!(
                    "Path: {}\nFile Name: {}\nFile Type: {}\nIs Dir: {}\n",
                    entry.path,
                    entry.file_name,
                    entry.file_type,
                    entry.is_dir
                ));
            }
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
    fn new(path: String, file_name: String, file_type: String, is_dir: bool) -> Result<Self> {
        Ok(Self {
            path,
            file_name,
            file_type,
            is_dir,
        })
    }

    pub fn default() -> Result<Self> {
        Self::new("".to_string(), "".to_string(), "".to_string(), false)
    }

    pub fn get_dirs(dir: &std::path::Path) -> Result<Self> {
        if dir.is_dir() {
            let entries = read_dir(dir)?
                .map(|res| res.map(|e| e.path()))
                .collect::<std::io::Result<Vec<_>>>()?;

            for entry in entries {
                debug!("{:?}", entry);
                let file_name = entry
                    .file_name()
                    .unwrap()
                    .to_str()
                    .ok_or_else(|| Error::Generic(f!("Invalid Path {entry:?}")))?;
                let file_type = match entry.extension() {
                    Some(ext) => ext.to_str().unwrap().to_string(),
                    None => "".to_string(),
                };
                let is_dir = entry.is_dir();
                let entry: String = W(&entry).try_into()?;
                let dir_entry = Self::new(entry, file_name.to_string(), file_type, is_dir);
                let temp_default = Self::default().expect("Failed to get DirEntry");
                let temp_entry = dir_entry.as_ref().unwrap_or(&temp_default);
                temp_entry.print_dir_entry()?;
                let dir_entries = DirEntries::default();
                dir_entries
                    .dir_entries
                    .unwrap_or(Vec::new())
                    .push(dir_entry?);
            }
        }
        Ok(Self::default().expect("Failed to get DirEntry"))
    }

    pub fn print_dir_entry(&self) -> Result<()> {
        let dir_entry: String = W(self).try_into()?;
        debug!("{}", dir_entry);
        Ok(())
    }
}
