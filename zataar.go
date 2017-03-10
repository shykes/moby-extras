// NOTE: this prototype is a modification of a previous prototype
// called 5-citizen-kane.
//
// This is an experiment in a new kind of distributed development flow.
// It's inspired by the "one tree per maintainer" model of Linux, and
// informed by the experience of developing Docker at large scale.
//
// We introduce 2 key features:
//
// 1: auto-pull from trusted maintainers
//
// The idea is to declare authoritative upstream repositories for different
// parts of the repository, then to assemble the official repository from these
// various sources. As the project is reorganized, and new maintainers take ownership
// of new subsystems, the map of upstreams changes.
// Then, at regular intervals, new content is PULLED from the upstreams, and a
// new official repo is assembled (with perhaps intermediary steps like integration
// tests etc). This means upstreams are implicitly trusted: unlike regular contributors,
// they do not need to send pull requests.
//
//
// 2: routing of pull requests.
//
// If an upstream is hosted on github, the tool can scan them for inbound pull requests,
// detect which pull requests affect which upstream, and "move" them to the corresponding
// repos.
//
// The result: a meta-rpeo.
//
// When your project grows and becomes a platform - a collection of tools rather than a single
// tool, you face a dilemma: one big repo, or many small ones? There is no easy answer.
//
// There are benefits to having everything in one big repository: easier integration, less
// moving parts when managing dependencies, easier to discover components and express
// structure, etc.
//
// But there are also benefits to breaking up into many small repositories: it makes it gives
// a more distinct identity to each sub-project, makes it more likely that people will discover
// it and evaluate it as a standalone tool. It makes it easier for aspiring contributors to send
// patches, since a small repo with less activity can be less intimidating. It makes it easier
// to attract contributors, because each project gets a "place" which is less crowded: less pull
// requests, less pending issues, less people to interact with. This makes it a more personal and
// welcoming place. It also makes it easier for tools such as "go get", which expect a repository
// to correspond to a single, small package.
//
// The idea of a meta-repo is to get the best of both worlds: you can still break up your platform
// in as many small repos as needed, for community and visibility reasons. Then, using auto-pull,
// you can assemble them all into a single meta-repo which is kept always up-to-date.

package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"path"
	"strings"

	"github.com/spf13/cobra"
)

func main() {
	newTransform().Execute()
}

type Transform struct {
	*cobra.Command
	BranchIn string
	Mapping  [][2]string
}

func newTransform() *Transform {
	t := &Transform{
		Command: &cobra.Command{
			Use: "transform a git branch",
		},
	}
	t.Command.Flags().StringVar(&(t.BranchIn), "input", "master", "input branch")
	t.Command.Run = t.run
	return t
}

func (t *Transform) run(cmd *cobra.Command, args []string) {
	buildid := RandomString()[:4]
	for _, arg := range args {
		words := strings.SplitN(arg, ":", 2)
		if len(words) == 1 {
			words = append(words, "")
		}
		t.Mapping = append(t.Mapping, [2]string{words[0], words[1]})
	}
	if len(t.Mapping) == 0 {
		t.Mapping = [][2]string{{"/", "/"}}
	}

	fmt.Printf("# Starting build %s\n", buildid)
	fmt.Printf("set -e\n")

	dupBranch(t.BranchIn, dstBranch(buildid))
	dupBranch(t.BranchIn, getBaseBranch(buildid))

	// Apply mapping
	baseBranch := getBaseBranch(buildid)
	for mapid, mapping := range t.Mapping {
		var (
			from = path.Clean(mapping[0])
			to   = path.Clean(mapping[1])
		)
		mapBranch := getMapBranch(buildid, mapid)
		dupBranch(baseBranch, mapBranch)
		// 1: apply the source path ('subdirectory-filter')
		if from != "/" {
			zoomIn(mapBranch, from)
		}

		// 2: apply the dst path ('tree-filter')
		if to != "" {
			zoomOut(mapBranch, to)
		}

		// 3: apply the resulting branch as a new layer
		mergeLayer(dstBranch(buildid), mapBranch)
	}

}

func mergeLayer(bottom, top string) {
	fmt.Printf("# mergeLayer(%s, %s)\n", bottom, top)
	fmt.Printf("git checkout '%s' && git merge -X ours '%s' && git checkout '%s' && git merge '%s'\n\n",
		top, bottom, bottom, top,
	)
}

func zoomIn(branch, dir string) {
	fmt.Printf("# zoomIn(%s, %s)\n", branch, dir)
	fmt.Printf("(cd $(git rev-parse --show-toplevel) && git filter-branch -f --subdirectory-filter '%s' '%s')\n\n", dir, branch)
}

func zoomOut(branch, dir string) {
	fmt.Printf("# zoomOut(%s, %s)\n", branch, dir)
	tmp := RandomString()[:8]
	fmt.Printf(`(
	cd $(git rev-parse --show-toplevel) \
	&& git filter-branch -f --tree-filter "mkdir .'%s' && mv * .'%s'/ && mkdir -p '%s' && mv .'%s'/* '%s'/ && rm -r .'%s'" '%s'
)

`, tmp, tmp, dir, tmp, dir, tmp, branch,
	)
}

func dupBranch(src, dst string) {
	fmt.Printf("# dupBranch(%s, %s)\n", src, dst)
	fmt.Printf("{ git branch -D '%s' 2>/dev/null || true; } && git branch -f '%s' '%s'\n\n", dst, dst, src)
}

func dstBranch(buildid string) string {
	return path.Join("citizenkane", buildid, "dst")
}

func getBaseBranch(buildid string) string {
	return path.Join("citizenkane", buildid, "base")
}

func getMapBranch(buildid string, mapid int) string {
	return path.Join("citizenkane", buildid, "map", fmt.Sprintf("%d", mapid))
}

// borrowed from github.com/docker/docker/utils/random.go
func RandomString() string {
	id := make([]byte, 32)

	if _, err := io.ReadFull(rand.Reader, id); err != nil {
		panic(err) // This shouldn't happen
	}
	return hex.EncodeToString(id)
}
