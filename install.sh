#!/bin/sh

export TEXTFILE_COLLECTOR_DIR=/var/lib/node_exporter

touch /usr/bin/collect_sormas_metrics
chmod +x /usr/bin/collect_sormas_metrics
echo "*/5 * * * *   root   sh -c '/usr/bin/collect_sormas_metrics'" >> /etc/crontab
cat <<EOF> /usr/bin/collect_sormas_metrics
#!/bin/sh
curl localhost:3014/metrics > $TEXTFILE_COLLECTOR_DIR/sormas_exporter.prom.download && \
mv $TEXTFILE_COLLECTOR_DIR/sormas_exporter.prom.download $TEXTFILE_COLLECTOR_DIR/sormas_exporter.prom 
EOF