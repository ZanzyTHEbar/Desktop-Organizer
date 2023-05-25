use clap::Parser;

#[derive(Parser, Debug)]
#[clap(author, version, about)]
pub struct DesktopCleanerArgs {
    /// The directory to clean
    pub directory: Option<String>,
    // Whether or not to include hidden files
    //#[clap(long, action)]
    //pub hidden: Option<bool>,
    // Whether or not to include subdirectories
    //#[clap(long, short, action)]
    //pub recursive: Option<bool>,
}
