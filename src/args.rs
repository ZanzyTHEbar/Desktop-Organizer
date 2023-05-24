use clap::{Args, Parser, Subcommand};

#[derive(Parser, Debug)]
#[clap(author, version, about)]
pub struct DesktopCleanerArgs {
    #[clap(subcommand)]
    pub subcmd: EntityType,
}

#[derive(Subcommand, Debug)]
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
    //#[clap(subcommand)]
    //pub recursive: RecursiveSubCommand,
}

/* #[derive(Subcommand, Debug)]
pub enum RecursiveSubCommand {
    /// Whether or not to include subdirectories
    Recursive(RecursiveCommand),
}

#[derive(Args, Debug)]
pub struct RecursiveCommand {
    /// Enable or disable recursive cleaning
    pub recursive: bool,
} */
