# terraform-provider-segment

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
This will build the provider binary and move it to the `~/.terraform.d/` directory so that it's ready to be imported and used in a Terraform project.

### Releasing a new version

A new version automatically gets released in CI when pushing a new tag. To create a new release conveniently, use the following command from the `main` branch:
```shell
$ make release TYPE=[major|minor|patch]
```
