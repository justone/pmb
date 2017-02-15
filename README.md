# Personal Message Bus


It is a little difficult to explain what the Personal Message Bus (PMB) is, so it's perhaps easier to state first what it can help you do:

* **Remote copy** - copy small and large snippets of text from any remote system, no matter how many levels of SSH deep, into your local clipboard
* **Opening URLs in local browser** - send URLs from your remote editor or IRC client and open it locally in your desktop browser, works for html content as well
* **Long running job notification** - run that 20 minute deployment script, kick off that 11Gb database import, or run any other long job and get a notification on your desktop when it is complete
* **Log file streaming** - send log files from a dozen servers and aggregate them locally or on a remote system
* **Remote command coordination** - update a cloud database from one server and then run a database import on another, making sure that the commands are done in order and as soon as possible, receiving notifications for each step
* **Mobile and watch notification** - leave work with a deployment running and receive a message on your watch halfway through dinner that the deployment succeeded
* **File download notification** - download large files and get a notification when the download is complete.

All this, and all communication is encrypted and sent over an SSL connection.  There are no complex firewall holes to punch, everything goes through a [RabbitMQ](http://www.rabbitmq.com/) instance out in the cloud.

To learn more and get started using PMB, head over [here](http://docs.pmb.io/getting_started/).

The rest of this document details how to **hack** on PMB.

# Hacking on PMB

```
$ go get github.com/justone/pmb
# cd $GOPATH/src/github.com/justone/pmb
# go build
```

If you need a nice clean go environment, try out [skeg](http://skeg.io).

# Similar projects

* [remotecopy](https://github.com/justone/remotecopy) - My first foray into copying data from remote systems to my local clipboard.  PMB started because I really liked the power of remotecopy and wanted to apply that power to other problems.
* [remote-pbcopy](https://seancoates.com/blogs/remote-pbcopy/) - The inspiration for remotecopy and, transitively, PMB.
* [lemonade](https://github.com/pocke/lemonade)
* [DoIt](http://www.chiark.greenend.org.uk/~sgtatham/doit/)

# Thank you

This project wouldn't be possible without the following libraries:

* [amqp](https://github.com/streadway/amqp)
* [go-flags](https://github.com/jessevdk/go-flags)
* [loggo](https://github.com/loggo/loggo)
* [osext](http://godoc.org/bitbucket.org/kardianos/osext)

# License

Copyright © 2014-2017 Nate Jones

Distributed under the Apache License Version 2.0.
