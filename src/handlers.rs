use actix_files as fs;
use actix_multipart::Multipart;
use actix_web::http::header::{self, HeaderValue};
use actix_web::{error::ErrorNotFound, Result};
use actix_web::{web, Error as ActixError, HttpRequest, HttpResponse, Responder};
use base64::{engine::general_purpose::STANDARD as base64_standard, Engine as _};
use futures_util::stream::TryStreamExt;
use mime::Mime;
use qstring::QString;
use std::path::Path;
use std::str::FromStr;
use std::sync::Arc;
use std::time::Duration;
use uuid::Uuid;

use actix::{Actor, ActorContext, AsyncContext};

use actix_web_actors::ws;
use bytes::Bytes;
use futures_util::stream::unfold;
use tokio::sync::mpsc as TokioMpsc;

use crate::errors::AppError;
use crate::image_processing::generate_output_path;
use crate::models::{
    AppState, Config, ImageIdQuery, ImageJob, JobQueuedResponse, JobState, JobStatus,
    NewImageRequest, TargetFormat,
};

fn verify_apikey(req: &HttpRequest, config: Arc<Config>) -> Result<(), HttpResponse> {
    if let Some(apikey) = req.headers().get(config.api_key_header.as_str()) {
        if apikey == config.api_key.as_str() {
            return Ok(());
        } else {
            return Err(HttpResponse::Unauthorized().body("Invalid API Key"));
        }
    }
    Err(HttpResponse::BadRequest().body("Missing API Key"))
}

pub fn escape_file_identifier(identifier: &str) -> Option<String> {
    if identifier.is_empty() {
        return None;
    }

    let sanitized: String = identifier
        .chars()
        .filter(|&c| c.is_alphanumeric() || c == '-' || c == '_')
        .collect();

    if sanitized.is_empty() {
        return None;
    }

    let path = Path::new(&sanitized);
    if path
        .components()
        .any(|comp| matches!(comp, std::path::Component::ParentDir))
    {
        return None;
    }

    Some(sanitized)
}

pub async fn new_image_json(
    req: HttpRequest,
    state: web::Data<AppState>,
    payload: web::Json<NewImageRequest>,
) -> Result<impl Responder, AppError> {
    if let Err(error_response) = verify_apikey(&req, state.config.clone()) {
        return Ok(error_response);
    }

    tracing::info!("Received new image request via JSON for id: {}", payload.id);
    let image_data = base64_standard.decode(&payload.image)?;
    let job_id = Uuid::new_v4();

    let image_id = escape_file_identifier(&payload.id);
    if image_id.is_none() {
        return Ok(HttpResponse::BadRequest().body("Invalid image id"));
    }

    let job = ImageJob {
        job_id,
        image_id: image_id.unwrap(),
        image_data,
        target_resolutions: state.config.convert_to_res.clone(),
        target_formats: state.config.target_formats.clone(),
        avif_quality: state.config.avif_encode_parameters.clone(),
    };

    state.job_statuses.insert(
        job_id,
        JobState {
            status: JobStatus::Queued,
        },
    );
    // -----------------------------------------

    state.job_sender.send(job).await?;
    tracing::info!("Job {} queued for image id {}", job_id, payload.id);
    Ok(HttpResponse::Accepted().json(JobQueuedResponse {
        job_id,
        image_id: payload.id.clone(),
        status: "Queued".to_string(), //TODO Change to JobStatus::Queued
    }))
}

pub async fn new_image_multipart(
    req: HttpRequest,
    state: web::Data<AppState>,
    query: web::Query<ImageIdQuery>,
    mut payload: Multipart,
) -> Result<impl Responder, AppError> {
    if let Err(error_response) = verify_apikey(&req, state.config.clone()) {
        return Ok(error_response);
    }
    let mut image_data: Option<Vec<u8>> = None;

    let image_id = escape_file_identifier(query.into_inner().image_id.as_str());
    if image_id.is_none() {
        return Ok(HttpResponse::BadRequest().body("Invalid image id"));
    }

    tracing::info!(
        "Received new image request via Multipart for id: {}",
        image_id.clone().unwrap()
    );

    while let Some(mut field) = payload.try_next().await? {
        let disposition_opt = field.content_disposition();
        let field_name = disposition_opt
            .as_ref()
            .and_then(|d| d.get_name())
            .unwrap_or("");
        if field_name == "imageFile" {
            let mut field_data = Vec::new();
            while let Some(chunk) = field.try_next().await? {
                field_data.extend_from_slice(&chunk);
            }
            image_data = Some(field_data);
        } else {
            tracing::warn!("Ignoring unexpected multipart field: {}", field_name);
            while field.try_next().await?.is_some() {}
        }
    }
    let image_data =
        image_data.ok_or_else(|| AppError::BadRequest("Missing 'imageFile' field".to_string()))?;
    // ------------------------------------------------------

    let job_id = Uuid::new_v4();
    let job = ImageJob {
        job_id,
        image_id: image_id.clone().unwrap(),
        image_data,
        target_resolutions: state.config.convert_to_res.clone(),
        target_formats: state.config.target_formats.clone(),
        avif_quality: state.config.avif_encode_parameters.clone(),
    };

    state.job_statuses.insert(
        job_id,
        JobState {
            status: JobStatus::Queued,
        },
    );
    // -----------------------------------------

    state.job_sender.send(job).await?;
    tracing::info!(
        "Job {} queued for image id {}",
        job_id,
        image_id.clone().unwrap()
    );
    Ok(HttpResponse::Accepted().json(JobQueuedResponse {
        job_id,
        image_id: image_id.unwrap(),
        status: "Queued".to_string(),
    }))
}

pub async fn get_job_status(
    req: HttpRequest,
    state: web::Data<AppState>,
    path: web::Path<Uuid>,
) -> Result<impl Responder, AppError> {
    if let Err(error_response) = verify_apikey(&req, state.config.clone()) {
        return Ok(error_response);
    }

    let job_id = path.into_inner();
    tracing::debug!("Checking status for job: {}", job_id);

    match state.job_statuses.get(&job_id) {
        Some(job_state_entry) => {
            let job_state = job_state_entry.value().clone();
            tracing::info!("Status for job {}: {:?}", job_id, job_state.status);
            Ok(HttpResponse::Ok().json(job_state))
        }
        None => {
            tracing::warn!("Job not found: {}", job_id);
            Err(AppError::BadRequest(format!(
                "Job with id {} not found",
                job_id
            )))
        }
    }
}

struct JobStatusWs {
    job_id: Uuid,
    state: web::Data<AppState>,
}

impl JobStatusWs {
    fn check_status(&self, ctx: &mut ws::WebsocketContext<Self>) {
        match self.state.job_statuses.get(&self.job_id) {
            Some(entry) => {
                let state = entry.value().clone();
                let status_json = serde_json::to_string(&state).unwrap_or_else(|e| {
                    tracing::error!("Failed to serialize job status for {}: {}", self.job_id, e);
                    format!("{{\"error\":\"Failed to serialize status: {}\"}}", e)
                });

                ctx.text(status_json);

                match state.status {
                    JobStatus::Completed | JobStatus::Failed(_) => {
                        tracing::info!(
                            "Job {} finished ({:?}). Closing WebSocket.",
                            self.job_id,
                            state.status
                        );
                        ctx.stop();
                    }
                    JobStatus::Queued | JobStatus::Processing => {}
                }
            }
            None => {
                tracing::warn!(
                    "Job {} not found during WS check. Closing WebSocket.",
                    self.job_id
                );
                let error_msg = format!("{{\"error\":\"Job with id {} not found\"}}", self.job_id);
                ctx.text(error_msg);
                ctx.stop();
            }
        }
    }
}

impl Actor for JobStatusWs {
    type Context = ws::WebsocketContext<Self>;
    fn started(&mut self, ctx: &mut Self::Context) {
        tracing::info!("WebSocket connection started for job: {}", self.job_id);

        ctx.run_interval(Duration::from_secs(1), |act, ctx_interval| {
            act.check_status(ctx_interval);
        });
        self.check_status(ctx);
    }

    fn stopped(&mut self, _ctx: &mut Self::Context) {
        tracing::info!("WebSocket connection stopped for job: {}", self.job_id);
    }
}

impl actix::StreamHandler<Result<ws::Message, ws::ProtocolError>> for JobStatusWs {
    fn handle(&mut self, msg: Result<ws::Message, ws::ProtocolError>, ctx: &mut Self::Context) {
        match msg {
            Ok(ws::Message::Ping(msg)) => ctx.pong(&msg),
            Ok(ws::Message::Text(text)) => {
                tracing::debug!(
                    "Received WS message from client for job {}: {}",
                    self.job_id,
                    text
                );
                ctx.text(format!("{{\"message\":\"Received your text: {}\"}}", text));
            }
            Ok(ws::Message::Close(reason)) => {
                tracing::info!(
                    "Client closed WebSocket for job {}: {:?}",
                    self.job_id,
                    reason
                );
                ctx.close(reason);
                ctx.stop();
            }
            _ => (),
        }
    }
}

pub async fn websocket_job_status(
    req: HttpRequest,
    stream: web::Payload,
    state: web::Data<AppState>,
    path: web::Path<Uuid>,
) -> Result<HttpResponse, ActixError> {
    if let Err(error_response) = verify_apikey(&req, state.config.clone()) {
        return Ok(error_response);
    }

    let job_id = path.into_inner();
    tracing::info!("Initiating WebSocket connection for job: {}", job_id);
    if !state.job_statuses.contains_key(&job_id) {
        tracing::error!(
            "Attempted WebSocket connection for non-existent job: {}",
            job_id
        );
        return Ok(HttpResponse::NotFound().body(format!("Job with id {} not found", job_id)));
    }

    ws::start(
        JobStatusWs {
            job_id,
            state: state.clone(),
        },
        &req,
        stream,
    )
}

pub async fn sse_job_status(
    req: HttpRequest,
    state: web::Data<AppState>,
    path: web::Path<Uuid>,
) -> Result<HttpResponse, ActixError> {
    if let Err(error_response) = verify_apikey(&req, state.config.clone()) {
        return Ok(error_response);
    }

    let job_id = path.into_inner();
    tracing::info!("Initiating SSE connection for job: {}", job_id);

    if !state.job_statuses.contains_key(&job_id) {
        tracing::error!("Attempted SSE connection for non-existent job: {}", job_id);
        return Ok(HttpResponse::NotFound().body(format!("Job with id {} not found", job_id)));
    }

    let (tx, rx) = TokioMpsc::channel::<JobState>(10);
    let shared_statuses = state.job_statuses.clone();

    tokio::spawn(async move {
        let mut last_status_sent: Option<JobStatus> = None;

        loop {
            let current_state = match shared_statuses.get(&job_id) {
                Some(entry) => entry.value().clone(),
                None => {
                    let error_state = JobState {
                        status: JobStatus::Failed(format!(
                            "Job {} disappeared unexpectedly",
                            job_id
                        )),
                    };
                    let _ = tx.send(error_state).await;
                    tracing::error!(
                        "Job {} disappeared from state map during SSE monitoring.",
                        job_id
                    );
                    break;
                }
            };

            if last_status_sent.as_ref() != Some(&current_state.status) {
                if tx.send(current_state.clone()).await.is_err() {
                    tracing::info!(
                        "SSE receiver closed for job {}. Stopping monitor task.",
                        job_id
                    );
                    break;
                }
                last_status_sent = Some(current_state.status.clone());
                tracing::debug!(
                    "Sent status update via SSE for job {}: {:?}",
                    job_id,
                    current_state.status
                );
            }

            if matches!(
                current_state.status,
                JobStatus::Completed | JobStatus::Failed(_)
            ) {
                tracing::info!(
                    "Job {} finished ({:?}). Stopping SSE monitor task.",
                    job_id,
                    current_state.status
                );
                break;
            }

            tokio::time::sleep(Duration::from_secs(1)).await;
        }
    });

    let initial_state = (rx, job_id);

    let body = unfold(initial_state, |(mut rx, job_id)| async move {
        rx.recv().await.map(|received_state| {
            let json = serde_json::to_string(&received_state).unwrap_or_else(|e| {
                tracing::error!("SSE serialization error for job {}: {}", job_id, e);
                "{\"error\":\"Serialization failed\"}".to_string()
            });
            let event_bytes = Bytes::from(format!("data: {}\n\n", json));
            (Ok::<_, ActixError>(event_bytes), (rx, job_id))
        })
    });

    Ok(HttpResponse::Ok()
        .content_type("text/event-stream")
        .insert_header(("Cache-Control", "no-cache"))
        .insert_header(("Connection", "keep-alive"))
        .streaming(body))
}

fn get_best_image_extension(
    accept: &str,
    force: &str,
    target_formats: Vec<TargetFormat>,
) -> (&'static str, &'static str) {
    let active_formats: Vec<TargetFormat> = target_formats.clone();
    let preferred_formats = ["image/avif", "image/webp", "image/jpeg"];
    if force != "" {
        return match TargetFormat::from_str(force) {
            Ok(format) if active_formats.contains(&format) => (
                format.extension(),
                match format {
                    TargetFormat::Avif => "image/avif",
                    TargetFormat::WebP => "image/webp",
                    TargetFormat::Jpeg => "image/jpeg",
                },
            ),
            _ => ("jpg", "image/jpeg"),
        };
    }

    let accept_parts: Vec<&str> = accept
        .split(',')
        .map(|part| part.trim().split(';').next().unwrap_or(""))
        .collect();

    for &preferred in &preferred_formats {
        if accept_parts.contains(&preferred) {
            let target_format = match preferred {
                "image/avif" => TargetFormat::Avif,
                "image/webp" => TargetFormat::WebP,
                "image/jpeg" => TargetFormat::Jpeg,
                _ => TargetFormat::Jpeg,
            };
            if active_formats.contains(&target_format) {
                return (
                    target_format.extension(),
                    match target_format {
                        TargetFormat::Avif => "image/avif",
                        TargetFormat::WebP => "image/webp",
                        TargetFormat::Jpeg => "image/jpeg",
                    },
                );
            }
        }
    }
    ("jpg", "image/jpeg")
}

fn parse_resolution(resolution_str: &str) -> Option<(u32, u32)> {
    if resolution_str.contains("..")
        || resolution_str.contains('/')
        || resolution_str.contains('\\')
        || resolution_str == ""
    {
        return None;
    }

    let parts: Vec<&str> = resolution_str.split('x').collect();

    if parts.len() != 2 {
        return None;
    }

    let x = parts[0].parse::<u32>().ok()?;
    let y = parts[1].parse::<u32>().ok()?;

    Some((x, y))
}

pub async fn get_file(
    req: HttpRequest,
    config: web::Data<Arc<Config>>,
) -> Result<HttpResponse, ActixError> {
    //TODO Get from cache
    //TODO Validate referer
    let qs = QString::from(req.query_string());
    let image_id = escape_file_identifier(req.match_info().query("image_id"));
    if image_id.is_none() {
        return Ok(HttpResponse::BadRequest().body("Invalid image id"));
    }
    let resolution_str = req.match_info().query("resolution");
    let accept = req
        .headers()
        .get("accept")
        .and_then(|val| val.to_str().ok())
        .unwrap_or("");
    let force_type = qs.get("extension").unwrap_or_default();
    let selected_ext = get_best_image_extension(accept, force_type, config.target_formats.clone());

    //let cache_key = format!("{}_{}.{}", image_id, resolution_str, selected_ext.0);

    let name = match parse_resolution(resolution_str) {
        Some((x, y)) => format!(
            "{}_{}x{}.{}",
            image_id.clone().unwrap(),
            x,
            y,
            selected_ext.0
        ),
        None => format!("{}_orginal.{}", image_id.clone().unwrap(), selected_ext.0),
    };

    let base_path =
        generate_output_path(image_id.clone().unwrap().as_str()).map_err(ErrorNotFound)?;
    let output_path = base_path.join(name);

    //tracing::info!("Get file: {:?}", output_path,);

    let file = fs::NamedFile::open(&output_path).map_err(|_| ErrorNotFound("File not found"))?;

    let mut response = file
        .use_etag(true)
        .use_last_modified(true)
        .set_content_type(Mime::from_str(selected_ext.1).unwrap())
        .disable_content_disposition()
        .into_response(&req);

    response.headers_mut().append(header::CACHE_CONTROL, HeaderValue::from_str(config.cache_control_header.clone().as_str()).unwrap());
    Ok(response)
}
