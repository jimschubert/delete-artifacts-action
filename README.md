# Delete Artifacts Action

TODO

## Inputs

### `GITHUB_TOKEN`

**Required** Default: `${{github.token}}`

GitHub token used to access the repository defined in the GITHUB_REPOSITORY input.

It is recommended to [create a new personal access token](https://github.com/settings/tokens/new) with the least permissions (e.g. public_repo).
Using a service account for the GitHub Token is also highly recommended.

[Learn more about using secrets](https://help.github.com/en/actions/automating-your-workflow-with-github-actions/creating-and-using-encrypted-secrets)

### `GITHUB_REPOSITORY`

**Required** Default: `${{github.repository}}`

The target github repo in the format owner/repo

TODO

See also [jimschubert/delete-artifacts](https://github.com/jimschubert/delete-artifacts).

## Outputs

The action itself does not output any arguments.

## Usage

### Create a workflow

TODO

#### Include the Action 

TODO

#### Full Workflow Example

TODO

## License

This project is [licensed](./LICENSE) under Apache 2.0.
