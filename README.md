# What is it?

Console progress bar for commands that takes long time to finish and produce too much output (i.e. build commands).

It hides the output from your eyes and shows a nice progress bar instead.

# How it works?

- The position of the progress bar is calculated from the output of previous execution of the command.
- When you run it first time, it saves the output lines and timestamps to a temporary file.
- When you run it again, it uses the timestamp information saved from the previous output to create the progress bar.
- At the end of the execution, it opens the captured log file with your pager.

# Install

```
go install github.com/cenkalti/pb@latest
```

# Usage

Prefix the actual command with `pb`.
```
pb make all
```

# Example run

[![asciicast](https://github.com/cenkalti/pb/blob/main/usage.gif)](https://asciinema.org/a/RV1nMSEhxukhbbqPW9JebwaIO)
