function ReservePort {
    port=$1
    echo -n "    INFO Opening firewall port $port :"
    firewall-cmd --permanent --zone=public --add-port=$port/tcp

    if grep ${port} /proc/sys/net/ipv4/ip_local_reserved_ports > /dev/null
    then
      echo "    INFO port $port reserved already"
    else
      existing=`cat /proc/sys/net/ipv4/ip_local_reserved_ports`
      if [ "$existing" != "" ]
      then
         existing="$existing,"
      fi
      existing="${existing}${port}"
      echo "    INFO reserving port $existing"
      echo $existing > /proc/sys/net/ipv4/ip_local_reserved_ports
    fi
}

#ports="23632"
ports="23633"
for port in $ports
do
  ReservePort $port
done
sysctl net.ipv4.ip_local_reserved_ports > /etc/sysctl.d/98-sequel.conf
echo -n "    INFO Reloading firewall :"
firewall-cmd --reload


if [ ! -d /var/log/pacbio ]
then
  echo "    Creating /var/log/pacbio"
  mkdir -p  /var/log/pacbio
  chmod 777 /var/log/pacbio
  chmod o+t /var/log/pacbio
fi

if [ ! -d /var/run/pacbio ]
then
   echo "    Creating /var/run/pacbio"
   mkdir -p  /var/run/pacbio
   chmod 777 /var/run/pacbio
   chmod o+t /var/run/pacbio
fi

#if grep show_hugepage_usage /etc/sudoers > /dev/null
#then
#  echo "    INFO sudoers list already updated"
#else
#  echo "    INFO adding to sudoers list"
#  echo "ALL ALL=NOPASSWD: $RPM_INSTALL_PREFIX/bin/show_hugepage_usage.sh" >> /etc/sudoers
#fi

#mkdir -p /etc/modulefiles/pacbio-pa-wsgo
#envsubst '$RPM_INSTALL_PREFIX' < $RPM_INSTALL_PREFIX/etc/modulefiles/pacbio-pa-wsgo/0.0.0 > /etc/modulefiles/pacbio-pa-wsgo/0.0.0


if [ ! -e /etc/pacbio/pacbio-pa-wsgo.conf ]
then
   echo "    Creating template /etc/pacbio/pacbio-pa-wsgo.conf, as it does not exist yet."
   mkdir -p /etc/pacbio
   cp -f $RPM_INSTALL_PREFIX/systemd/pacbio-pa-wsgo.conf /etc/pacbio/pacbio-pa-wsgo.conf
fi

if [ ! -e /etc/pacbio/app-common.json ]
then
   echo " Creating empty system configuration file /etc/pacbio/app-common.json"
   echo "{}" > /etc/pacbio/app-common.json
fi

echo "    Installing systemd service pacbio-pa-wsgo-0.0.0.service"
if [[ $RPM_INSTALL_PREFIX != "" ]]
then
    envsubst '$RPM_INSTALL_PREFIX' < $RPM_INSTALL_PREFIX/systemd/pacbio-pa-wsgo-0.0.0.service > /etc/systemd/system/pacbio-pa-wsgo-0.0.0.service
    systemctl daemon-reload
fi    


exit 0
