# terraform-provider-segment

The `terraform-provider-segment` is a custom [Terraform](https://www.terraform.io/) plugin to manage Segment infrastructure via code.

## Requirements

* [Terraform](https://www.terraform.io/downloads.html) 0.14.x or higher
* [Go](https://golang.org/) 1.16+

## Contributing

### Adding Dependencies

This provider uses [Go modules](https://github.com/golang/go/wiki/Modules).
Please see the Go documentation for the most up to date information about using Go modules.

To add a new dependency `github.com/author/dependency` to your Terraform provider:

```
go get github.com/author/dependency
go mod tidy
```

Then commit the changes to `go.mod` and `go.sum`.

### Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

To build the provider run:
```shell
$ make build
```
This will build the provider binary and move it to the `bin/` directory.

To generate or update documentation, run `go generate`.


### Writing Acceptance Tests

Acceptance tests should be written for every new resource/data source. `resource_destination_filter_test.go` can be used as an example. A Segment token with read/write access to Sources and Tracking Plans will be required to run the tests and be stored in `SEGMENT_ACCESS_TOKEN`. The workspace to run the tests in must also be specified in `SEGMENT _WORKSPACE`.
