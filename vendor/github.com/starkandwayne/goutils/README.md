# goutils
S&amp;W Go Utilities

**NOTICE: THIS REPOSITORY IS BEING RETIRED**

You may want to refer to these other repos, which will be
supported moving forward:

- [jhunt/go-log](https://github.com/jhunt/go-log)
- [jhunt/go-ansi](https://github.com/jhunt/go-ansi)

Additionally, `timestamp` and `botta` are both being deprecated
for lack of use.

## How To Use

Get:

```
go get github.com/starkandwayne/goutils/{{pkg}}
```

Import:

```
import "github.com/starkandwayne/goutils/{{pkg}}"
```

## Log

Setup logging with `SetupLogging()`:
 * Type: logging mode to use - file, syslog, console
 * Level: debug, info, error, etc. (See all levels below.)
 * Facility: syslog facility to log to - daemon, misc, etc.
 * File: path to log to file if in file mode.

e.g.:

```
log.SetupLogging(LogConfig{ Type: "console", Level: "warning" })
```

If logging is not setup, then the messages will simply go to `stdout`. If logging cannot be setup for `file` or `syslog`, then the default `stdout` will be used. An error message will print to `stderr` to notify you if this occurs.

Log has the following levels defined:

* Debug
* Info
* Notice
* Warn
* Error
* Crit
* Alert
* Emerg

Usage is the same as `Sprintf`/`Printf` statements - simply append an `f` to the desired level. e.g.:

```
dbug_mesg := "This isn't a bug."
log.Debugf("I really need to know this in debug mode: %s", dbug_mesg)
```
