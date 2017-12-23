# Description

[Prometheus]() exporter for [Electroneum Pool](https://github.com/electroneum/electroneum-pool), 
probably could be used with all cryptonight pool types due to similar API.

Stats being collected:

* Total Hashrate - mh/s
* Network difficulty


# Installation

```
go get github.com/murat1985/cpool_exporter
```

or using Docker container:

```
git clone http://github.com/murat1985/cryptonight_exporeter.git
```

Build local image:
```
docker build . -t cpool_exporeter:local
```

Run it:
```
docker run -d -t -i -e CPOOL_DIAL_ADDR='192.168.1.1;192.168.1.2;192.168.1.3' -e CPOOL_PORT=8117 -p 10333:10333 --name cpool_exporter cpool_exporter:local
```
