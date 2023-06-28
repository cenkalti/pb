# What?

Console progress bar for commands that took long time to finish and produce too much output (i.e. build commands).

It hides the output from your eyes and shows a nice progress bar instead.

The position of the progress bar is calculated from the previous output of the command.
So when you run it first time, it only saves the lines to a temporary file.
When you run it again, it uses the timestamp information saved from the previous output to create the progress bar.
At the end execution, it opens the captured log file with your pager.

# Usage

Prefix the actual command with `pb`.
```
pb make all
```

# Example run

[![asciicast](https://asciinema.org/a/RV1nMSEhxukhbbqPW9JebwaIO.svg)](https://asciinema.org/a/RV1nMSEhxukhbbqPW9JebwaIO)
