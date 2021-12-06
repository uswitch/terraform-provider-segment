# terraform-provider-segment                                                                    |

The `terraform-provider-segment` is a custom [Terraform](https://www.terraform.io/) plugin to manage Segment infrastructure via code.

## Requirements

* [Terraform](https://www.terraform.io/downloads.html) 0.14.x or higher
* [Go](https://golang.org/) 1.16+

## Contributing

### Building the provider

To build the provider first update the architecture variable in the [Makefile](https://github.com/uswitch/terraform-provider-segment/blob/main/Makefile#L4) to your system architecture (supported architectures by Terraform can be found [here](https://www.terraform.io/docs/registry/providers/os-arch.html)) and then run:
```shell
$ make build
```
This will build the provider binary and move it to the `~/.terraform.d/plugins/uswitch.com/segment/segment/<PROVIDER VERSION>` directory so that it's ready to be imported and used in a Terraform project.

Make sure that the version being used in the terraform project uses the one built:
```tf
required_providers {
  segment = {
    source  = "uswitch.com/segment/segment"
    version = "<PROVIDER VERSION>"
  }
}
```

### Releasing a new version

A new version automatically gets released in CI when pushing a new tag. To create a new release conveniently, use the following command from the `main` branch:
```shell
$ make release TYPE=[major|minor|patch]
```

### Writing Acceptance Tests

Acceptance tests should be written for every new resource/data source. `resource_destination_filter_test.go` can be used as an example. A Segment token with read/write access to Sources and Tracking Plans will be required to run the tests and be stored in `SEGMENT_ACCESS_TOKEN`. The workspace to run the tests in must also be specified in `SEGMENT _WORKSPACE`.

### Building the docs

To build or update the provider documentation pages, run `go generate`.
