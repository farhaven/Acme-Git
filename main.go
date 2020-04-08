package main

import (
	"log"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

func main() {
	log.Println("here we go")

	repo, err := git.PlainOpen(".")
	if err != nil {
		log.Fatalln("can't open repo:", err)
	}

	log.Println("got repo", repo)

	branches, err := repo.Branches()
	if err != nil {
		log.Fatalln("can't get branches:", err)
	}
	defer branches.Close()

	log.Println("branches:")
	branches.ForEach(func(ref *plumbing.Reference) error {
		log.Println(ref.Name())
		return nil
	})
}