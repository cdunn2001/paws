#!/usr/bin/env python3
# Our build machines only have python2, but hey, that's end-of-life!
"""
Goal: Generate RPM for our Go executable

Method:
  * Substitute cmake-style "@FOO@" variables in ".in" files.
  * Put them into the directory structure we want.
  * Also copy static files.
  * 'tar' that directory.
  * Run tar2rpm.sh to generate the '.rpm'.

Caller should 'rm -rf BUILD/tard' (currently hard-coded) before running.
"""
import os, sys

# tar2rpm takes the 'extra' scriptlets from ./BUILD/extra/
IN_FILES = {
    'systemd/pacbio-pa-X.conf.in': 'BUILD/tard/systemd/pacbio-pa-@NAME@.conf',
    'systemd/pacbio-pa-X.service.in': 'BUILD/tard/systemd/pacbio-pa-@NAME@-@V@.service',
    'systemd/precheck-pa-wsgo.sh.in': 'BUILD/tard/bin/precheck-pa-@NAME@.sh',
    'extra/preInstall.sh': './BUILD/extra/preInstall.sh',
    'extra/preUninstall.sh': './BUILD/extra/preUninstall.sh',
    'extra/postInstall.sh': './BUILD/extra/postInstall.sh',
}
NAME = 'wsgo'  # Call it "pa-wsgo" for now. TODO: Use NAME=pa-wsgo
STATICS = {
    '../bin/pawsgo': 'BUILD/tard/bin/pa-wsgo', # Note dash.
}
def Init(version, rpm_version):
  global SUBS
  SUBS = {
    "@V@": version,
    "@RV@": rpm_version,
    "@NAME@": NAME,
    "@SYSTEM_EXEC@": "pa-wsgo",
    "@APP_VERSION@": version,
    "@SOFTWARE_VERSION@": "11.1.0", # Is this 'repository'? I.e. 'sequel-11.1.0'?
    "@SYSTEMD_DEPENDENCIES@": "",
    "@SYSTEMD_CONF_PATH@": "", #opt/pacbio/pa-@NAME@-@RV@/systemd/pacbio-pa-@NAME@.conf
    "@SYSTEMD_PREEXEC1@": "",
    "@SYSTEMD_COMMON_JSON@": "/etc/pacbio/pa-common.json",
    "@SYSTEMD_ALIAS@": "pacbio-pa-wsgo",
  }
def Log(msg):
  print(msg + '\n', file=sys.stderr)
def System(call, nothrow=False):
  Log(call)
  rc = os.system(call)
  if rc:
    raise(f'Go {rc} from "{call}"')
def Copy(ifn, ofn):
  System(f'cp -f {ifn} {ofn}')
def CopyStatics():
  for ifn, ofn in STATICS.items():
    ofn = CmakeSub(ofn)
    Copy(ifn, ofn)
def SubstituteAll():
  for (ifn, ofn) in IN_FILES.items():
    ofn = CmakeSub(ofn)
    Substitute(ifn, ofn)
def CmakeSub(str):
  for at_key, repl in SUBS.items():
    str = str.replace(at_key, repl)
  return str
def MakeDirs(dn):
  try:
    os.makedirs(dn)
  except FileExistsError:
    pass
def Substitute(ifn, ofn):
  content = open(ifn).read()
  substituted = CmakeSub(content)
  MakeDirs(os.path.dirname(ofn))
  with open(ofn, 'w') as fout:
    fout.write(substituted)
def MoveToDirectories():
  pass
def Tar():
  pass
def GenerateRpm():
  pass
def Build(prog, version, rpm_version):
  Init(version, rpm_version)
  SubstituteAll()
  CopyStatics()
  Tar()
  GenerateRpm()

if __name__ == "__main__":
  Build(*sys.argv)
