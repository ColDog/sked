FROM alpine:3.4
ADD dist/consul-scheduler_linux-amd64 /usr/local/bin/consul-scheduler
ENTRYPOINT ["consul-scheduler", "combined"]
