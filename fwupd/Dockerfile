FROM balenalib/rpi-raspbian:buster-20200503

RUN apt-get update \
&&  apt-get install -y dbus

COPY updater.sh /updater.sh

ENTRYPOINT ["/updater.sh"]
