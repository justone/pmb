# Release instructions

## 1. Run the `build.sh` script to build a new release.

This will generate the binaries and the bootstrap script.

```
$ ./build.sh 
Building 2015-04-14-1952-d7476be

Number of parallel builds: 4

-->    darwin/amd64: _/home/nate/pmb
-->     linux/amd64: _/home/nate/pmb
-->       linux/arm: _/home/nate/pmb
$ find 2015-04-14-1952-d7476be/
2015-04-14-1952-d7476be/
2015-04-14-1952-d7476be/pmb_linux_amd64
2015-04-14-1952-d7476be/pmb_linux_arm
2015-04-14-1952-d7476be/bootstrap
2015-04-14-1952-d7476be/pmb_darwin_amd64
```

## 2. Rsync the release to the <http://get.pmb.io/> server.

```
rsync -var 2015-04-14-1952-d7476be get.pmb.io:/var/path/to/get.pmb.io/
sending incremental file list
2015-04-14-1952-d7476be/
2015-04-14-1952-d7476be/bootstrap
2015-04-14-1952-d7476be/pmb_darwin_amd64
2015-04-14-1952-d7476be/pmb_linux_amd64
2015-04-14-1952-d7476be/pmb_linux_arm

sent 19,470,522 bytes  received 99 bytes  230,421.55 bytes/sec
total size is 19,465,459  speedup is 1.00

```

## 3. Finally, on the get.pmb.io server, update the 'latest' symlink to point to the new release.

```
$ rm latest && ln -s 2015-04-14-1952-d7476be latest
```
