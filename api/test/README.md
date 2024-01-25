This test package offers a simple structure for writing tests. A `fixtures` folder has to exist in the location where the `_test.go` files are so that files can be read. This also ensures that fixtures are kept close to the test and in the relevant package.

Optional usage of mocks

Mocks are stored with the rest of the fixtures as `.json` files in a `mocks` folder within `fixtures`. Mocking a request consists of pairing a request URL with a status code and optionally a response.

NOTE: mocks are single-use. If your code repeatedly calls the same endpoint expecting the same response, writing this once is not enough. You need to either replicate the mock itself or extend the current implementation to make use of [times](https://pkg.go.dev/github.com/h2non/gock#Request.Times) to make a request reuse a mock a specific number of times.

Example
```go
it := newtest.NewIntegrationTest(
    t,
    newtest.IntegrationTestWithMocks{
        BaseURL: "/v1",
    },
)
defer func() {
    it.TearDown()
}()

handler := mockHandler()

tests := []newtest.Data{
    {
        Name:   "my test case",
        Method: http.MethodGet,
        Params: newtest.Params{
            Path: map[string]string{
                "id": testdata.One,
            },
        },
        Mocks: []newtest.MockCall{
            {
                Config: &newtest.MockConfig{
                    UrlPath:    fmt.Sprintf("/v1/forecasts/%s", testdata.One),
                    StatusCode: http.StatusNotFound,
                },
            },
        },
        Handler:           handler.MyEndpoint,
        ExpectCode:        http.StatusNotFound,
        ExpectErrResponse: true,
    },
}
apitest.AssertAllWithCustomContext(it, tests) //or AssertAll(it, tests) for basic functionality
```
