#!/usr/bin/env bash
#
# root-normalizer
#
# Attmepts to create a Kubernetes friendly environment. Currently only supports BalenaOS.

bind_mounts=("/etc/kubernetes" "/var/lib/rancher" "/var/lib/kubelet" "/etc/cni/net.d" "/opt/cni/bin" "/etc/ceph")

echo " :: Mounting / as rw"
nsenter -t 1 -m sh -- -c "mount -o remount,rw /"

for bm in "${bind_mounts[@]}"; do
  isMounted=$(nsenter -t 1 -m sh -- -c "mount | grep $bm")
  if [[ -z "$isMounted" ]]; then
    alt_src=$(awk -F ':' '{ print $1 }' <<< "$bm")
    alt_name=$(awk -F ':' '{ print $2 }' <<< "$bm")

    src="$bm"
    dir_name=$(basename "$bm")

    if [[ -n "$alt_name" ]]; then
      dir_name="$alt_name"
      src="$alt_src"
    fi

    # We use PID 1, since we need the "root" mount namespace
    echo " :: Creating '$bm' mount"
    nsenter -t 1 -m sh -- -c "mkdir -p '/mnt/data/$dir_name'; mkdir -p '$src'"
    nsenter -t 1 -m sh -- -c "mount --rbind '/mnt/data/$dir_name' '$src'"
    nsenter -t 1 -m sh -- -c "mount --make-rshared '$src'"
  fi
done 

systemdUnitFile="/lib/systemd/system/balena.service"
needsMountFlagsPatch=$(nsenter -t 1 -m sh -- -c "grep MountFlags=slave '$systemdUnitFile'")
if [[ -n "$needsMountFlagsPatch" ]]; then
  echo " :: Patching balena.service"
  nsenter -t 1 -m sh -- -c "sed -i 's/^MountFlags=slave/MountFlags=shared/' '$systemdUnitFile'"
  # This will restart this container, so we make sure that we do it only once.
  nsenter -t 1 -p -m sh -- -c "systemctl daemon-reload && systemctl restart balena.service"
fi

echo " :: Remounting / as ro"
nsenter -t 1 -m sh -- -c "mount -o remount,ro /"

echo " :: Creating /var/run/docker.sock"
exec socat "UNIX-LISTEN:/proc/1/root/var/run/docker.sock,fork" "UNIX-CONNECT:/proc/1/root/var/run/balena.sock"
