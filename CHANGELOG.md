# 0.2.2 (???)

* Updated go version to 1.21
* Updated fiber to 2.48.0
* Updated go image to 0.11.0
* Small fix for logs directory create
* Added app version and build time in code
* Added endpoint for monitoring expvar
* Added endpoint for monitoring Pprof

# 0.2.1 (18.10.2022)

* Fixed logs path in container

# 0.2.0 (18.10.2022)

* Added fiber
* Added Etag generate for public endpints
* Added setting Cache-Control
* Added generate docker image for arm64
* Added multipart upload endpint (more information in readme)
* Updated go version to 1.18
* Rewritted image processing
* Removed error with missing .env file in docker container
* Removed unused addr configuration
* Removed unused files path configuration
* Removed bimg library
* Public api is not require adding source name for getting default image
* Moved docker image from alpine to distro-less
* Implemented app logger with saving to file.
* Fixed some warnings and better error handling.

# 0.1.2 (07.01.2022)

* Implemented detect image format by accept in browser.
* Updated go version to 1.17
* Clean code.
* Removed support for `config.json`, all configuration current works on enviroments or `.env` file

# 0.1.1 (13.10.2021)

* First release
