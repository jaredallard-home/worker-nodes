FROM balenalib/rpi-raspbian

RUN apt-get update -y \
&&  apt-get install -y --no-install-recommends dbus rpi-eeprom

COPY updater.sh /updater.sh

ENTRYPOINT ["/updater.sh"]
