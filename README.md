# port-check-tool

Quick app to check the status of a tcp port

## How to build

`go build`

## Options

Can either pass a set of command line options or set the env variables

```bash
./port-check-tool -h
Usage of ./port-check-tool:
  -i duration
        The number of seconds to wait before each check  (default 30s)
  -p string
        Port to scan (default "22")
  -t duration
        The number of minutes to continue checking the port (default 5m0s)
```

Using environment variables.

```bash
export PORT=22
export TIMELIMIT=1m
export CHECKINTERVAL=20s
export HOSTS=`<<EOF
127.0.0.1
192.168.1.6
192.168.1.32
foo.bar.net
EOF`
```

Now run the tool

```bash
./port-check-tool
```

## How to run

Example using the default options.

```bash
./port-check-tool <<EOF
127.0.0.1
192.168.1.6
192.168.1.32
EOF
```

Example changing the port number.

```bash
./port-check-tool -p 25 <<EOF
127.0.0.1
192.168.1.6
192.168.1.32
EOF
```

Example changing the port number and the time limit.

```bash
./port-check-tool -p 25 -t 1m <<EOF
127.0.0.1
192.168.1.6
192.168.1.32
foo.bar.net
EOF
```

Example output.

```bash
./port-check-tool -t 1m -i 5s <<EOF
127.0.0.1
192.168.1.6
192.168.1.32
foo.bar.net
EOF

192.168.1.6 is reporting ok
127.0.0.1 is reporting ok
192.168.1.32 is reporting ok
foo.bar.net is down
Finished processing results
```
