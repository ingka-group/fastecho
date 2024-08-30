# ocp-go-utils
[![Quality Gate Status](https://sonarqube.ct.blue.cdtapps.com/api/project_badges/measure?project=ocp.go-utils&metric=alert_status&token=34cf6663e71a1d1e346d56eb479ee51ae7a1f774)](https://sonarqube.ct.blue.cdtapps.com/dashboard?id=ocp.go-utils) [![Reliability Rating](https://sonarqube.ct.blue.cdtapps.com/api/project_badges/measure?project=ocp.go-utils&metric=reliability_rating&token=34cf6663e71a1d1e346d56eb479ee51ae7a1f774)](https://sonarqube.ct.blue.cdtapps.com/dashboard?id=ocp.go-utils) [![Security Rating](https://sonarqube.ct.blue.cdtapps.com/api/project_badges/measure?project=ocp.go-utils&metric=security_rating&token=34cf6663e71a1d1e346d56eb479ee51ae7a1f774)](https://sonarqube.ct.blue.cdtapps.com/dashboard?id=ocp.go-utils) [![Coverage](https://sonarqube.ct.blue.cdtapps.com/api/project_badges/measure?project=ocp.go-utils&metric=coverage&token=34cf6663e71a1d1e346d56eb479ee51ae7a1f774)](https://sonarqube.ct.blue.cdtapps.com/dashboard?id=ocp.go-utils)

A collection of Golang utils that can be used between multiple microservices

- `agg`: A small aggregation package that helps with common use cases like aggregating by IKEA
  weeks. Read the [documentation](./agg/README.md) for details.
- `api`: Contains common handlers for building REST APIs and boilerplate code for testing (e.g. integration with testcontainers).
- `docs`: A utility to generate a Golang file with the functions and their documentation instructed to look up. Under the hood
  the `go doc` is used. This is particularly useful to allow `swaggo` to generate the OpenAPI docs for handlers that do not
  directly exist in the codebase of the repository that is using `ocp-go-utils`. In such cases, we have to create dummy functions
  and "import" their corresponding comments from `ocp-go-utils` manually so that `swaggo` can read and generate the specification
  for the imported handlers. The utility `docs` is a helper to automate this process. It can be called with `go generate` before
  the `make swagger` to generate the relevant files before `swaggo` translates them into the OpenAPI specification. As this is a
  CLI tool you can run it using `--help` flag to see the required arguments.
- `echozap`: Configuration of Uber´s Zap logger, and a middleware for Golang Echo framework for logging HTTP requests
- `fp`: A functional programming package that includes common functions like `Map`, `Reduce`
  and `Filter`. Read the [documentation](./fp/README.md) for details.
- `rest`: A rest client with which we can perform HTTP requests, and easily mock in our services for unit testing. There are
  two functions that can be used for the same purpose, namely, `Request` and `DoRequest`. It's best to use the latter as the former
  will be deprecated in future versions.
- `gcpstorage`: A storage client for GCS (Google Cloud Storage) with which we can manipulate data on a bucket, and easily mock in our services for unit testing
- `date.IKEAWeek`: Returns the year and week number in which the given date (specified by year, month, day) occurs,
  according to IKEA week numbering scheme which happens to match the US CDC epiweeks, i.e. Weeks start on Sundays
  and first WOY contains four days. Week ranges from 1 to 53; Jan 01 to Jan 03 of year n might belong to week 52 or
  53 of year n-1, and Dec 29 to Dec 31 might belong to week 1 of year n+1.
- `date.ISODate`: A custom type to unmarshal a date string to a time.Time object
- `stringutils`: Provides a bunch of functions that deal with strings or conversions from strings
- `fastecho`: Opinionated easy to set up golang microservice in the style of Python's fastAPI
- `otel`: Go implementations of an OpenTelemetry collector for managing observability data.

## How to make a new release?
Raise a PR and merge the code to the `main` branch, this will trigger a workflow that is responsible to tag the new release with the necessary version.
