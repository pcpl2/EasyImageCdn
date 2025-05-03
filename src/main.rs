use actix_web::{middleware::Logger, web, App, HttpServer};
use config::read_env;
use dashmap::DashMap;
use futures_util::future::try_join;
use std::sync::Arc;
use tokio::sync::mpsc;
use uuid::Uuid;

mod config;
mod errors;
mod handlers;
mod image_processing;
mod models;
mod worker;

use models::{AppState, ImageJob, JobState};

const IMAGE_SERVER_PORT: u16 = 9555;
const ADMIN_SERVER_PORT: u16 = 9324;

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    let config = read_env();
    let my_filter = EnvFilter::builder()
        .with_default_directive(LevelFilter::INFO.into())
        .from_env_lossy();
    tracing_subscriber::registry()
        .with(fmt::layer())
        .with(my_filter)
        .init();

    // TODO: Buffer size add to config
    let (job_sender, job_receiver) = mpsc::channel::<ImageJob>(100);

    let job_statuses = Arc::new(DashMap::<Uuid, JobState>::new());
    let worker_statuses = job_statuses.clone();

    tokio::spawn(worker::run_worker(job_receiver, worker_statuses));

    let app_state = web::Data::new(AppState {
        job_sender: job_sender.clone(),
        job_statuses: job_statuses.clone(),
        config: Arc::new(config),
    });

    let main_server = HttpServer::new(move || {
        App::new()
            .app_data(app_state.clone())
            .app_data(web::PayloadConfig::new((app_state.config.max_file_size.clone() as usize) * 1024 * 1024))
            .wrap(Logger::default())
            .service(
                web::scope("/v1")
                    .route("/newImage", web::post().to(handlers::new_image_json))
                    .route("/newImageMp", web::post().to(handlers::new_image_multipart))
                    .route(
                        "/job/{job_id}/status",
                        web::get().to(handlers::get_job_status),
                    )
                    .route(
                        "/job/{job_id}/ws",
                        web::get().to(handlers::websocket_job_status),
                    )
                    .route("/job/{job_id}/sse", web::get().to(handlers::sse_job_status)),
            )
    })
    .bind(("0.0.0.0", ADMIN_SERVER_PORT))?
    .run();

    let image_server = HttpServer::new(move || {
        App::new()
        .route("/{image_id}", web::get().to(handlers::get_file))
        .route("/{image_id}/", web::get().to(handlers::get_file))
        .route(
              "/{image_id}/{resolution}",
             web::get().to(handlers::get_file),
          )
    })
    .bind(("0.0.0.0", IMAGE_SERVER_PORT))?
    .run();

    tracing::info!(
        "Starting main API server on http://localhost:{}",
        ADMIN_SERVER_PORT
    );
    tracing::info!(
        "Starting image server on http://localhost:{}",
        IMAGE_SERVER_PORT
    );

    match try_join(main_server, image_server).await {
        Ok((main_res, image_res)) => {
            tracing::info!("Main server finished: {:?}", main_res);
            tracing::info!("Image server finished: {:?}", image_res);
            Ok(())
        }
        Err(e) => {
            tracing::error!("A server encountered an error: {}", e);
            Err(e)
        }
    }
}
