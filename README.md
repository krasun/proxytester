proxytester
===========

A simple CLI utility to test an HTTP proxy.

## Usage

To run: 

```
go run *.go -proxy <url>  -requests 10
```

The result will look like:

```
Results:
Metric                    Average         P95             Unit
Connect Time              0.04            0.00            seconds
First Byte Time           1.60            1.51            seconds
Total Time                1.60            1.51            seconds
Error Count               0               -               -
Error Rate                0.00            -               %

Status Code Distribution:
Status Code     Count
200             10
```

## License 

`proxytester` is released under [the MIT license](LICENSE).