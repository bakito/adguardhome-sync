# Install

```bash
mkdir -p /opt/AdGuardHomeSync/

sudo cp adguardhome-sync /opt/AdGuardHomeSync/adguardhome-sync
sudo cp adguardhome-sync.yaml /opt/AdGuardHomeSync/adguardhome-sync.yaml

sudo cp AdGuardHomeSync.service /etc/systemd/system/AdGuardHomeSync.service
sudo systemctl enable AdGuardHomeSync
```