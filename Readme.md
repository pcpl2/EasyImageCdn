# EasyImageCdn

Application to create a simple cdn server for images.

This application automatically converts the uploaded image to webp format and to all resolutions defined in the configuration.

## How to use

```sh
docker run --name imagecdn -v /my/images/location:/images -e API_KEY=EnterAdminKey -d ghcr.io/pcpl2/easy_image_cdn:0.1.2
```

OR

```sh
docker run --name imagecdn -v /my/images/location:/images -e API_KEY=EnterAdminKey -d pcpl2/easy_image_cdn:0.1.2
```

This command launches the application with image conversion to 1024x720 and 800x600 enabled, with a maximum file size of 10Mb and your api key.

### Endpoints

#### Admin

`http://localhost:9324/v1/newImage` -> For send and update Image
Payload:

```json
{
    "id": {Your image id as string},
    "image": {Your image in base64}
}
```

#### Public

`http://localhost:9555/{Your image id}/source` -> Has return source image (if you have `image/webp` in accept header server will return the image in webp format).

For get image in coverted resolution you must replace `source` to resolution value. Example:
`http://localhost:9555/{Your image id}/1024x720`

## Configuration

### Example .env file

```env
ADMIN_HTTP_ADDR=0.0.0.0:9324
PUBLIC_HTTP_ADDR=0.0.0.0:9555
API_KEY=00000000-0000-0000-0000-000000000000
API_KEY_HEADER=key
FILES_PATH=/var/lib/images
CONVERT_TO_RES=1024x720,800x600
MAX_FILE_SIZE=10
```

### Config values description

| Configuration key | Default value | Description |
| ----------- | --------- | ----------- |
| ADMIN_HTTP_ADDR | 0.0.0.0:9324 | Http address and port for upload images. |
| PUBLIC_HTTP_ADDR | 0.0.0.0:9555 | Http address for getting images. |
| API_KEY | 00000000-0000-0000-0000-000000000000 | Api key for upload images |
| API_KEY_HEADER | key | Header name for api key in request. |
| FILES_PATH | /var/lib/images | Path to the directory where the files will be saved . |
| CONVERT_TO_RES | 1024x720,800x600 | List of resolutions to which images will be converted. |
| MAX_FILE_SIZE | 10 | Maximum size of file sent to the application in megabytes. |
| CACHE_TIME | 30 | Image cache lifetime set in minutes. |
