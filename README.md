# paws

For most up-to-date docs, see "pawsgo for dummies":

* https://confluence.pacificbiosciences.com/display/PA/pawsgo+for+dummies

## Setup
[Help with Go setup](docs/SETUP.md)

## Running

    make serve

(For now, we use port "5000". No reason.)

* http://$HOSTNAME:5000/sockets/cdunn/basecaller

## Update semantic version

Simply modify the VERSION line of the `makefile`. (The git sha1 will be
appended automaically.)
