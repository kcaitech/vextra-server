#!/bin/bash

curl -sSL http://update.aegis.aliyun.com/download/uninstall.sh | bash
curl -sSL http://update.aegis.aliyun.com/download/quartz_uninstall.sh | bash

pkill aliyun-service
rm -fr /etc/init.d/agentwatch /usr/sbin/aliyun-service
rm -rf /usr/local/aegis*
rm -rf /usr/sbin/aliyun-service
rm -rf /usr/sbin/aliyun-service.backup
rm -rf /usr/sbin/aliyun_installer
rm -rf /etc/systemd/system/aliyun.service
rm -rf /lib/systemd/system/aliyun.service

iptables -I INPUT -s 140.205.201.0/28 -j DROP
iptables -I INPUT -s 140.205.201.16/29 -j DROP
iptables -I INPUT -s 140.205.201.32/28 -j DROP
iptables -I INPUT -s 140.205.225.192/29 -j DROP
iptables -I INPUT -s 140.205.225.200/30 -j DROP
iptables -I INPUT -s 140.205.225.184/29 -j DROP
iptables -I INPUT -s 140.205.225.183/32 -j DROP
iptables -I INPUT -s 140.205.225.206/32 -j DROP
iptables -I INPUT -s 140.205.225.205/32 -j DROP
iptables -I INPUT -s 140.205.225.195/32 -j DROP
iptables -I INPUT -s 140.205.225.204/32 -j DROP

systemctl stop AssistDaemon
systemctl disable AssistDaemon
rm -rf /etc/systemd/system/AssistDaemon.service
rm -rf /usr/local/share/assist-daemon
systemctl daemon-reload
