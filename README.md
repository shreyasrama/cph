# cph
`cph` (CodePipeline Helper) is a CLI tool written in Go, and designed to make interacting with AWS CodePipeline a little bit easier.

## Usage
```
# List pipelines using a search term
cph list --name pipeline_name

# Run pipelines using a search term
cph run --name pipeline_name

## Approve pipelines using a search term
cph approve --name pipeline_name
```

## Installation
go install github.com/shreyasrama/cph@latest

### Taskfile
Task (https://taskfile.dev/) is a simple build tool used to help automate some tasks with `cph`.

## Todo
- ~~Proper error handling everywhere (clean up os.exits too)~~ `done`
- ~~Refactor list.go to use awsutil functions~~ `done`
- ~~Accept selection of multiple pipelines~~ `done`
- Testing framework
- Sorting out function and variable case
- Several more functions (not in order of importance): ~~get approvals and multi approve~~ `done`, detailed view of a single pipeline
- ~~Setting up releases in Github and releasing via Taskfile~~ `done`
- ~~Use https://github.com/olekukonko/tablewriter instead of tabwriter~~ `done`
- Investigate if there are useful GH Actions that can be added for CI
- Add search by tag
