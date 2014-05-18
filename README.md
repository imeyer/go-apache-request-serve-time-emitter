# Apache Request Time Reporter

## Summary
Takes input from a ```CustomLog``` format of ```%D``` (time to serve request in microseconds), calculates the median and sends it and the number of requests received to the configured [InfluxDB](http://influxdb.org).

## Configuration

This was created on a [CentOS](http://centos.org) system so interpret configuration with that in mind.

1. Add ```INFLUXDB_PASSWORD=password``` to ```/etc/sysconfig/httpd```
1. Add the following section to ```/etc/httpd/conf/httpd.conf```
```
LogFormat "%D" request_time
CustomLog "||/usr/bin/go-apache-request-serve-time-reporter -interval 15s -influxdb-database test -influxdb-host web001.nyc.keymanager.io -influxdb-username test" request_time
```
1. Verify configuration is correct
```
$ sudo service httpd configtest
Syntax OK
```
1. **PROFIT!**

## Getting data from InfluxDB

![InfluxDB graphs via Grafana](http://cl.ly/image/0z40403J2b0p/apache_request_time_graphs.png)

The queries I used in the above graphs (using [Grafana](http://grafana.org)) are as follows.
```
select request_time_median from /apache\..*/
select requests / 15 from /apache\..*/
select go_HeapIdle, go_HeapReleased, go_Alloc from /apache\..*/
```

## Caveats
I'm still changing some things around, finding the right memory stats to measure from Go, etc, etc.
