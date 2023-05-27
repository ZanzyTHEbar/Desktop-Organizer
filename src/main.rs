// NOTE: this rule is not supported by rust-analyzer or JetBrains Rust plugin go to definition/refactoring tools so disable it until it's supported properly
#![allow(clippy::uninlined_format_args)]

#[macro_use]
extern crate log;
use crate::prelude::*;
use clap::Parser;

use crate::cli::clean::CleanHandler;
use crate::cli::colors::Color;
use crate::cli::completion::CompletionHandler;
use crate::cli::root_command::{Cli, Command};

mod cli;
mod config;
mod error;
mod handle_dir;
mod logger;
mod prelude;
mod utils;

/**
 *  Logic Loop
 *  1. Get config
 *  2. Get dir entries by getting the directory to Desktop or home if no directory passed in
 *     and then iterating through the input path and creating a DirEntry struct for each entry
 *  3. Iterate through each DirEntry and move the file to the appropriate folder
 *   - If the file is a symlink, skip it
 *   - If the file is a hidden file, skip it
 *   - if the appropriate folder does not exist, create it
 *   - If the file is a directory, recursively call the function if that config option is enabled
 *   - If the file is not a directory, move the file to the appropriate folder
 *  4. Print the number of files moved
 */
fn main() -> Result<()> {
    // get args
    let cli = Cli::parse();
    init_ctrl_c_handler();
    handle(run(cli));
    Ok(())
}

// TODO: move to `walkdir` crate
// TODO: impl skipping hidden files
// TODO. impl skipping symlinks
// TODO: impl searching subdirectories
// TODO: impl creating sub directories
// TODO: impl support for other mount points  and targeting a specific destination
fn run(cli: Cli) -> Result<()> {
    match cli.cmd {
        Command::Clean {
            directory,
            //recursive,
            //hidden,
        } => CleanHandler::new(directory, None, None).run(),
        Command::Completion { shell } => CompletionHandler::new(shell).run(),
    }
}

// NOTE: this is needed to restore the cursor if CTRL+C is
// pressed during the asset selection (https://github.com/mitsuhiko/dialoguer/issues/77)
fn init_ctrl_c_handler() {
    ctrlc::set_handler(move || {
        let term = dialoguer::console::Term::stderr();
        let _ = term.show_cursor();
        std::process::exit(1);
    })
    .expect("Error initializing CTRL+C handler")
}

fn handle(result: Result<()>) {
    if let Err(error) = result {
        match error {
            Error::Generic(msg) => {
                dc_stderr!(&msg);
                std::process::exit(1)
            }
            Error::OperationCancelled(msg) => {
                dc_stdout!(f!("Operation cancelled: {}", Color::new(&msg).bold()).as_str());
            }
            Error::IO(error) => {
                dc_stderr!(&error.to_string());
                std::process::exit(1)
            }
        }
    }
}
