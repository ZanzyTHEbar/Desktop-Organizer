#[macro_use]
extern crate log;

use crate::prelude::*;

mod error;
mod logger;
mod prelude;
mod utils;
mod handle_dir;

fn main() -> Result<()> {
    logger::DesktopCleanerLogger::init(log::LevelFilter::max()).unwrap();

    info!("Hello, world!");
    handle_dir::DirEntry::get_dir_entry("./".to_string())?;

    Ok(())
}
