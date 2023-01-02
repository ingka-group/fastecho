# ocp-go-utils
[![Quality Gate Status](https://sonarqube.ct.blue.cdtapps.com/api/project_badges/measure?project=ocp.go-utils&metric=alert_status&token=34cf6663e71a1d1e346d56eb479ee51ae7a1f774)](https://sonarqube.ct.blue.cdtapps.com/dashboard?id=ocp.go-utils) [![Reliability Rating](https://sonarqube.ct.blue.cdtapps.com/api/project_badges/measure?project=ocp.go-utils&metric=reliability_rating&token=34cf6663e71a1d1e346d56eb479ee51ae7a1f774)](https://sonarqube.ct.blue.cdtapps.com/dashboard?id=ocp.go-utils) [![Security Rating](https://sonarqube.ct.blue.cdtapps.com/api/project_badges/measure?project=ocp.go-utils&metric=security_rating&token=34cf6663e71a1d1e346d56eb479ee51ae7a1f774)](https://sonarqube.ct.blue.cdtapps.com/dashboard?id=ocp.go-utils) [![Coverage](https://sonarqube.ct.blue.cdtapps.com/api/project_badges/measure?project=ocp.go-utils&metric=coverage&token=34cf6663e71a1d1e346d56eb479ee51ae7a1f774)](https://sonarqube.ct.blue.cdtapps.com/dashboard?id=ocp.go-utils)

A collection of Golang utils that can be used between multiple microservices

- `echozap`: Configuration of Uber´s Zap logger, and a middleware for Golang Echo framework for logging HTTP requests
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
- `timeutils`: Contains functions that deal with time.Time. These can be conversions or other utilities

## How to make a new release?
Raise a PR and merge the code to the `main` branch, this will trigger a workflow that is responsible to tag the new release with the necessary version.
