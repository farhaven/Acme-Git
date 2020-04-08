package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"9fans.net/go/acme"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

// winFatal writes a formatted message to win's Error window and calls log.Fatalf
func winFatal(win *acme.Win, fmt string, args ...interface{}) {
	win.Errf(fmt, args...)
	log.Fatalf(fmt, args...)
}

func refresh(win *acme.Win, repo *git.Repository) error {
	win.Clear()

	win.Fprintf("data", "got repo: %v\n\n", repo)

	// Working tree status
	tree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("can't get work tree: %w", err)
	}
	status, err := tree.Status()
	if err != nil {
		return fmt.Errorf("can't get status: %w", err)
	}

	if status.IsClean() {
		win.Fprintf("data", "Tree is clean\n\n")
	} else {
		win.Fprintf("data", "Tree status: Ci\n%s\n", status)
	}

	head, err := repo.Head()
	if err != nil {
		return fmt.Errorf("can't get repo head: %w", err)
	}
	win.Fprintf("data", "    Head: %v\n", head)

	// List branches
	branches, err := repo.Branches()
	if err != nil {
		return fmt.Errorf("can't get branches: %w", err)
	}
	defer branches.Close()

	win.Fprintf("data", "Branches:\n")
	err = branches.ForEach(func(ref *plumbing.Reference) error {
		extra := ""
		if ref.Hash() == head.Hash() {
			extra = "(current)"
		}
		win.Fprintf("data", "\tCo %s %s\n", ref.Name(), extra)
		return nil
	})
	if err != nil {
		return fmt.Errorf("can't list branches: %w", err)
	}

	win.Ctl("clean")

	return nil
}

func doCheckout(win *acme.Win, repo *git.Repository, cmd string) error {
	// TODO: Handle checking out commits and tags

	parts := strings.Fields(cmd)
	// Check if we have command `Co thing`
	if len(parts) != 2 {
		return fmt.Errorf("unexpected length of command: %d", len(parts))
	}
	if parts[0] != "Co" {
		return fmt.Errorf("called for unexpected command")
	}

	// Check out existing branch
	opts := git.CheckoutOptions{
		Branch: plumbing.ReferenceName(parts[1]),
		Keep:   true,
	}

	tree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("can't get work tree: %w", err)
	}

	err = tree.Checkout(&opts)
	if err != nil {
		return fmt.Errorf("checkout failed: %w", err)
	}

	return refresh(win, repo)
}

func doInteractiveCommit(win *acme.Win, repo *git.Repository, cmd string) error {
	// TODO: make sure the working directory is properly set
	// TODO: convince git commit to use acme as the editor (plumb edit?)
	command := exec.Command("win", "git", "commit", "-s", "--interactive")
	err := command.Run()
	if err != nil {
		return fmt.Errorf("can't run interactive git commit: %w", err)
	}

	return refresh(win, repo)
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
	win.Fprintf("tag", "Get ")

	repo, err := git.PlainOpen(".")
	if err != nil {
		winFatal(win, "can't open repo: %s", err)
	}

	err = refresh(win, repo)
	if err != nil {
		winFatal(win, "can't refresh repo state: %s", err)
	}

	for event := range win.EventChan() {
		code := fmt.Sprintf("%c%c", event.C1, event.C2)

		if event.C1 == 0x00 && event.C2 == 0x00 {
			// Zero event
			break
		}

		switch event.C1 {
		case 'K', 'E', 'F':
			continue
		}

		switch event.C2 {
		case 'i', 'I', 'd', 'D':
			// ignore
		case 'l', 'L':
			err := win.WriteEvent(event)
			if err != nil {
				winFatal(win, "can't write event %#v: %w", event, err)
			}
		case 'x', 'X':
			// TODO: Deal with command args somehow?
			switch true {
			case string(event.Text) == "Get":
				err = refresh(win, repo)
				if err != nil {
					winFatal(win, "can't refresh repo state: %s", err)
				}
				continue
			case bytes.HasPrefix(event.Text, []byte("Co ")):
				log.Println("running checkout command")
				err = doCheckout(win, repo, string(event.Text))
				if err != nil {
					winFatal(win, "can't check out branch: %w", err)
				}
			case bytes.Equal(event.Text, []byte("Ci")):
				log.Println("interactive commit")
				err = doInteractiveCommit(win, repo, string(event.Text))
				if err != nil {
					winFatal(win, "can't run interactive commit: %w", err)
				}
			default:
				log.Printf("Execute: %q", event.Text)
				err := win.WriteEvent(event)
				if err != nil {
					winFatal(win, "can't write event %#v: %w", event, err)
				}
			}
		default:
			log.Printf("got event with code %s: %#v", code, event)
		}
	}
}
