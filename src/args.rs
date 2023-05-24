use clap::{Args, Parser, Subcommand};

#[derive(Parser, Debug)]
#[clap(author, version, about)]
pub struct DesktopCleanerArgs {
    #[clap(subcommand)]
    pub subcmd: EntityType,
}

#[derive(Subcommand, Debug)]
pub enum EntityType {

}