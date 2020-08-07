# gNXI Test Client

A CLI tool for orchestrating tests against a gNXI target.

## CLI

### Installing

```
go get github.com/google/gnxi/gnxi_tester
go install github.com/google/gnxi/gnxi_tester
```

### Running

Tests are provided in `~/.gnxi.yml`. By default, the following processes & clients will be run: 
- `provision`
- `gnoi_os`
- `gnoi_cert`
- `gnoi_reset`

If `[test_names]` are provided, only those tests are ran.
```
gnxi_tester run [test_names] \ 
--cert certs/ca.crt \
--key certs/ca.key \
--target_name target.com \
--target_address localhost:9339 \
--files "os_path:/path/to/image other_file:/path_to_file"
```

### Files required by service
Files to be passed in the `--files` flag.
#### `gnoi_os`
- `os_path`: Path to OS file used in `install` operation.

