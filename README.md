## NAME
git-org-fork - A Git extension to fork repos of a GitHub org

## SYNOPSIS
    git org-fork [-h|--help] <github_org>

## DESCRIPTION
Given:

* GitHub org, github_org, as positional arguments
* API environment variables:
	* `$GITHUB_USERNAME`
	* `$GITORGFORK_GITHUB_API_TOKEN`

This program idempotently:

* Forks repos by hitting GitHub API

## EXAMPLES
    $ git org-fork -h
    $ git org-fork $COMPANY_GITHUB_ORG

## AUTHOR
* Payam Tanaka, @tanakapayam
