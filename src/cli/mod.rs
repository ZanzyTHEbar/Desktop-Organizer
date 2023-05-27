pub mod colors;
pub mod completion;
pub mod clean;
pub mod root_command;

pub fn get_env(name: &str) -> Option<String> {
    std::env::var(name).ok()
}
