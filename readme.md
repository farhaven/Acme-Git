# ACME-Git

This is a Git UI for the ACME text editor.

## TODO list
- [x] show local branches
- [x] allow checking out other branches
- [ ] easy creation of local branches
- [ ] show remote branches
- [ ] easy checkout of remote branches
	- combine `git checkout origin/foo` and `git checkout -b foo`
- [ ] show stashes
- [ ] allow popping/creating new stashes
- [ ] clean up merged branches
- [ ] tag management:
	- [ ] list (10?) most recent tags
	- [ ] create new (annotated) tags
- [ ] push tags, branches
- [ ] fetch?␣push?
- [ ] notes?
- [ ] Allow launching in subdirectories
	- Walk tree upwards to first `.git`
	- Use that as the repo path
- [x] `Get` command to refresh display
- [ ] Autorefresh on changes to the repo?