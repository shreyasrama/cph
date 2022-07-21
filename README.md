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

## Installation/Development
(Temporary)
1. Clone the repo:

    `git clone https://github.com/shreyasrama/cph.git`

1. Build a binary:

    `go build -mod=mod -o bin/cph main.go`

1. Test it out:

    `bin/cph`

### Taskfile
Task (https://taskfile.dev/) is a simple build tool used to help automate some tasks with `cph`.

## Todo
- ~Proper error handling everywhere (clean up os.exits too)~
- ~Refactor list.go to use awsutil functions~
- Accept selection of multiple pipelines -- underway
- Testing framework
- Sorting out function and variable case
- Several more functions (not in order of importance): get approvals and multi approve, detailed view of a single pipeline
- Setting up releases in Github and releasing via Taskfile
- Use https://github.com/olekukonko/tablewriter instead of tabwriter
