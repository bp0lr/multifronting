# multifronting

A tool written in go to check for valid domain fronting

## Installing

Requires [Go](https://golang.org/dl/)

`go install -v github.com/bp0lr/multifronting@latest`


## How To Use:

Examples: 
- `cat domains\live.com.txt | multifronting -f something.azureedge.net -n yourstring -o output.txt --use-pb`

Options:
```
-f, --fronturl string   your host
-n, --needle string     the string to confirm that fronting works
-o, --output string     Output file to save the results to
-p, --proxy string      Add a HTTP proxy
-u, --testurl string    host to test
    --use-pb            use a progress bar
-v, --verbose           Display extra info about what is going on
-w, --workers int       Workers amount (default 50)
```

## Practical Use

You need to setup your host to response to http/s queries $_POST['op'] = "d3bug".

On this response, the text used with -n need to be displayed to validated domain fronting.

your target list(cat list or -u) and your host (-f), must be used without scheme.
