use crate::prelude::*;
use std::fs::DirEntry;
use std::path::PathBuf;

impl TryFrom<W<&PathBuf>> for String {
    type Error = Error;
    fn try_from(val: W<&PathBuf>) -> Result<String> {
        val.0
            .to_str()
            .map(String::from)
            .ok_or_else(|| Error::Generic(f!("Invalid Path {:?}", val.0)))
    }
}

impl TryFrom<W<&DirEntry>> for String {
    type Error = Error;
    fn try_from(val: W<&DirEntry>) -> Result<String> {
        val.0
            .path()
            .to_str()
            .map(String::from)
            .ok_or_else(|| Error::Generic(f!("Invalid Path {:?}", val.0)))
    }
}
