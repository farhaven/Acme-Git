package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"9fans.net/go/acme"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// winFatal writes a formatted message to win's Error window and calls log.Fatalf
func winFatal(win *acme.Win, msg string, args ...interface{}) {
	err := fmt.Errorf(msg, args...)
	win.Errf("%s", err)
	log.Fatalf("%s", err)
}

func refresh(win *acme.Win, repo *git.Repository) error {
	win.Clear()

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
	win.Fprintf("data", "Head:\n")
	win.Fprintf("data", "\t%s\n", head.Hash())
	win.Fprintf("data", "\t%s\n\n", head.Name())

	// List branches
	win.Fprintf("data", "Local branches:\n")
	branches, err := repo.Branches()
	if err != nil {
		return fmt.Errorf("can't get branches: %w", err)
	}
	defer branches.Close()

	err = branches.ForEach(func(ref *plumbing.Reference) error {
		extra := ""
		if ref.Hash() == head.Hash() {
			extra = " (current)"
		}
		win.Fprintf("data", "\tCo %s%s\n", strings.TrimPrefix(ref.Name().String(), "refs/heads/"), extra)
		return nil
	})
	if err != nil {
		return fmt.Errorf("can't list branches: %w", err)
	}

	// List tags
	win.Fprintf("data", "\nTags: NewTag \n")
	tags, err := repo.Tags()
	if err != nil {
		return fmt.Errorf("can't get tags: %w", err)
	}
	defer tags.Close()

	err = tags.ForEach(func(ref *plumbing.Reference) error {
		// TODO: Show annotations?
		extra := ""
		if ref.Hash() == head.Hash() {
			extra = " (current)"
		}
		win.Fprintf("data", "\tCo %s%s\n", strings.TrimPrefix(ref.Name().String(), "refs/tags/"), extra)
		return nil
	})
	if err != nil {
		return fmt.Errorf("can't list tags: %w", err)
	}

	win.Ctl("clean")

	return nil
}

func listRemotes(win *acme.Win, repo *git.Repository) error {
	// Remote branches
	win.Fprintf("data", "Remote branches:\n")
	remotes, err := repo.Remotes()
	if err != nil {
		return fmt.Errorf("can't get remotes: %w", err)
	}
	for _, remote := range remotes {
		remoteBranches, err := remote.List(&git.ListOptions{})
		if err != nil {
			return fmt.Errorf("can't list remote branches for %s: %w", remote.Config().Name, err)
		}
		for _, ref := range remoteBranches {
			win.Fprintf("data", "\tRCo %s %s\n", remote.Config().Name, ref.Name())
		}
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

func doNewTag(win *acme.Win, repo *git.Repository, cmd string) error {
	// TODO: Create annotated tags

	parts := strings.Fields(cmd)
	// Check if we have command `Co thing`
	if len(parts) != 2 {
		return fmt.Errorf("unexpected length of command: %d", len(parts))
	}
	if parts[0] != "NewTag" {
		return fmt.Errorf("called for unexpected command")
	}

	tagger := object.Signature{
		Name:  "XXX",             // TODO
		Email: "xxx@example.com", // TODO
		When:  time.Now(),
	}

	// TODO: Tag message?
	opts := git.CreateTagOptions{
		Tagger:  &tagger,
		Message: fmt.Sprintf("Tag %s", parts[1]), // TODO: Prompt for message?
	}

	head, err := repo.Head()
	if err != nil {
		return fmt.Errorf("can't get repo head: %w", err)
	}

	tagRef, err := repo.CreateTag(parts[1], head.Hash(), &opts)
	if err != nil {
		return fmt.Errorf("can't create tag: %w", err)
	}

	log.Println("created tag with ref", tagRef)
	return nil
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
		log.Fatalf("can't open ACME window: %s", err)
	}

	wd, err := os.Getwd()
	if err != nil {
		winFatal(win, "can't get working directory: %s", err)
	}

	win.Name("%s/-git", wd)
	win.Fprintf("tag", "Get Remotes ")

	repo, err := git.PlainOpen(".")
	if err != nil {
		winFatal(win, "can't open repo: %s", err)
	}

	err = refresh(win, repo)
	if err != nil {
		winFatal(win, "can't refresh repo state: %s", err)
	}

	for event := range win.EventChan() {
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
			txt := strings.TrimLeft(string(event.Text), " \t")

			// TODO: Deal with command args somehow?
			switch true {
			case txt == "Get":
				err = refresh(win, repo)
				if err != nil {
					winFatal(win, "can't refresh repo state: %s", err)
				}
				continue
			case strings.HasPrefix(txt, "Co "):
				log.Println("running checkout command")
				err = doCheckout(win, repo, string(event.Text))
				if err != nil {
					winFatal(win, "can't check out branch: %w", err)
				}
			case strings.HasPrefix(txt, "Remotes"):
				err = listRemotes(win, repo)
				if err != nil {
					winFatal(win, "can't list remotes: %w", err)
				}
			case strings.HasPrefix(txt, "Ci"):
				log.Println("interactive commit")
				err = doInteractiveCommit(win, repo, string(event.Text))
				if err != nil {
					winFatal(win, "can't run interactive commit: %w", err)
				}
			case strings.HasPrefix(txt, "NewTag "):
				log.Println("running tag command")
				err = doNewTag(win, repo, string(event.Text))
				if err != nil {
					winFatal(win, "can't create new tag: %w", err)
				}
			default:
				log.Printf("Execute: %q", event.Text)
				err := win.WriteEvent(event)
				if err != nil {
					winFatal(win, "can't write event %#v: %w", event, err)
				}
			}
		default:
			code := fmt.Sprintf("%c%c", event.C1, event.C2)

			log.Printf("got event with code %s: %#v", code, event)
		}
	}
}
