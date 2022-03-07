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
	tar cvf ${RPM_BASENAME}.tar ./opt
	tar tvf ${RPM_BASENAME}.tar
rpm:
	./tar2rpm.sh --version 0.0.0 --arch x86_64 --target ${CURDIR}/BUILD ${RPM_BASENAME}.tar
	#rm -rf BUILD/
clean:
	rm -rf opt/
	rm -rf /tmp/tar2rpm-*
	rm -f ${RPM_BASENAME}.tar
	rm -f ${RPM_BASENAME}.rpm

# Unfortunately, 'rpmbuild' is not on all pacbio hosts, so this can be
# difficult to test locally.
