# pstree
## Introduction
`pstree` is a small program that shows the process listing (`ps`) as a tree (as the name implies...). It has several options to make selection criteria and to change the output style. I ported the code from FredHucht's [repository](https://github.com/FredHucht/pstree) and added some more functionality.

It should compile under Linux, FreeBSD, OpenBSD, macOS, and Windows.

It uses [gopsutil](https://github.com/shirou/gopsutil) for gathering process information. I may use `ps` later so I can support more Un*x distirbutions, but for now, this is what you get.

## Compiling
* Clone this repository
* `cd` to the repository root
* Type `make build` and the binary will live under `bin` in the repository root
* You will need to manually copy `share/man/man1/pstree.1` to your `$MANPATH`
* If you're using macOS, you can also use homebrew
    * `brew tap gdanko/homebrew`
    * `brew update`
    * `brew install gdanko/homebrew/pstree`

## Usage
```
$ pstree --help
pstree $Revision: 0.6.3 $ by Gary Danko (C) 2025

Usage: pstree [-acUimgtuvw] [--age] [-all] [-C, --color <attr>] [--colorize]
          [-s, --contains <pattern>] [-l, --level <level>]
          [--no-pids] [-o, --order-by <field>] [-p, --pid <pid>]
          [--rainbow] [--user <user> ...]
   or: pstree -V

Display a tree of processes.

      --age               show the age of the process using the format (dd:hh:mm:ss)
      --all               equivalent to -a --age -c -g -m -t
  -a, --arguments         show command line arguments
  -C, --color string      color the process name by given attribute; valid options are: age, cpu, mem;
                          cannot be used with --colorize or --rainbow
      --colorize          add some beautiful color to the pstree output; cannot be used with --color or --rainbow
  -s, --contains string   show only branches containing processes with <pattern> in the command line
  -c, --cpu               show CPU utilization percentage with each process, e.g., (c:0.00%)
  -d, --debug             show debugging data
  -U, --exclude-root      don't show branches containing only root processes; cannot be used with --user
  -h, --help              help for pstree
  -i, --ibm-850           use IBM-850 line drawing characters
  -l, --level int         print tree to <level> level deep
  -m, --memory            show the memory usage with each process, e.g., (m:x.y MiB)
      --no-pids           do not show process IDs
  -o, --order-by string   sort the results by <field>; valid options are: age, cpu, mem, pid, threads, user
  -g, --pgid              show process group IDs
  -p, --pid int32         show only branches containing process <pid>
      --rainbow           please don't; cannot be used with --color or --colorize
  -t, --threads           show the number of threads with each process, e.g., (t:xx)
      --user strings      show only branches containing processes of <user>; this option can be used more than and cannot be used with --exclude-root
  -u, --utf-8             use UTF-8 (Unicode) line drawing characters
  -V, --version           display version information
  -v, --vt-100            use VT-100 line drawing characters
  -w, --wide              wide output, not truncated to window width

Process group leaders are marked with '='.
```
