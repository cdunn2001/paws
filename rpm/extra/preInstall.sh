# To debug RPM installation, uncomment this line. $1 is the package count
# echo "    INFO dollar 1 is $1"
# if $1 == 1, this is a first installation.
# if $1 == 2, this is an upgrade.


if systemctl status pacbio-pa-ws-0.0.0.service >/dev/null
then
  echo "    ERROR pacbio-pa-ws-0.0.0.service is RUNNING, won't install."
  echo "       Please issue the following commands before installing this package."
  echo "           sudo systemctl stop    pacbio-pa-ws-0.0.0.service"
  exit 1
fi

