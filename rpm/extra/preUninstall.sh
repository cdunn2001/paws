
unalias=`systemctl show -p Id pacbio-pa-wsgo | cut -d= -f2`

if systemctl status pacbio-pa-wsgo-0.0.0.service >/dev/null
then
  echo "    ERROR pacbio-pa-wsgo-0.0.0.service is RUNNING, won't uninstall."
  echo "       Please issue the following commands before uninstalling this package."
  echo "           sudo systemctl stop    pacbio-pa-wsgo-0.0.0.service"
  exit 1
#else
#  echo "INFO  pacbio-pa-wsgo-0.0.0.service is not running, continuing..."
fi

if [ "$unalias" == "pacbio-pa-wsgo-0.0.0.service" -a "$1" == "0" ]
then
    if systemctl is-enabled pacbio-pa-wsgo>/dev/null 2>&1
    then
      echo "    ERROR pacbio-pa-wsgo is ENABLED, won't uninstall."
      echo "        Please issue the following command before uninstalling this package."
      echo "            sudo systemctl disable pacbio-pa-wsgo-0.0.0.service "
      exit 1
    else
      echo "    INFO pacbio-pa-wsgo is disabled, continuing..."
    fi
fi


if [ "$1" == "0" ]
then
    echo Removing systemd server pacbio-pa-wsgo-0.0.0.service
    rm -f /etc/systemd/system/pacbio-pa-wsgo-0.0.0.service
    #rm -f /etc/modulefiles/pacbio-pa-wsgo/0.0.0
else
    echo Leaving systemd server pacbio-pa-wsgo-0.0.0.service
fi
