# Improvements

- `spruce merge` can now read input from STDIN, if you are piping
  another program's YAML output to it. You can make use of it anywhere
  in the merge order using the `-` filename. Additionally, if no
  files are mentioned, and data is being piped in, spruce will
  happily read the data from STDIN, process operators, and provide
  the output.

  Examples:

      echo my: value | spruce merge
      echo my: value | spruce merge -
      echo my: value | spruce merge first.yml - last.yml
      echo my: value | spruce merge first.yml last.yml -

