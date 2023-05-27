//! Crate Logger
pub use crate::cli::colors::Color;
use env_logger::filter::{Builder, Filter};
use log::{LevelFilter, Log, Metadata, Record, SetLoggerError};
use std::format as f;

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
        if self.inner.matches(record) {
            println!(
                "{}",
                format_args!(
                    "{} {}",
                    Color::new(
                        f!(
                            "[Desktop Cleaner - {}]:",
                            Color::new(record.level().as_str())
                                .map_level(record.level())
                                .bold(),
                        )
                        .as_str()
                    )
                    .bold()
                    .green(),
                    Color::new(f!("{}", record.args()).as_str()).cyan()
                )
            );
        }
    }

    fn flush(&self) {}
}

#[macro_export]
macro_rules! dc_stderr {
    ($($arg:tt)+) => (eprintln!("{}", f!("{} {}", Color::new("[Desktop Cleaner]:").bold().green(), Color::new($($arg)+).red())));
}

#[macro_export]
macro_rules! dc_stdout {
    ($($arg:tt)+) => (println!("{}", f!("{} {}", Color::new("[Desktop Cleaner]:").bold().green(), Color::new($($arg)+).green())));
}

pub(crate) use dc_stderr;
pub(crate) use dc_stdout;
