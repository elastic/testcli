# Contributing to terraform-provider-ec

Contributions are very welcome, these can include documentation, bug reports, issues, feature requests, feature implementations or tutorials.

## Reporting Issues

If you have found an issue or defect in the `testcli` framework or the latest documentation, use the GitHub [issue tracker](https://github.com/elastic/testcli/issues) to report the problem. Make sure to follow the template provided for you to provide all the useful details possible.

## Code Contribution Guidelines

For the benefit of all and to maintain consistency, we have come up with some simple guidelines. These will help us collaborate in a more efficient manner.

- Unless the PR is very small (e.g. fixing a typo), please make sure there is an issue for your PR. If not, make an issue with the provided template first, and explain the context for the change you are proposing.

  Your PR will go smoother if the solution is agreed upon before you've spent a lot of time implementing it. We encourage PRs to allow for review and discussion of code changes.

- We encourage PRs to be kept to a single change. Please don't work on several tasks in a single PR if possible.

- When you're ready to create a pull request, please remember to:
  - Make sure you've signed the Elastic [contributor agreement](https://www.elastic.co/contributor-agreement).

  - Have test cases for the new code. If you have questions about how to do this, please ask in your pull request.

  - Run `make format`, `make lint`.

  - Ensure that [unit](#unit) and [integration](#integration) tests succeed with `make unit integration`.

  - Use the provided PR template, and assign any labels which may fit your PR.

### Workflow

The codebase is maintained using the "contributor workflow" where everyone without exception contributes patch proposals using "pull requests". This facilitates social contribution, easy testing and peer review.

To contribute a patch, make sure you follow this workflow:

1. Fork repository.
2. Enable GitHub actions to run in your fork.
3. Create topic branch.
4. Commit patches.

### Commit Messages

In general commits should be atomic and diffs should be easy to read.

Commit messages should be verbose by default consisting of a short subject line. A blank line and detailed explanatory text as separate paragraph(s), unless the title alone is self-explanatory ("trivial: Fix comment typo in main.go"). Commit messages should be helpful to people reading your code in the future, so explain the reasoning for your decisions.

If a particular commit references another issue, please add the reference (e.g. refs #123 or closes #124).

Example:

```console
doc: Fix typo

Fixes a typo on the main readme.

Closes #1234
```

## Setting up a dev environment

### Environment prerequisites

- [Go](https://golang.org/doc/install) 1.16+

This project uses [Go Modules](https://blog.golang.org/using-go-modules) making it safe to work with it outside of your existing [GOPATH](http://golang.org/doc/code.html#GOPATH).
