//! Crate Prelude

pub use crate::cli::colors::Color;
pub use crate::error::Error;
pub use crate::logger::*;

pub type Result<T> = core::result::Result<T, Error>;

impl Error {
    pub fn new(message: String) -> Self {
        Self::Generic(message)
    }

    pub fn op_cancelled(message: &str) -> Self {
        Self::OperationCancelled(message.to_string())
    }
}
// Generic wrapper tuple struct for new type pattern , mostly used for type aliasing
pub struct W<T>(pub T);

pub use std::format as f;
