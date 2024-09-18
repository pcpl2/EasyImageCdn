use serde::Deserialize;

use actix_multipart::form::{
        tempfile::TempFile,
        MultipartForm,
    };

#[derive(Deserialize)]
pub struct ImagePayload {
    pub id: String,
    pub image: String
}

#[derive(Debug, MultipartForm)]
pub struct ImageMultipartPayload {
    #[multipart(rename = "imageFile")]
    pub files: Vec<TempFile>,
}