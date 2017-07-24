package main

import (
	gitorgfork "github.com/tanakapayam/git-org-fork/lib"
)

func main() {
	gitorgfork.ProcessRepos(
		gitorgfork.ParseArgs(),
	)
}
