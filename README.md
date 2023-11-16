## wiredrop

As known as Apple has an AirDrop, so why can't we have a wiredrop?  
Yeah, this repo brings us wiredrop! :chaos laugh:

wiredrop is written in Golang, fast and light-weight.  
It's responsible to accept PUT and GET http requests, then forwarding the stream to transmit any files between two peers!

Peers must PUT and GET at the same time (according to the configuration file), to start the file transmission.  
If one peer spends a long time waiting for PUT/GET, the server will kick off the peer. 

wiredrop completely does not cache the file, only forwards the data stream.

![lang](https://img.shields.io/static/v1?label=golang&message=1.18&color=blue)
![lang](https://img.shields.io/static/v1?label=LICENSE&message=MIT&color=blue)
![lang](https://img.shields.io/static/v1?label=wiredrop&message=1.0&color=pink)

### Secret Key
The path of the URL is the secret key you used to send the file to peers.  
~~hmmmm... TOTP/HTOP seems like a good scheme.~~

Example: http://wiredrop.example.com/the/secret/key  
Above, the secret key is `the/secret/key`

### Put File
```bash
$ curl http://wiredrop.example.com/file --upload-file <yourfile>
```

### Receive File

Sure, you can use curl, wget, and any command that can download the file
```bash
$ wget http://wiredrop.example.com/file
$ curl http://wiredrop.example.com/file -O file
```

### LICENSE
Licensed under MIT with ❤.
