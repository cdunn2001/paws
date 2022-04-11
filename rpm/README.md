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