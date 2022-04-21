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
