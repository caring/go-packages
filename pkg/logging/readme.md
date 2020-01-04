### Pretty Printing

The development logger outputs logs in a format usable by (this tool)[https://github.com/maoueh/zap-pretty].

First brew install it
```bash
$ brew install maoueh/tap/zap-pretty
```

Then when you run your service pipe the output to the pretty-printer
```bash
$ ./main | zap-pretty
```
