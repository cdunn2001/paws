# TODO: Use this VERSION for rpm stuff too.
NAME=pa-wsgo
VERSION=1.1.7
#RELEASE:=$(shell date --utc +%Y%m%dT%H%M%SZ)
RPM_VERSION:=${VERSION}
RPM_BASENAME:=${NAME}
ARTEFACT:=${RPM_BASENAME}.rpm.tar
RPM_FILENAME:=${RPM_BASENAME}-${VERSION}.x86_64.rpm

build-rpm:
	rm -rf opt/
	${MAKE} files
	${MAKE} tar
	${MAKE} rpm
	${MAKE} artefact
files:
	${MAKE} -C .. release VERSION=${VERSION}
	rm -rf BUILD/tard
	./generate_rpm_files.py ${VERSION} ${RPM_VERSION}
tar:
	BUILD/tard/bin/pa-wsgo --version
	rm -f BUILD/${RPM_BASENAME}.tar
	tar -C BUILD/tard -cvf BUILD/${RPM_BASENAME}.tar .
	tar tvf BUILD/${RPM_BASENAME}.tar
rpm:
	rm -rf ~/rpmbuild
	./tar2rpm.sh --name ${NAME}-${VERSION} --version ${RPM_VERSION} --arch x86_64 -A --target /opt/pacbio/pa-wsgo-${VERSION} BUILD/${RPM_BASENAME}.tar
artefact:
	rm -rf BUILD/tara
	mkdir -p BUILD/tara/rpm
	mv ~/rpmbuild/RPMS/x86_64/*.rpm BUILD/tara/rpm/${RPM_FILENAME}
	tar -C BUILD/tara -cvf BUILD/${ARTEFACT} .
	tar tvf BUILD/${ARTEFACT}
spec:
	./tar2rpm.sh --name ${NAME}-${VERSION} --version ${RPM_VERSION} --arch x86_64 -A --target /opt/pacbio/pa-wsgo-${VERSION} --print BUILD/${RPM_BASENAME}.tar > foo.spec
clean:
	rm -rf opt/
	rm -rf /tmp/tar2rpm-*
	rm -rf BUILD/
	rm -f ${RPM_BASENAME}.tar # old
	rm -f ${RPM_BASENAME}.rpm # old

# Unfortunately, 'rpmbuild' is not on all pacbio hosts, so this can be
# difficult to test locally.
