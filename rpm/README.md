## RPM content
We need the following inside the RPM:

./opt/pacbio/pa-X-VERSION/bin/pa-ws
./opt/pacbio/pa-X-VERSION/bin/precheck-pa-ws.sh
./opt/pacbio/pa-X-VERSION/systemd/pacbio-pa-X-VERSION.service
./opt/pacbio/pa-X-VERSION/systemd/pacbio-pa-X.conf

For more info, see

* https://confluence.pacificbiosciences.com/display/PA/Paws+Deployment

Release Notes


1.0.14 - Fixed the arguments passed to smrt-basecaller involving dark cal and image PSF. Requires smrt-basecaller 0.1.11 at a minimum.
1.0.16 - Implemented working ssh baz2bam
1.0.17 - changed baz2bam threading parameters to -j 128 -b 32
1.0.24 - added crosstalkFilter support
1.0.25 - add support for photoelectronSensitivity, refSnr and analogs
1.0.28 - fix reducestats (w/ 1.0.27)
1.0.29 - send ChipLayout to basecaller
1.0.30 - added support for NUMA node and GPU node settings
1.0.31 - /dashboard endpoint
1.0.32 - skip GET logging
1.1.11 - rtmetrics; lots of features/fixes
1.2.0  - enable multi-baz mode in baz2bam
1.2.1  - store trace files on ICC (not NRT)
1.2.2  - verbose logging
1.2.3  - UTC always; better logging
1.2.4  - include dir-existence in /status, but cached
