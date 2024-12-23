Running
--------------1------------
```
git clone
update docker-compose.yaml
docker compose up -d
```
--------------2------------
```
go build -o keepalived_metrics ./main.go
./keepalived_metrics -h
```



# docker-keppalived-metrics
About container keepalived export metrics to Prometheus


# Exposure metrics: RestartCount ContainerStatus VipFound VipChangeCount 


Container status and restart count, vip found and vip change count.


docker-compose.yaml
```
version: '3.8'
services:
  keepalived_exporter:
    image: registry.cn-hangzhou.aliyuncs.com/pro-exporter/promexporter:keepalived
    container_name: keepalived_exporter
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
    network_mode: "host"
    ports:
      - "2112:2112"
    environment:
      - PORT=2112
    command: ["/keepalived_expoter", "-container_name", "keepalived-master,keepalived-master1", "-vip", "192.168.7.123,192.168.7.124"]
```
-container_name #Specify the container name

-vip #keepalived virtual ip


prom/dashboards-keepalived-exporter.json is grafana templates
