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

Caller should 'rm -rf ./opt' before running.
"""
import os, sys

in_files = {
    'systemd/pacbio-pa-X.conf.in': './opt/pacbio/pa-@NAME@-@V@/systemd/pacbio-pa-@NAME@.conf',
    'systemd/pacbio-pa-X.service.in': './opt/pacbio/pa-@NAME@-@V@/systemd/pacbio-pa-@NAME@-@V@.service',
    'systemd/precheck-pa-wsgo.sh.in': './opt/pacbio/pa-@NAME@-@V@/bin/precheck-pa-@NAME@.sh',
}
VERSION = '0.0.0'
NAME = 'wsgo'  # Call it "pa-wsgo" for now.
subs = {
    "@V@": VERSION,
    "@NAME@": NAME,
    "@SYSTEM_EXEC@": "pa-wsgo",
    "@APP_VERSION@": "QAPP_VERSIONQ",
    "@SOFTWARE_VERSION@": "QSOFTWARE_VERSIONQ",
    "@SYSTEMD_DEPENDENCIES@": "",
    "@SYSTEMD_CONF_PATH@": "", #opt/pacbio/pa-@NAME@-@V@/systemd/pacbio-pa-@NAME@.conf
    "@SYSTEMD_PREEXEC1@": "",
    "@SYSTEMD_COMMON_JSON@": "/etc/pacbio/pa-common.json",
    "@SYSTEMD_ALIAS@": "pacbio-pa-wsgo",
}
statics = {
    '../bin/pawsgo': './opt/pacbio/pa-@NAME@-@V@/bin/pa-wsgo', # Note dash.
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
  for ifn, ofn in statics.items():
    ofn = CmakeSub(ofn)
    Copy(ifn, ofn)
def Build():
  SubstituteAll()
  CopyStatics()
  Tar()
  GenerateRpm()
def SubstituteAll():
  for (ifn, ofn) in in_files.items():
    ofn = CmakeSub(ofn)
    Substitute(ifn, ofn)
def CmakeSub(str):
  for at_key, repl in subs.items():
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

if __name__ == "__main__":
  Build()
