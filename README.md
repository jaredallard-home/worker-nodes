# worker-nodes

This contains everything needed to start a new triton kubernetes cluster, when it comes to worker nodes. These nodes will automatically register themselves with Wireguard access and join a Kubernetes cluster. 

We use Balena as our operating system and control platform for managing these nodes.

## Handy Snippets

Reset a Node:

```bash
balena stop $(balena ps | grep rancher | awk '{ print $1 }') && balena rm $(balena ps -aq) && balena volume rm $(balena volume ls -q); \
rm -rf /mnt/data/*; reboot
```

Allow IP Forwarding on the Server Node:

```
# general allow
iptables -A INPUT -m conntrack --ctstate RELATED,ESTABLISHED -j ACCEPT
iptables -A FORWARD -m conntrack --ctstate RELATED,ESTABLISHED -j ACCEPT

# allow wireguard
iptables -A INPUT -p udp -m udp --dport 51820 -m conntrack --ctstate NEW -j ACCEPT

# Allow forwarding for wg0 -> wg0
iptables -A FORWARD -i wg0 -o wg0 -m conntrack --ctstate NEW -j ACCEPT

# Configure NAT
iptables -t nat -A POSTROUTING -s 10.10.0.0/24 -o eth0 -j MASQUERADE
iptables -t nat -A POSTROUTING -s 10.42.0.0/16 -o eth0 -j MASQUERADE
```

## License

MIT