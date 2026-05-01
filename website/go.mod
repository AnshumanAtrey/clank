// Stub go.mod — exists solely to exclude this directory from the parent
// Go module (github.com/AnshumanAtrey/clank). Without this, `go install`
// would download website source files unnecessarily. The website is an
// Astro project, not Go code. See path-1/07-repo-architecture.md.
module github.com/AnshumanAtrey/clank/website

go 1.25
