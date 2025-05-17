use dashmap::DashMap;
use serde::{Deserialize, Serialize};
use std::sync::Arc;
use tokio::sync::mpsc;
use uuid::Uuid;

#[derive(Deserialize, Debug)]
pub struct NewImageRequest {
    pub id: String,
    pub image: String,
}

#[derive(Deserialize, Debug)]
pub struct ImageIdQuery {
    #[serde(rename = "imageId")]
    pub image_id: String,
}

#[derive(Debug, Clone, PartialEq, Eq, Hash, Serialize, Deserialize)]
pub enum TargetFormat {
    WebP,
    Avif,
    Jpeg,
}

impl TargetFormat {
    pub fn extension(&self) -> &'static str {
        match self {
            TargetFormat::WebP => "webp",
            TargetFormat::Avif => "avif",
            TargetFormat::Jpeg => "jpg",
        }
    }

    pub fn from_str(s: &str) -> Result<Self, String> {
        match s.to_lowercase().as_str() {
            "webp" => Ok(TargetFormat::WebP),
            "avif" => Ok(TargetFormat::Avif),
            "jpg" | "jpeg" => Ok(TargetFormat::Jpeg),
            _ => Err(format!("Invalid target format: {}", s)),
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub enum JobStatus {
    Queued,
    Processing,
    Completed,
    Failed(String),
}

#[derive(Debug, Clone, Serialize)]
pub struct JobState {
    pub status: JobStatus,
}

#[derive(Debug)]
pub struct ImageJob {
    pub job_id: Uuid,
    pub image_id: String,
    pub image_data: Vec<u8>,
    //pub original_filename: Option<String>,
    // TODO: move to configuration
    pub target_resolutions: Vec<(u32, u32)>,
    pub target_formats: Vec<TargetFormat>,
}

#[derive(Serialize)]
pub struct JobQueuedResponse {
    pub job_id: Uuid,
    pub image_id: String,
    pub status: String,
}

#[derive(Clone)]
pub struct AppState {
    pub job_sender: mpsc::Sender<ImageJob>,
    pub job_statuses: Arc<DashMap<Uuid, JobState>>,
    pub config: Arc<Config>,
}

#[derive(Debug)]
pub struct Config {
    pub api_key: String,
    pub api_key_header: String,
    pub convert_to_res: Vec<(u32, u32)>,
    pub max_file_size: u32,
    pub target_formats: Vec<TargetFormat>,
    pub cache_control_header: String,
}
