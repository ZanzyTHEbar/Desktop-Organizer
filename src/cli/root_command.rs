use clap::{Parser, ValueHint};
use std::path::PathBuf;
/// A simple CLI utility to organize your desktop (or a specified directory)
///
/// EXAMPLES:
/// Automated Desktop Cleaning:
/// $ desktop_cleaner
///
/// Automated Desktop Cleaning with a specified directory:
/// $ desktop_cleaner --directory /home/user/Downloads
///
/// More examples can be found at https://github.com/ZanzyTHEbar/Desktop-Cleaner#usage
#[derive(Debug, Parser)]
#[command(author, name = "desktop_cleaner", version, about, verbatim_doc_comment)]
pub struct Cli {
    #[command(subcommand)]
    pub cmd: Command,
}

#[derive(Debug, Parser)]
pub enum Command {
    /// The directory to clean
    Clean {
        #[arg(short, long, value_hint = ValueHint::AnyPath)]
        directory: Option<PathBuf>,
    },

    /// Generate shell completion
    Completion {
        /// Shell to generate completion for
        #[arg(value_enum)]
        shell: clap_complete::Shell,
    },
}
