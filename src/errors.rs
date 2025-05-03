use actix_web::{http::StatusCode, HttpResponse, ResponseError};
use derive_more::Display;

#[derive(Debug, Display)]
pub enum AppError {
    #[display("Bad Request: {}", _0)]
    BadRequest(String),
    #[display("Internal Server Error: {}", _0)]
    InternalError(String),
    #[display("Failed to queue job: {}", _0)]
    QueueError(String),
    // #[display("Image processing error: {}", _0)]
    // ImageProcessingError(String),
    #[display("Multipart error: {}", _0)]
    MultipartError(String),
}

impl ResponseError for AppError {
    fn status_code(&self) -> StatusCode {
        match *self {
            AppError::BadRequest(_) => StatusCode::BAD_REQUEST,
            AppError::InternalError(_) => StatusCode::INTERNAL_SERVER_ERROR,
            AppError::QueueError(_) => StatusCode::INTERNAL_SERVER_ERROR,
            //AppError::ImageProcessingError(_) => StatusCode::INTERNAL_SERVER_ERROR,
            AppError::MultipartError(_) => StatusCode::BAD_REQUEST,
        }
    }

    fn error_response(&self) -> HttpResponse {
        HttpResponse::build(self.status_code())
            .json(serde_json::json!({ "error": self.to_string() }))
    }
}

impl From<base64::DecodeError> for AppError {
    fn from(err: base64::DecodeError) -> Self {
        AppError::BadRequest(format!("Invalid base64 data: {}", err))
    }
}

impl From<tokio::sync::mpsc::error::SendError<crate::models::ImageJob>> for AppError {
    fn from(err: tokio::sync::mpsc::error::SendError<crate::models::ImageJob>) -> Self {
        AppError::QueueError(format!("Failed to send job to queue: {}", err))
    }
}

impl From<actix_multipart::MultipartError> for AppError {
    fn from(err: actix_multipart::MultipartError) -> Self {
        AppError::MultipartError(format!("Multipart stream error: {}", err))
    }
}

impl From<std::io::Error> for AppError {
    fn from(err: std::io::Error) -> Self {
        AppError::InternalError(format!("IO Error: {}", err))
    }
}

impl From<anyhow::Error> for AppError {
    fn from(err: anyhow::Error) -> Self {
        AppError::InternalError(format!("Processing failed: {}", err))
    }
}
