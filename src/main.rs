use actix_multipart::form::MultipartForm;
use actix_web::{get, post, web, App, HttpResponse, HttpServer, Responder};
use futures::future;
use actix_files as fs;
use actix_web::http::header::{ContentDisposition, DispositionType};
use actix_web::{ Error, HttpRequest};

mod models;

#[post("/v1/newImage")]
async fn post_new_image(form: web::Form<models::rest_models::ImagePayload>) -> HttpResponse {
    HttpResponse::Ok().body(format!("username: {}", form.id))
}

#[post("/v1/newImageMP")]
pub async fn post_image_multipart(MultipartForm(form): MultipartForm<models::rest_models::ImageMultipartPayload>) -> impl Responder {
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
            .service(post_image_multipart)
            .service(post_new_image)
    })
    .bind(("127.0.0.1", 9324))?
    .run();

    let public_api = HttpServer::new(|| {
        App::new().service(images)
    })
    .bind(("127.0.0.1", 9555))?
    .run();
    future::try_join(admin_api, public_api).await?;

    Ok(())
}
