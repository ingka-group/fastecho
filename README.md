# ocp-go-utils
A collection of Golang utils that can be used between multiple microservices

- `echozap`: A middleware for Golang Echo framework that provides integration with Uber´s Zap logging library for logging HTTP requests
- `rest`: A rest client with which we can perform HTTP requests, and easily mock in our services for unit testing
- `gcpstorage`: A storage client for GCS (Google Cloud Storage) with which we can manipulate data on a bucket, and easily mock in our services for unit testing

## How to make a new release?
Simply tag the main branch with a version string, the initial tag is v1.0.0
