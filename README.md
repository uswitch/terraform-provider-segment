# terraform-provider-segment

| What                 | Where                                                                                                                      |
| -------------        | -------------                                                                                                              |
| Team                 | Customer Platform                                                                                                          |
| Support              | #customer-platform-support                                                                                                 |
| Design Doc           | [Design Doc (in Notion)](https://www.notion.so/rvu/Terraform-Segment-Provider-Design-Doc-c7aec0b7cdd84d81910dccba6a3afe31) |
| Programming Language | Golang ([https://golang.org/](https://golang.org/))                                                                        |

The `terraform-provider-segment` is a custom [Terraform](https://www.terraform.io/) plugin to manage Segment infrastructure via code.

## Contributing

### Releasing a new version

A new version automatically gets released in CI when pushing a new tag. To create a new release conveniently, use the following command from the `main` branch:
```shell
> make release TYPE=[major|minor|patch]
```
