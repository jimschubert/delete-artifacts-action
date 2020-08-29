# Delete Artifacts Action

A GitHub Action to help with deleting workflow artifacts.

See also [jimschubert/delete-artifacts](https://github.com/jimschubert/delete-artifacts).

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

### `run_id`

**Optional**

A specific Actions run identifier. You can use this to specify the current run and do some post-run cleanup.

### `min_bytes`

**Required**

The minimum file size in bytes. The default is "0", effectively meaning it will start off by matching all files.

Consider applying another filter such as name or pattern if you'd like to avoid deleting everything.

**NOTE** This value must be a quoted integer.

### `max_bytes`

**Optional**

The maximum file size in bytes. If not specified, there is no limit.

This option is useful if you have many small artifacts and want to keep a specific larger artifact.

**NOTE** This value must be a quoted integer.

### `artifact_name`

**Optional**

The name of a specific artifact to delete. If not specified, this may result in _all_ artifacts being deleted.

### `pattern`

**Optional**

A POSIX regular expression. This is useful, for example, if you have matrix artifacts with a common prefix or suffix.


### `active_duration`

**Optional**

A duration string which defines the duration during which artifacts are considered "active" and will therefore not be deleted.

This setting is useful to avoid deleting artifacts for current or very recent runs (e.g. specify "10m"), or to allow
for debugging of artifacts for some amount of time (e.g. specify "23h59m"). This acts as a retention period for artifacts
when using otherwise aggressive deletion settings.

This format follows the [go Duration](https://golang.org/pkg/time/#ParseDuration) formatting.

### `log_level`

**Optional**

Specifies a custom log level.

Choose from these options:

* debug
* info
* warn
* error

### `dry_run`

**Optional**

Perform a dry-run. It's recommended to do this first to be sure you have correct settings.

## Outputs

The action itself does not output any arguments.

## Usage

### Create a triggered workflow

Artifacts become associated with a workflow once the workflow completes. You can use a repository dispatch to "kick off" an artifact cleanup.

**This step requires a Personal Access Token, rather than the `GITHUB_TOKEN` available by default**

Create this first file at `.github/workflows/generate.yml`:

```yaml
name: Generate artifacts and trigger cleanup
on:
  workflow_dispatch:
    inputs:
      size:
        description: 'Target file size'
        required: true
        default: '1M'
      name:
        description: 'Artifact name'
        required: true
        default: 'artifact.bin'

defaults:
  run:
    shell: bash

jobs:

  # This job represents your "standard" job. notice that the job outputs `github.run_id` as an example of passing variables to a triggering job
  generate:
    runs-on: ubuntu-latest
    continue-on-error: false
    outputs:
      # We can use outputs across jobs in the final trigger job
      target_run_id: ${{ github.run_id }}
    steps:
      # Generate a file of empty bytes (this could be a build task)
      - name: Generate a file
        run: |
          truncate -s ${{ github.event.inputs.size }} ${{ github.event.inputs.name }}
          echo "Created file ${{ github.event.inputs.name }} (${{ github.event.inputs.size }})"
      # Upload the artifact. Artifacts expire after 90 days, which may count toward storage costs on private repositories
      - name: Upload artifact as ${{ github.event.inputs.name }}
        uses: actions/upload-artifact@v1
        with:
          name: ${{ github.event.inputs.name }}
          path: ${{ github.event.inputs.name }}

  # This example triggers via separate job to demonstrate passing variables between jobs and workflows, but the step can exist in a single job. 
  trigger:
    runs-on: ubuntu-latest
    needs: generate
    continue-on-error: false
    steps:
      # See https://docs.github.com/en/rest/reference/repos#create-a-repository-dispatch-event
      # This requires a Personal Access Token (PAT), the GITHUB_TOKEN provided to workflows will not work.
      - name: Repository Dispatch
        uses: peter-evans/repository-dispatch@v1.0.0
        with:
          # THIS MUST BE A PERSONAL ACCESS TOKEN
          token: ${{secrets.TRIGGER_TOKEN}}
          repository: ${{github.repository}}
          event-type: triggered-event
          # Payload properties can be any name. Note size, base_name, and run_id referring to this workflow's inputs and the generate tasks's outputs
          client-payload: '{"ref": "${{ github.ref }}", "sha": "${{ github.sha }}", "size": "${{ github.event.inputs.size }}", "base_name": "${{ github.event.inputs.name }}", "run_id": "${{needs.generate.outputs.target_run_id}}"}'
```

Create the following file at `.github/workflows/cleanup.yml`:

```yaml
name: Cleanup
on:
  repository_dispatch:
    types:
      - triggered-event

jobs:

  # Demonstrate how to delete by name
  delete-by-name:
    runs-on: ubuntu-latest
    continue-on-error: true
    steps:
      - name: Checkout
        uses: actions/checkout@v2

      - name: Delete Artifact by Name
        uses: jimschubert/delete-artifacts-action@v1.0.0
        with:
          log_level: 'debug'
          artifact_name: '${{ github.event.client_payload.base_name }}-by-name'
          min_bytes: '0'

  # Demonstrate how to delete by active_duration
  delete-by-retention-duration:
    needs: delete-by-name
    continue-on-error: true
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v2

      - name: Delete Artifact by active duration (expect no deletions)
        uses: jimschubert/delete-artifacts-action@v1.0.0
        with:
          log_level: 'debug'
          min_bytes: '0'
          active_duration: '30m'

      - name: Sleep 2 minutes
        run: |
          echo "Sleeping for 2 minutes, then we'll use duration 1m30s"
          sleep 120

      - name: Delete Artifact by active duration (expect deletions)
        uses: jimschubert/delete-artifacts-action@v1.0.0
        with:
          log_level: 'debug'
          min_bytes: '0'
          active_duration: '1m30s'

  # Demonstrates how to use POSIX pattern for deleting artifacts
  delete-by-pattern:
    needs: delete-by-retention-duration
    continue-on-error: true
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v2

      - name: Delete by pattern
        uses: jimschubert/delete-artifacts-action@v1.0.0
        with:
          log_level: 'debug'
          min_bytes: '0'
          pattern: '\.pat'

  # Demonstrates how to use a run id for deleting artifacts
  # This could be used for overall cleanup
  delete-by-runId:
    needs: [delete-by-pattern]
    continue-on-error: true
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v2

      - name: Delete by run id
        if: ${{ github.event.client_payload.run_id }}
        uses: jimschubert/delete-artifacts-action@v1.0.0
        with:
          log_level: 'debug'
          min_bytes: '0'
          run_id: '${{ github.event.client_payload.run_id }}'
```

#### Create a cron workflow

The following example runs cleanup of _all_ artifacts weekly at a specific time.

In your repository as `.github/workflows/cleanup.yml`:

```yaml
name: Weekly cleanup

on:
  schedule:
    - cron: "0 18 * * 6"

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v1
      - name: Weekly Artifact Cleanup
        uses: jimschubert/delete-artifacts-action@v1.0.0
        with:
          log_level: 'error'
          min_bytes: '0'
```

## License

This project is [licensed](./LICENSE) under Apache 2.0.
