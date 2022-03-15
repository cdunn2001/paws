RPM_BASENAME?=pa-wsgo

build-rpm:
	rm -rf opt/
	${MAKE} files
	${MAKE} tar
	${MAKE} rpm
files:
	./generate_rpm_files.py
tar:
	rm -f ${RPM_BASENAME}.tar
	tar -C ./tard -cvf ${RPM_BASENAME}.tar .
	tar tvf ${RPM_BASENAME}.tar
rpm:
	rm -rf ~/rpmbuild
	./tar2rpm.sh --version 0.0.0 --arch x86_64 --target /opt/pacbio/pa-wsgo-0.0.0 ${RPM_BASENAME}.tar
spec:
	./tar2rpm.sh --version 0.0.0 --arch x86_64 --target /opt/pacbio/pa-wsgo-0.0.0 --print ${RPM_BASENAME}.tar > foo.spec
clean:
	rm -rf opt/
	rm -rf /tmp/tar2rpm-*
	rm -f ${RPM_BASENAME}.tar
	rm -f ${RPM_BASENAME}.rpm

# Unfortunately, 'rpmbuild' is not on all pacbio hosts, so this can be
# difficult to test locally.
