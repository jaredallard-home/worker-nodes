#!/usr/bin/env bash

echo "raspberry-pi firmware version"
vcgencmd bootloader_version
if [[ $? -ne 0 ]]; then
  echo "failed to get version, exiting"
  exit 1
fi

# TODO: We should be able to use any process outside of the
# container here, so let's not make this a requirement.
# Also we could add a message about needing pid: host being set.
if ! pgrep balenad >/dev/null 2>&1; then
  echo "Error: Failed to find pid of balenad, unable to update firmware"
fi

echo "updating raspberry-pi firmware version"
apt-get -y update
apt-get install -y rpi-eeprom

rpi-eeprom-update -a -i
if [[ ! -e "/boot/recovery.bin" ]]; then
  echo "Nothing to do."
else 
  echo "moving firmware update to host /boot"
  cp -rv /boot/{pieeprom.sig,pieeprom.upd,recovery.bin,vl805.bin,vl805.sig} "/proc/$(pgrep balenad)/root/mnt/boot/"

  echo "rebooting host device in 20s"
fi

exec sleep infinity