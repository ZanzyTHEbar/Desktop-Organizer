use clap::{Args, Parser, Subcommand};

#[derive(Parser, Debug)]
#[clap(author, version, about)]
pub struct DesktopCleanerArgs {
    /// The directory to clean
    pub directory: Option<String>,
    /// Whether or not to include hidden files
    #[clap(long, action)]
    pub hidden: Option<bool>,
    /// Whether or not to include subdirectories
    #[clap(long, short, action)]
    pub recursive: Option<bool>,
}

/* #[derive(Subcommand, Debug)]
pub enum EntityType {
    /// Grab the directory to clean
    /// if none is provided we will use the default for your desktop
    Clean(DirectoryCommand),
}

#[derive(Args, Debug)]
#[group(required = false, multiple = false)]
pub struct DirectoryCommand {
    /// The directory to clean
    pub directory: Option<String>,
    /// Whether or not to include hidden files
    #[clap(long, short, action)]
    pub hidden: Option<bool>,
    /// Whether or not to include subdirectories
    #[clap(long, short, action)]
    pub recursive: Option<bool>,
 */
