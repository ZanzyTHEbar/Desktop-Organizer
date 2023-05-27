//! Main crate Error

#[derive(thiserror::Error, Debug)]
pub enum Error {
    #[error("Generic error: {0}")]
    Generic(String),
    #[error(transparent)]
    IO(#[from] std::io::Error),
    #[error("Operation Canceled error: {0}")]
    OperationCancelled(String),
}
