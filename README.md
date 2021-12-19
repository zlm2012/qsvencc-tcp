# TCP Wrapper for QSVEncC

This is a tcp wrapper implementation for QSVEncC command from [QSVEnc](https://github.com/rigaya/QSVEnc),
simply sending the input & args to the server-side, and returning the output to the client
with progress info printed to stderr.

(C) 2021 Lingmo Zhu

日本語は[ここ](./README_ja.md)

## Usage Scenario Considerable

* QSV encoding is preferred but the machine with QSV support is not the same one (Is this a valid case?)
* QSV is preferred in a VM, but it is hard or impossible to access it from the VM due to no GVT-g or 
  reasonable SR-IOV support (i.e. Integrated 11th Gen Xe Graphics)

## Components

* `cmd/qsvencc-server` is the server implementation, with configurable options below
  * `-l --listen`: tcp address to listen, with default value `192.168.122.1:11111`
  * `-c --cmd-path`: path to qsvencc executable, with default value `qsvencc`
* `cmd/qsvencc-client` is the client implementation which could be a drop-in replacement of original qsvencc in the view of arguments, with additional options below
  * `--remote-addr`: tcp address to connect, with default value `192.168.122.1:11111`

Why `192.168.122.1`? Because that's the default IP address for KVM host from default bridge `virbr0`.

## Special Note on KVM

Some configurations need to be tuned for an acceptable experience:

* virtio NIC is strongly recommended if your input is a raw video file / stream.
* at least `mode='host-model'` for the cpu configuration, or the CPU performance could be decreased dramatically.

## License

MIT License is applied. please check [LICENSE](./LICENSE) for detail.