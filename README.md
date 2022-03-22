# Thoronic Collector

*Thoronic Collector is a generic log collector which can be run as a CLI application
or as an imported library*

Collector was built as an extension to **Security Center** for importing and
processing security events.

## Install

Installation of Collector is dead-simple - just download and extract the zip
containing the release for your system, and run the binary. Binary releases are built
for Windows, Mac, and Linux platforms.

### Building From Source

If you are building from source, please note that Collector requires Go v1.18 or 
above!

To build Collector from source, simply checkout the repository from Github, cd into the
project source directory. Then, run `make all`. After this, you should have a binary for
your system in the `bin` directory.

### Docker

You can also use Collector via the official Docker container 
[here](https://hub.docker.com/r/thoronic/collector/).

### Running the collector

Collector takes a directory of `.conf` files. These files are JSON configs specifying
an input, a processor pipeline, and outputs.

```shell
./collector start --config <CONFIG-DIRECTORY>
```

### Documentation

Documentation can be found on our [site](http://docs.thoronic.com/collector). Find
something missing? Let us know by filing an issue!

### Issues

Find a bug? Want more features? Find something missing in the documentation? Let us know!
Please don't hesitate to [file an issue](https://github.com/ThoronicLLC/collector/issues/new).
