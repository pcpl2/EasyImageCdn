# EasyImageCdn

[![Build](https://github.com/pcpl2/EasyImageCdn/actions/workflows/buildApp.yml/badge.svg)](https://github.com/pcpl2/EasyImageCdn/actions/workflows/buildApp.yml) ![Docker Image Size with architecture (latest by date/latest semver)](https://img.shields.io/docker/image-size/pcpl2/easy_image_cdn?arch=amd64&label=Image%20size%20amd64&sort=date) ![Docker Image Size with architecture (latest by date/latest semver)](https://img.shields.io/docker/image-size/pcpl2/easy_image_cdn?arch=arm64&label=Image%20size%20arm64&sort=date) ![Docker Pulls](https://img.shields.io/docker/pulls/pcpl2/easy_image_cdn) ![GitHub](https://img.shields.io/github/license/pcpl2/EasyImageCdn) ![Docker Image Version (tag latest semver)](https://img.shields.io/docker/v/pcpl2/easy_image_cdn/0.3.0-beta.1) [![CodeFactor](https://www.codefactor.io/repository/github/pcpl2/easyimagecdn/badge)](https://www.codefactor.io/repository/github/pcpl2/easyimagecdn) ![GitHub Sponsors](https://img.shields.io/github/sponsors/pcpl2)


Application to create a simple cdn server for images.

This application automatically converts the uploaded image to webp format and to all resolutions defined in the configuration.

### Warning! The version on this branch is experimental and does not have all functionalities implemented.

## How to use

```sh
docker run --name imagecdn -v /my/images/location:/output -e API_KEY=EnterAdminKey -d ghcr.io/pcpl2/easy_image_cdn:0.3.0-beta.1
```

OR

```sh
docker run --name imagecdn -v /my/images/location:/var/lib/images -e API_KEY=EnterAdminKey -d pcpl2/easy_image_cdn:0.3.0-beta.1
```

This command launches the application with image conversion to 1024x720 and 800x600 enabled, with a maximum file size of 10Mb and your API key.

### Example docker-compose config

```yml
name: 'my-cdn'
  cdn:
    image: pcpl2/easy_image_cdn:0.3.0-beta.1
    restart: always
    environment:
      API_KEY: 'EnterAdminKey'
      CONVERT_TO_RES: '1024x720,800x600'
      MAX_FILE_SIZE: 15
    ports:
      - '9324:9324'
      - '9555:9555'
    volumes:
      - './images:/output'
      - './logs:/var/log/eic'
```

## Endpoints
All api definitions has moved to [swagger https://pcpl2.github.io/EasyImageCdn/](https://pcpl2.github.io/EasyImageCdn/?urls.primaryName=EasyImageCdn+0.3.0-beta.1)

## Configuration

### Example .env file

```env
API_KEY=00000000-0000-0000-0000-000000000000
API_KEY_HEADER=key
CONVERT_TO_RES=1024x720,800x600
MAX_FILE_SIZE=10
TARGET_FORMATS=jpg,webp,avif
```

### Config values description

| Configuration key | Default value | Description |
| ----------- | --------- | ----------- |
| API_KEY |  | Api key for upload images. If not set application throw error. |
| API_KEY_HEADER |  | Header name for an API key in the request. If not set application throw error. |
| CONVERT_TO_RES | 1024x720,800x600 | List of resolutions to which images will be converted. |
| TARGET_FORMATS | jpg | List of target image formats. Current supported is `jpg`,`webp`,`avif` |
| MAX_FILE_SIZE | 10 | Maximum size of the file sent to the application in megabytes. |
| CACHE_CONTROL_HEADER | max-age=180, public | Sets HTTP header `Cache-Control` for control caching in browsers. [More information](https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Headers/Cache-Control) |
| CACHE_TIME | 30 | Image cache lifetime set in minutes. **(Not implemented yet)** |

### Volumes configuration in container

| Path | Description |
| ----------- | ----------- |
| `/output` | Location for saving all images |
| `/var/log/eic` | Location for Application log files **(Not implemented yet)** |

## Community

* ❓ Ask questions on [GitHub Discussions](https://github.com/pcpl2/EasyImageCdn/discussions).

## Roadmap to 0.3.0
- [ ] Loging system
- [ ] Migration from 0.2.x
- [ ] Cache for public api

## Sponsors

TODO()
