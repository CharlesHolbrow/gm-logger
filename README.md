# mg-logger

The MIDI logger I always wanted.

This will print out the available devices and their device IDs. It will start
listening on the default device (which is chosen by `portmidi`).

```bash
$ go get github.com/CharlesHolbrow/mg-logger
$ go install github.com/CharlesHolbrow/mg-logger
$ # Assuming $GOPATH/bin is in your $PATH
$ mg-logger
```

To specify a non-default device:

```bash
$ mg-logger -device=1
```
