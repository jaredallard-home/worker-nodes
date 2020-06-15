#!/usr/bin/env bash

echo ":: Current Raspberry Pi firmware version"
if ! vcgencmd bootloader_version; then
  echo ":: Warning: Failed to get version, exiting"
  exit 0
fi

# Enable here if you want to JIT upgrade to latest firmware
# Otherwise you can use the baked in version
#echo "updating raspberry-pi firmware version"
#apt-get -y update
#apt-get install --no-install-recommends -y rpi-eeprom

rpi-eeprom-update -a -i
if [[ ! -e "/boot/recovery.bin" ]]; then
  echo " :: Nothing to do."
else 
  echo " :: Moving firmware update to host /boot"
  cp -rv /boot/{pieeprom.sig,pieeprom.upd,recovery.bin,vl805.bin,vl805.sig} "/proc/1/root/mnt/boot/"

  echo " :: Rebooting host device in 20s"
  sleep 20

  echo " :: Rebooting"
  DBUS_SYSTEM_BUS_ADDRESS=unix:path=/host/run/dbus/system_bus_socket \
  dbus-send \
  --system \
  --print-reply \
  --dest=org.freedesktop.systemd1 \
  /org/freedesktop/systemd1 \
  org.freedesktop.systemd1.Manager.Reboot
fi

if [[ -n "$DEBUG_MODE" ]]; then
  exec sleep infinity
fi