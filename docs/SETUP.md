Setting up a Go workspace at pacbio
===================================

First, you need a workspace directory, and you set $GOPATH
to that.

With "go mod" I don't think you need to set $GOPATH, but
I'm not sure yet.

I use GOPATH=~/GO

Then,

    cd $GOPATH/src
    git checkout paws

Because of our go.mod definition, this acts like

    $GOPATH/src/pacb.com/seq/paws

Note: If you want to use your bitbucket fork, just create a
git-remote.

    git remote add me ssh://....../cdunn/paws

## Private local repos
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


## Resources

* https://github.com/golang-standards/project-layout
