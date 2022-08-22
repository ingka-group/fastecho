# ocp-go-utils
A collection of Golang utils that can be used between multiple microservices

- `echozap`: Configuration of Uber´s Zap logger, and a middleware for Golang Echo framework for logging HTTP requests
- `rest`: A rest client with which we can perform HTTP requests, and easily mock in our services for unit testing
- `gcpstorage`: A storage client for GCS (Google Cloud Storage) with which we can manipulate data on a bucket, and easily mock in our services for unit testing
- `date.IKEAWeek`: returns the year and week number in which the given date (specified by year, month, day) occurs,
  according to IKEA week numbering scheme which happens to match the US CDC epiweeks, i.e. Weeks start on Sundays
  and first WOY contains four days. Week ranges from 1 to 53; Jan 01 to Jan 03 of year n might belong to week 52 or
  53 of year n-1, and Dec 29 to Dec 31 might belong to week 1 of year n+1.

## How to make a new release?
Simply tag the main branch with a version string, the initial tag is v1.0.0
