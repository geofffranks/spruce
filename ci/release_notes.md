#Bug Fixes

- Starting in 1.4.1, error messages were mistakenly sent to os.Stdout, making
  lots of things difficult to detect.
- Fixed issues where colorized output does not appear for errors, when stdout of
  spruce was redirected to a file.
