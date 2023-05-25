//! Crate Logger

use env_logger::filter::{Builder, Filter};

use log::{LevelFilter, Log, Metadata, Record, SetLoggerError};

pub struct DesktopCleanerLogger {
    inner: Filter,
}

impl DesktopCleanerLogger {
    pub fn new(log_level: LevelFilter) -> DesktopCleanerLogger {
        let mut builder = Builder::new();

        //builder
        //    .filter(None, LevelFilter::Info)
        //    .filter(Some("desktop_cleaner"), LevelFilter::Debug);

        builder.filter_level(log_level);
        DesktopCleanerLogger {
            inner: builder.build(),
        }
    }

    pub fn init(log_level: LevelFilter) -> Result<(), SetLoggerError> {
        let logger = Self::new(log_level);
        log::set_max_level(logger.inner.filter());
        log::set_boxed_logger(Box::new(logger))
    }
}

impl Log for DesktopCleanerLogger {
    fn enabled(&self, metadata: &Metadata) -> bool {
        self.inner.enabled(metadata)
    }

    fn log(&self, record: &Record) {
        // Check if the record is matched by the logger before logging
        if self.inner.matches(record) {
            println!("[Desktop Cleaner - {}]: {}", record.level(), record.args());
        }
    }

    fn flush(&self) {}
}

#[macro_export]
macro_rules! dc_stderr {
    ($($arg:tt)+) => (eprintln!("[Desktop Cleaner]: {}", $($arg)+));
}

#[macro_export]
macro_rules! dc_stdout {
    ($($arg:tt)+) => (println!("[Desktop Cleaner]: {}", $($arg)+));
}

pub(crate) use dc_stderr;
pub(crate) use dc_stdout;
