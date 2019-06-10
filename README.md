# mackerel-check-fireworq
This is a mackerel check plugin for fireworq. This plugin sends an alert when permanent failure occurs in any queue.

## Install
```
mkr plugin install fireworq/mackerel-check-fireworq@<release_tag>
```
You can install manually from [releases](https://github.com/fireworq/mackerel-check-fireworq/releases), too.
## Usage
```
# mackerel-check-fireworq --help
  -host string
        Host (default "localhost")
  -name string
        Name (default "Fireworq")
  -port string
        Port (default "8080")
  -scheme string
        Scheme (default "http")
  -tempfile string
        Temp file name
```

See [official docs](https://mackerel.io/docs/entry/custom-checks) for mackerel configurations.
