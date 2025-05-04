use crate::image_processing;
use crate::models::{ImageJob, JobState, JobStatus};
use tokio::sync::mpsc;

use dashmap::DashMap;
use std::sync::Arc;
use uuid::Uuid;

// Zmieniamy sygnaturę, aby przyjmował AppState
pub async fn run_worker(
    mut receiver: mpsc::Receiver<ImageJob>,
    // Potrzebujemy dostępu do mapy statusów
    job_statuses: Arc<DashMap<Uuid, JobState>>,
) {
    tracing::info!("Image processing worker started");
    while let Some(job) = receiver.recv().await {
        let job_id = job.job_id;
        let image_id = job.image_id.clone();
        tracing::info!("Processing job: {} for image_id: {}", job_id, image_id);

        // --- ZMIANA: Aktualizuj status na Processing ---
        job_statuses.entry(job_id).and_modify(|state| {
            state.status = JobStatus::Processing;
            // state.started_at = Some(chrono::Utc::now()); // Opcjonalnie
            tracing::debug!("Job {} status updated to Processing", job_id);
        });
        // ---------------------------------------------

        let statuses_clone = job_statuses.clone();

        let result = tokio::task::spawn_blocking(move || {
            image_processing::process_image(job)
        })
        .await;

        // --- ZMIANA: Aktualizuj status po zakończeniu ---
        match result {
            Ok(Ok(())) => {
                statuses_clone.entry(job_id).and_modify(|state| {
                    state.status = JobStatus::Completed;
                    // state.finished_at = Some(chrono::Utc::now()); // Opcjonalnie
                    tracing::info!("Job {} status updated to Completed", job_id);
                });
                tracing::info!(
                    "Job {} (image_id: {}) completed successfully",
                    job_id,
                    image_id
                );
            }
            Ok(Err(e)) => {
                // Błąd zwrócony przez process_image (anyhow::Error)
                let error_msg = format!("{}", e); // Konwertuj błąd na string
                statuses_clone.entry(job_id).and_modify(|state| {
                    state.status = JobStatus::Failed(error_msg.clone()); // Klonujemy string
                                                                         // state.finished_at = Some(chrono::Utc::now()); // Opcjonalnie
                    tracing::error!("Job {} status updated to Failed", job_id);
                });
                tracing::error!(
                    "Job {} (image_id: {}) failed during processing: {}",
                    job_id,
                    image_id,
                    error_msg
                );
            }
            Err(e) => {
                // Błąd paniki w spawn_blocking
                let error_msg = format!("Task panicked: {}", e);
                statuses_clone.entry(job_id).and_modify(|state| {
                    state.status = JobStatus::Failed(error_msg.clone());
                    // state.finished_at = Some(chrono::Utc::now()); // Opcjonalnie
                    tracing::error!("Job {} status updated to Failed due to panic", job_id);
                });
                tracing::error!(
                    "Job {} (image_id: {}) panicked during processing: {}",
                    job_id,
                    image_id,
                    error_msg
                );
            }
        }
        // -------------------------------------------------
    }
    tracing::info!("Image processing worker stopped");
}
