use dotenv::dotenv;
use std::env;

use crate::models::{Config, TargetFormat};

pub fn read_env() -> Config {
    // Check if running in Docker
    let in_docker = env::var("IN_DOCKER").unwrap_or_else(|_| "0".to_string()) == "1";

    // Load .env if not in Docker
    if !in_docker {
        dotenv().ok();
    }

    // Read environment variables
    let api_key = env::var("API_KEY").expect("API_KEY is not set");
    let api_key_header = env::var("API_KEY_HEADER").expect("API_KEY_HEADER is not set");
    let convert_to_res = env::var("CONVERT_TO_RES").expect("CONVERT_TO_RES is not set");
    let max_file_size: u32 = env::var("MAX_FILE_SIZE")
        .expect("MAX_FILE_SIZE is not set")
        .parse()
        .expect("MAX_FILE_SIZE must be a valid u32");
    let target_formats = env::var("TARGET_FORMATS").unwrap_or_else(|_| "jpg".to_string());

    // Parse resolutions
    let convert_to_res: Vec<(u32, u32)> = convert_to_res
        .split(',')
        .map(|res| {
            let parts: Vec<&str> = res.split('x').collect();
            if parts.len() != 2 {
                panic!("Invalid resolution format: {}", res);
            }
            let width = parts[0].parse().expect("Invalid width in resolution");
            let height = parts[1].parse().expect("Invalid height in resolution");
            (width, height)
        })
        .collect();

    // Parse target formats
    let target_formats: Vec<TargetFormat> = target_formats
        .split(',')
        .map(|format| TargetFormat::from_str(format).expect("Invalid target format"))
        .collect();

    Config {
        api_key,
        api_key_header,
        convert_to_res,
        max_file_size,
        target_formats,
    }
}
