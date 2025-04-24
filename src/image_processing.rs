use crate::models::{ImageJob, TargetFormat};
use anyhow::{Context, Result};
use image::{DynamicImage, GenericImageView, ImageFormat};
use std::{fs, path::PathBuf};

use xxhash_rust::const_xxh3::xxh3_64 as const_xxh3;
const OUTPUT_BASE_DIR: &'static str = "output"; // TODO: Move to config

pub fn process_image(job: ImageJob) -> Result<()> {
    tracing::debug!("Starting image processing for job: {}", job.job_id);

    let img = image::load_from_memory(&job.image_data)
        .with_context(|| format!("Failed to load image from memory for job {}", job.job_id))?;
    let (original_width, original_height) = img.dimensions();
    tracing::debug!("Image loaded: {}x{}", original_width, original_height);

    let output_dir = generate_output_path(&job.image_id)?;

    fs::create_dir_all(&output_dir)
        .with_context(|| format!("Failed to create output directory: {:?}", output_dir))?;

    match image::guess_format(&job.image_data) {
        Ok(format) => {
            if let Some(extension) = format.extensions_str().first() {
                let original_filename = format!("{}_original.{}", job.image_id, extension);
                let original_output_path = output_dir.join(&original_filename);

                tracing::debug!(
                    "Attempting to save original file (format: {:?}) to: {:?}",
                    format,
                    original_output_path
                );

                fs::write(&original_output_path, &job.image_data).with_context(|| {
                    format!(
                        "Failed to write original image to {:?}",
                        original_output_path
                    )
                })?;

                tracing::info!(
                    "Successfully saved original file: {:?}",
                    original_output_path
                );
            } else {
                tracing::warn!(
                        "Could not determine a file extension for the original image format: {:?}. Original file not saved.",
                        format
                    );
            }
        }
        Err(e) => {
            tracing::warn!(
                    "Could not guess format of the original image data for job {}: {}. Original file not saved.",
                    job.job_id,
                    e
                );
        }
    }

    let filename_base = job
        .image_id
        .split(|c| c == '/' || c == '\\')
        .last()
        .unwrap_or(&job.image_id); // Podstawowa nazwa pliku

    for (target_width, target_height) in &job.target_resolutions {
        // TODO: Dodać logikę skalowania (np. zachowanie proporcji, nie powiększanie)
        // Proste skalowanie:
        // let resized_img = img.resize_exact(*target_width, *target_height, imageops::FilterType::Lanczos3);
        //tracing::debug!("Resized image to {}x{}", *target_width, *target_height);
        let resized_img = img.thumbnail(*target_width, *target_height);
        let (actual_width, actual_height) = resized_img.dimensions();
        tracing::debug!(
            "Image resized from {}x{} to {}x{} (target: {}x{}) while preserving aspect ratio",
            original_width,
            original_height,
            actual_width,
            actual_height,
            target_width,
            target_height
        );

        for format in &job.target_formats {
            let file_name_prefix = format!(
                "{}_{}x{}",
                filename_base,
                target_width,
                target_height
            );

            let _ = save_image_to_format(file_name_prefix, format, &resized_img, &output_dir);
        }
    }

    //Save orginal size
    for format in &job.target_formats {
         let file_name_prefix = format!(
             "{}_orginal",
             filename_base,
         );
         let _ = save_image_to_format(file_name_prefix, format, &img, &output_dir);
    }

    Ok(())
}

fn save_image_to_format(
    file_name_prefix: String,
    format: &TargetFormat,
    resized_img: &DynamicImage,
    output_dir: &PathBuf) -> Result<()> {
    let filename = format!(
        "{}.{}",
        file_name_prefix,
        format.extension()
    );
    let output_path = output_dir.join(&filename);

    tracing::debug!("Saving image to: {:?} in format {:?}", output_path, format);

    match format {
        TargetFormat::WebP => {
            resized_img
                .save_with_format(&output_path, ImageFormat::WebP)
                .with_context(|| {
                    format!(
                        "Failed to save AVIF using 'image' crate to {:?}",
                        output_path
                    )
                })?;
        }
        TargetFormat::Avif => {
            resized_img
                .save_with_format(&output_path, ImageFormat::Avif)
                .with_context(|| {
                    format!(
                        "Failed to save AVIF using 'image' crate to {:?}",
                        output_path
                    )
                })?;
        }
        TargetFormat::Jpeg => {
            resized_img
                .save_with_format(&output_path, ImageFormat::Jpeg)
                .with_context(|| {
                    format!(
                        "Failed to save AVIF using 'image' crate to {:?}",
                        output_path
                    )
                })?;
        }
    }
    tracing::info!("Successfully saved: {:?}", output_path);
    Ok(())
}

pub fn generate_output_path(image_id: &str) -> Result<PathBuf> {
    let hash = const_xxh3(image_id.as_bytes());
    let hash_hex = format!("{:016x}", hash);

    if hash_hex.len() < 6 {
        return Err(anyhow::anyhow!("Hash too short to generate path structure"));
    }

    let first_3 = &hash_hex[0..3];
    let last_3 = &hash_hex[hash_hex.len() - 3..];

    let path = PathBuf::from(OUTPUT_BASE_DIR)
        .join(first_3)
        .join(last_3)
        .join(image_id);
    Ok(path)
}
