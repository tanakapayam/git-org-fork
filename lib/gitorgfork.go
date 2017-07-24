/*
  Purpose:
    This program forks repos by hitting GitHub API based on username, token and organization(s).

  Environment Variables:
    * GITHUB_USERNAME
    * GITORGFORK_GITHUB_API_TOKEN

  Command-Line Arguments:
    * githubOrgs (one or many)

  Stdout:
    * Per-Repo Fork URLs

  Stderr:
    * Missing GITHUB_USERNAME
    * Missing GITORGFORK_GITHUB_API_TOKEN
    * Failures from GitHub API

  Original Author:
    * Payam Tanaka (@tanakapayam)
*/

package gitorgfork

import (
	"bufio"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/user"
	"path"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/google/go-github/github"
	flag "github.com/ogier/pflag"
	"golang.org/x/oauth2"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
)

const (
	README  = "README.md"
	VERSION = "VERSION.txt"
	DELAY   = 500 * time.Millisecond
)

var (
	bold           = color.New(color.Bold).SprintFunc()
	green          = color.New(color.FgGreen).SprintFunc()
	yellow         = color.New(color.FgHiYellow).SprintFunc()
	red            = color.New(color.FgHiRed).SprintFunc()
	currentUser, _ = user.Current()

	// Assuming id_rsa is available at ~/.ssh/id_rsa
	sshAuth, _ = ssh.NewPublicKeysFromFile(
		"git",
		currentUser.HomeDir+"/.ssh/id_rsa",
		"",
	)
)

func check(e error) {
	if e != nil {
		fmt.Println(e)
	}
}

func ParseArgs() []string {
	var h2 = regexp.MustCompile(`^#+ (.+)$`)
	var indented = regexp.MustCompile(`^[ \t]`)

	flag.Usage = func() {
		_, program, _, _ := runtime.Caller(0)

		// read in README.md
		if readme, err := os.Open(
			path.Dir(program) + "/../" + README,
		); err == nil {
			// make sure it gets closed
			defer readme.Close()

			// create a new scanner and read the file line by line
			scanner := bufio.NewScanner(readme)
			for scanner.Scan() {
				// check for #
				h2Result := h2.FindStringSubmatch(scanner.Text())
				if h2Result != nil {
					fmt.Printf("%s\n", bold(h2Result[1]))
				} else {
					indentedResult := indented.FindStringSubmatch(scanner.Text())
					if indentedResult == nil {
						fmt.Println("    " + scanner.Text())
					} else {
						fmt.Println(scanner.Text())
					}
				}
			}

			// check for errors
			if err = scanner.Err(); err != nil {
				log.Fatal(err)
			}
		} else {
			log.Fatal(err)
		}

		// read in VERSION.txt
		if version, err := ioutil.ReadFile(
			path.Dir(program) + "/../" + VERSION,
		); err == nil {
			fmt.Printf("\n%s\n    %s\n", bold("VERSION"), string(version))
		} else {
			log.Fatal(err)
		}

		flag.PrintDefaults()
	}
	flag.Parse()

	githubOrgs := flag.Args()

	if len(githubOrgs) == 0 {
		flag.Usage()
	}

	return githubOrgs
}

// Throw if environment variable is not set
func getEnv(ev string) string {
	if os.Getenv(ev) == "" {
		log.Fatal(ev, " needs to be set")
	}

	return os.Getenv(ev)
}

func ProcessRepos(githubOrgs []string) {
	// processing...
	// list all repositories for the authenticated user
	// fork repos
	// synchronize
	// pull repos

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: getEnv("GITORGFORK_GITHUB_API_TOKEN")},
	)
	tc := oauth2.NewClient(ctx, ts)

	gitClient := github.NewClient(tc)

	// list all repositories for the authenticated user
	for _, githubOrg := range githubOrgs {
		fmt.Println(bold("# "+githubOrg) + "\n")

		// get all pages of results
		fmt.Print(bold("Processing"))
		opt := &github.RepositoryListByOrgOptions{
			ListOptions: github.ListOptions{PerPage: 10},
		}
		var allRepos []*github.Repository
		for {
			repos, resp, err := gitClient.Repositories.ListByOrg(ctx, githubOrg, opt)
			if err != nil {
				log.Fatal(err)
			}
			allRepos = append(allRepos, repos...)
			if resp.NextPage == 0 {
				break
			}
			opt.Page = resp.NextPage
			time.Sleep(DELAY)
			fmt.Printf(yellow("."))
		}
		fmt.Println("\n")

		var postClient = &http.Client{
			Timeout: 10 * time.Second,
		}

		// fork repos
		fmt.Println(bold("## Forking") + "\n")
		c := make(chan bool)
		counter := 0
		for _, repo := range allRepos {
			if !strings.HasSuffix(*repo.ForksURL, "/forks") {
				continue
			}

			forkURL := strings.Replace(
				*repo.HTMLURL,
				*repo.Owner.Login,
				getEnv("GITHUB_USERNAME"),
				1,
			)
			fmt.Printf(
				"  %-70s -> %s\n",
				" "+forkURL,
				*repo.HTMLURL,
			)

			time.Sleep(DELAY)
			go forkRepo(*postClient, *repo, c)
			counter++
		}
		fmt.Println()

		// synchronize
		fmt.Print(bold("Processing"))
		if counter > 0 {
			for ; counter > 0; counter-- {
				<-c
				fmt.Printf(yellow("."))
			}
			fmt.Println("\n")
		} else {
			fmt.Println(".\n" + green("Already up-to-date.\n"))
		}
	}
}

func forkRepo(postClient http.Client, repo github.Repository, c chan bool) {
	req, err := http.NewRequest(
		"POST",
		*repo.ForksURL,
		nil,
	)
	check(err)

	req.Header.Set(
		"Authorization",
		"token "+getEnv("GITORGFORK_GITHUB_API_TOKEN"),
	)

	res, err := postClient.Do(req)
	check(err)

	defer res.Body.Close()

	c <- true
}
