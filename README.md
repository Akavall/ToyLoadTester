
Build with:

```bash
go build load_tester.go
```

Sample usage:

```bash
./load_tester -ngo=4 -npergo=10 -endpoint=http://localhost:8000/add_up -datafile=temp.json -headersfile=headers.json

```

Sample Output:
```
{max:11 min:4 mean:5.525 statusCodes:map[200:40] n_failed:0}
```

Timings are given in milliseconds


