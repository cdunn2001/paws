# TODO: 'module load golang' # 1.17+
#export GOPATH=/home/UNIXHOME/cdunn/repo/gopath
export GOROOT=/home/UNIXHOME/cdunn/local/go
export PATH=$GOROOT/bin:$PATH

# These do not work in Bamboo on vm-styx* machines. Asking Mj...
set +vx
type module >& /dev/null || . /mnt/software/Modules/current/init/bash
set -vx
module use /mnt/software/modulefiles
module unload python
module load python/3
