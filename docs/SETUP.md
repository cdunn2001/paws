# Using Go
Until we have an up-to-date GNU module, use mine:

    export GOROOT=/home/UNIXHOME/cdunn/local/go
    export PATH=$PATH:$GOROOT/bin

You do not actually need to set GOROOT.
And because we use "go mod vendor" (see [vendor sub-dir](../vendor/)), you
do not actually need a "workspace" nor "$GOPATH". But if you
want to try that, read on.

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
  - Very controversial. But our main goal is to separate go code (cmd, pkg, vendor) from non-go-code directories, as well as to separate our own from 3p code.
