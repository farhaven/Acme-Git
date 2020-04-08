package main

import (
	"log"
	"os"

	"9fans.net/go/acme"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

// winFatal writes a formatted message to win's Error window and calls log.Fatalf
func winFatal(win *acme.Win, fmt string, args ...interface{}) {
	win.Errf(fmt, args...)
	log.Fatalf(fmt, args...)
}

func main() {
	win, err := acme.New()
	if err != nil {
		log.Fatalln("can't open ACME window: %s", err)
	}

	wd, err := os.Getwd()
	if err != nil {
		winFatal(win, "can't get working directory: %s", err)
	}

	win.Name("%s/-git", wd)

	repo, err := git.PlainOpen(".")
	if err != nil {
		winFatal(win, "can't open repo: %s", err)
	}

	win.Fprintf("data", "got repo: %v\n\n", repo)

	// Working tree status
	tree, err := repo.Worktree()
	if err != nil {
		winFatal(win, "can't get work tree: %s", err)
	}
	status, err := tree.Status()
	if err != nil {
		winFatal(win, "can't get status: %s", err)
	}

	win.Fprintf("data", "Tree status:\n%s\n", status)

	// List branches
	branches, err := repo.Branches()
	if err != nil {
		winFatal(win, "can't get branches: %s", err)
	}
	defer branches.Close()

	win.Fprintf("data", "Branches:\n")
	branches.ForEach(func(ref *plumbing.Reference) error {
		win.Fprintf("data", "\t%s\n", ref.Name())
		return nil
	})

	win.Ctl("clean")
}
