# Local Filesystem Hook for logrus

[![GoDoc](http://godoc.org/github.com/skiy/skyhook?status.svg)](http://godoc.org/github.com/skiy/skyhook)

Sometimes developers like to write directly to a file on the filesystem. This is a hook for [`logrus`](https://github.com/sirupsen/logrus) which designed to allow users to do that. The log levels are dynamic at instantiation of the hook, so it is capable of logging at some or all levels.

### Note:
User who run the go application must have read/write permissions to the selected log files. If the files do not exists yet, then user must have permission to the target directory.
