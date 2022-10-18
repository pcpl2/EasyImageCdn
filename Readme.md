# EasyImageCdn 
[![Build](https://github.com/pcpl2/EasyImageCdn/actions/workflows/buildApp.yml/badge.svg)](https://github.com/pcpl2/EasyImageCdn/actions/workflows/buildApp.yml) ![Docker Image Size with architecture (latest by date/latest semver)](https://img.shields.io/docker/image-size/pcpl2/easy_image_cdn?arch=amd64&label=Image%20size%20amd64&sort=date) ![Docker Image Size with architecture (latest by date/latest semver)](https://img.shields.io/docker/image-size/pcpl2/easy_image_cdn?arch=arm64&label=Image%20size%20arm64&sort=date) ![Docker Pulls](https://img.shields.io/docker/pulls/pcpl2/easy_image_cdn) ![GitHub](https://img.shields.io/github/license/pcpl2/EasyImageCdn) ![Docker Image Version (tag latest semver)](https://img.shields.io/docker/v/pcpl2/easy_image_cdn/0.2.0) ![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/pcpl2/EasyImageCdn) [![CodeFactor](https://www.codefactor.io/repository/github/pcpl2/easyimagecdn/badge)](https://www.codefactor.io/repository/github/pcpl2/easyimagecdn)

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


### Example docker-compose config
```yml
version: '3.8'
  cdn:
    image: pcpl2/easy_image_cdn:0.1.2
    restart: always
    environment:
      API_KEY: 'EnterAdminKey'
      CONVERT_TO_RES: '1024x720,800x600'
      MAX_FILE_SIZE: 15
    ports:
      - '9324:9324'
      - '9555:9555'
    volumes:
      - './images:/var/lib/images'
      - './logs:/var/log/eic'
```

### Endpoints

#### Admin:

`http://localhost:9324/v1/newImage` -> For send and update Image
Payload:

```json
{
    "id": {Your image id as string},
    "image": {Your image in base64}
}
```
`http://localhost:9324/v1/newImageMp?imageId={Your image id}` -> For send and update Image as multipart and define image in multipart as `imageFile`

#### Public:

`http://localhost:9555/{Your image id}` -> Has return source image (if you have `image/webp` in accept header server will return the image in webp format).

For get image in coverted resolution you add resolution value after image id. Example:
`http://localhost:9555/{Your image id}/1024x720`

## Configuration

### Example .env file

```env
API_KEY=00000000-0000-0000-0000-000000000000
API_KEY_HEADER=key
CONVERT_TO_RES=1024x720,800x600
MAX_FILE_SIZE=10
```

### Config values description

| Configuration key | Default value | Description |
| ----------- | --------- | ----------- |
| API_KEY | 00000000-0000-0000-0000-000000000000 | Api key for upload images |
| API_KEY_HEADER | key | Header name for api key in request. |
| CONVERT_TO_RES | 1024x720,800x600 | List of resolutions to which images will be converted. |
| MAX_FILE_SIZE | 10 | Maximum size of file sent to the application in megabytes. |
| CACHE_TIME | 30 | Image cache lifetime set in minutes. |


### Volumes configuration in container
| Path | Description |
| ----------- | ----------- |
| `/var/lib/images` | Location for saving all images |
| `/var/log/eic` | Location for Application log files |


## Community
* ❓ Ask questions on [GitHub Discussions](https://github.com/pcpl2/EasyImageCdn/discussions).

## Sponsors
TODO()
