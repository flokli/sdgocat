# sdgocat

A small TCP forwarding proxy, systemd socket activated.

## Test
```console
$ go build ./main.go
$ systemd-socket-activate -l 10022 ./main
```

