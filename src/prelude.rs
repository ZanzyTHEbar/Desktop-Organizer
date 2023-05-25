//! Crate Prelude

pub use crate::error::Error;

pub type Result<T> = core::result::Result<T, Error>;

// Generic wrapper tuple struct for newtype pattern , mostly used for type aliasing
pub struct W<T>(pub T);

pub use std::format as f;

