
unalias=`systemctl show -p Id pacbio-pa-@NAME@ | cut -d= -f2`

if systemctl status pacbio-pa-@NAME@-@V@.service >/dev/null
then
  echo "    ERROR pacbio-pa-@NAME@-@V@.service is RUNNING, won't uninstall."
  echo "       Please issue the following commands before uninstalling this package."
  echo "           sudo systemctl stop    pacbio-pa-@NAME@-@V@.service"
  exit 1
#else
#  echo "INFO  pacbio-pa-@NAME@-@V@.service is not running, continuing..."
fi

if [ "$unalias" == "pacbio-pa-@NAME@-@V@.service" -a "$1" == "0" ]
then
    if systemctl is-enabled pacbio-pa-@NAME@>/dev/null 2>&1
    then
      echo "    ERROR pacbio-pa-@NAME@ is ENABLED, won't uninstall."
      echo "        Please issue the following command before uninstalling this package."
      echo "            sudo systemctl disable pacbio-pa-@NAME@-@V@.service "
      exit 1
    else
      echo "    INFO pacbio-pa-@NAME@ is disabled, continuing..."
    fi
fi


if [ "$1" == "0" ]
then
    echo Removing systemd server pacbio-pa-@NAME@-@V@.service
    rm -f /etc/systemd/system/pacbio-pa-@NAME@-@V@.service
    #rm -f /etc/modulefiles/pacbio-pa-@NAME@/@V@
else
    echo Leaving systemd server pacbio-pa-@NAME@-@V@.service
fi

