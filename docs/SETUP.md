# Using Go
Until we have an up-to-date GNU module, use mine:

    export GOROOT=/home/UNIXHOME/cdunn/local/go
    export PATH=$PATH:$GOROOT/bin

Because we use "go mod vendor" (see [vendor sub-dir](../vendor/)), you
do not actually need a "workspace" nor "$GOPATH". But if you
want to try that, read on.

## Setting up a Go workspace at pacbio

First, you need a workspace directory, and you set $GOPATH
to that.

I use GOPATH=~/repo/gopath

Then,

    cd $GOPATH/src
    git clone paws...

Because of our go.mod definition, this acts like

    $GOPATH/src/pacb.com/seq/paws

Note: If you want to use your bitbucket fork, just create a
git-remote.

    git remote add me ssh://....../cdunn/paws

### Private local repos
(Note: We use "go mod vendor", which is much simpler.)

Go wants public access to all dependencies. That can be tough
for enterprise development. There are lots of work-arounds,
but the simplest is to have a single pacbio repo with
other pacbio repos as git-submodules. Then create a
"replace" alias in `go.mod.

	replace api/fooclient v0.0.0 => ../api/fooclient
	require api/fooclient v0.0.0

You would "import" those repos as usual.

    import "api/fooclient"

* https://github.com/golang/go/issues/26134#issuecomment-516272405

## vendor directory
We use go >= 1.17, which supports "go mod vendor". To update,

Add imports to github.com etc. into your ".go" files. Then:

    go mod vendor  # to copy current versions from the web into /vendor/ dirs
    go mod tidy    # to remove unused vendor pkgs
    go mod verify  # to double-check

And of course:

    git add .
    git commit

## Resources

* https://github.com/golang-standards/project-layout
  - Very controversial. But our main goal is to separate go code (cmd, pkg, vendor) from non-go-cod directories.
