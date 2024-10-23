use std::io::Cursor;

use actix_files as fs;
use actix_multipart::form::MultipartForm;
use actix_web::http::header::{ContentDisposition, DispositionType};
use actix_web::middleware::Logger;
use actix_web::{get, post, web, App, HttpResponse, HttpServer, Responder};
use actix_web::{Error, HttpRequest};
use base64::prelude::BASE64_STANDARD;
use base64::Engine;
use futures::future;
use image::ImageReader;
use xxhash_rust::const_xxh3::xxh3_64 as const_xxh3;

mod models;

#[post("/v1/newImage")]
async fn post_new_image(form: web::Json<models::rest_models::ImagePayload>) -> HttpResponse {
    let imageDecoded = BASE64_STANDARD.decode(form.image.clone());
    if imageDecoded.is_err() {
        return HttpResponse::BadRequest().body("Cannot decode image");
    }
    let img2 = ImageReader::new(Cursor::new(imageDecoded.unwrap()))
        .with_guessed_format().unwrap()
        .decode();
    let img3 = img2.unwrap().resize(100, 100, image::imageops::FilterType::Gaussian);
    img3.save_with_format("converted.avif", image::ImageFormat::Avif);

    let ID_HASH = const_xxh3(form.id.as_bytes());
    HttpResponse::Ok().body(format!("username: {:x}", ID_HASH))
}

#[post("/v1/newImageMP")]
pub async fn post_image_multipart(
    MultipartForm(form): MultipartForm<models::rest_models::ImageMultipartPayload>,
) -> impl Responder {
    format!(
        "Uploaded file {}",
        form.files.first().unwrap().file_name.as_ref().unwrap()
    )
}

#[get("/{filename:.*}")]
async fn images(req: HttpRequest) -> Result<fs::NamedFile, Error> {
    let path: std::path::PathBuf = req.match_info().query("filename").parse().unwrap();
    let file = fs::NamedFile::open(path)?;
    Ok(file
        .use_last_modified(true)
        .set_content_disposition(ContentDisposition {
            disposition: DispositionType::Attachment,
            parameters: vec![],
        }))
}

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    let admin_api = HttpServer::new(|| {
        App::new()
            .wrap(Logger::default())
            .service(post_image_multipart)
            .service(post_new_image)
    })
    .bind(("127.0.0.1", 9324))?
    .run();

    let public_api = HttpServer::new(|| App::new().service(images))
        .bind(("127.0.0.1", 9555))?
        .run();
    future::try_join(admin_api, public_api).await?;

    Ok(())
}
