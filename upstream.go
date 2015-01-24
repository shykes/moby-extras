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
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/BurntSushi/toml"
)

type Manifest struct {
	Sources []Source `toml:"source"`
}

type Source struct {
	Name    string      `toml:"name"`
	Owner   string      `toml:"owner"`
	Url     string      `toml:"url"`
	Branch  string      `toml:"branch"`
	Mapping [][2]string `toml:"mapping"`
}

func main() {
	f, err := os.Open("UPSTREAM")
	if err != nil {
		log.Fatal(err)
	}
	data, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatal(err)
	}

	var m Manifest
	if err := toml.Unmarshal(data, &m); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Loaded %d sources from ./UPSTREAM\n", len(m.Sources))
	for _, src := range m.Sources {
		if src.Name == "" {
			fmt.Printf("skipping unnamed source\n")
			continue
		}
		fmt.Printf("git fetch %s %s:refs/heads/upstream/%s\n", src.Url, src.Branch, src.Name)
	}
}
