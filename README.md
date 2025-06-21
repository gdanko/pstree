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
    * `brew tap servicenow/gdanko ssh://git@gitlab.servicenow.net:29418/gary.danko/homebrew.git`
    * `brew update`
    * `brew install servicenow/gdanko/pstree`

## Usage
```
$ pstree --help
pstree $Revision: 0.7.4 $ by Gary Danko (C) 2025

Usage: pstree [OPTIONS]

Display a tree of processes.

Application Options:
  -G, --age                show the age of the process using the format (dd:hh:mm:ss)
  -A, --all                equivalent to -a -c -g -G -m -O -p -t -I; cannot be used with --uid-transitions or --user-transitions
  -a, --arguments          show command line arguments
  -k, --color string       color the process name by given attribute; implies --compact-not; valid options are: age, cpu, mem;
                           cannot be used with --colorize or --rainbow
  -C, --colorize           add some beautiful color to the pstree output; cannot be used with --color or --rainbow
  -n, --compact-not        do not compact identical subtrees in output
  -s, --contains string    show only branches containing processes with <pattern> in the command line; implies --compact-not
  -c, --cpu                show CPU utilization percentage with each process, e.g., (c:0.00%); implies --compact-not
  -d, --debug              show debugging data
  -X, --exclude-root       don't show branches containing only root processes; cannot be used with --user
  -h, --help               help for pstree
  -H, --hide-threads       hide threads, show only processes
  -i, --ibm-850            use IBM-850 line drawing characters
  -l, --level int          print tree to <level> level deep
  -m, --memory             show the memory usage with each process, e.g., (m:x.y MiB); implies --compact-not
  -o, --order-by string    sort the results by <field>; valid options are: age, cpu, mem, pid, threads, user
  -g, --pgid               show process group IDs
  -P, --pid int32          show only branches containing process <pid>
  -r, --rainbow            please don't; cannot be used with --color or --colorize
  -O, --show-owner         show the owner of the process
  -p, --show-pids          show PIDs
  -t, --threads            show the number of threads with each process, e.g., (t:xx)
  -I, --uid-transitions    show processes where the user ID changes from the parent process, e.g., (uid→uid); cannot be used with --user-transitions
      --user strings       show only branches containing processes of <user>; this option can be used more than and cannot be used with --exclude-root
  -U, --user-transitions   show processes where the user changes from the parent process, e.g., (user→user); cannot be used with --uid-transitions or --all
  -u, --utf-8              use UTF-8 (Unicode) line drawing characters
  -V, --version            display version information
  -v, --vt-100             use VT-100 line drawing characters
  -w, --wide               wide output, not truncated to window width

Process group leaders are marked with '=' for ASCII, '¤' for IBM-850, '◆' for VT-100, and '●' for UTF-8.
