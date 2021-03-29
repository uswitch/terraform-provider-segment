# terraform-provider-segment

## Contributing

### Releasing a new version

A new version automatically gets released in CI when pushing a new tag. To create a new release conveniently, use the following command from the `main` branch:
```shell
> make release TYPE=[major|minor|patch]
```