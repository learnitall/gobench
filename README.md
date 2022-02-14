# gobench

gobench is a benchmark wrapper utility written in go, based off of [benchmark-wrapper](https://github.com/cloud-bulldozer/benchmark-wrapper).

It helps automate the process of running a benchmark, parsing its results, and exporting those results.

## Installing

gobench can be installed just like any other published go module, using the go CLI. Right now, gobench is under development and still refining its structure, therefore we are still publishing under the `v0` tag:

`$ go install github.com/learnitall/gobench@v0`

gobench can also be built and used through the included `Containerfile`s and `Makefile`, which is based on [podman](https://github.com/containers/podman). The main advantage of this route is that all benchmark-specific dependencies are packaged within the benchmark's `Containerfile`, requiring less setup on your end. For instance, to build gobench's base image (which all benchmark images inherit from), and build the `uperf` image, just run the following from the project root:

`$ make uperf`

Note that the base image contains a lot of cached content and therefore is a relatively large size, `>1.5GB`. This was purposeful, as it allows for very speedy and small benchmark image builds, which are `<100MB`.

By default, all images built using the makefile will be tagged with `latest` and a manifest will be created to support multi-arch builds. To change the default tag, set the `TAG` environment variable:

`$ TAG=mytag make uperf`

For more information, please take a look at the [Makefile].

When running and building from source, please note that `gobench` uses benchmark-specific build tags to optimize build-time when only using a subset of available benchmarks. For instance, if you'd like to test uperf locally, you'd need to add the `uperf` and `uperf_test` build flags to go:

`$ cd benchmarks/uperf && go test -tags uperf,uperf_test .`

## High-Level Structure

gobench is built to run benchmarks, parse the output, and export the results. The main flow of execution is as follows:

1. Perform universal setup tasks, such as parsing flags and arguments and setting the log level.
2. Instantiate exporter objects based on given configuration.
3. Setup each exporter and perform a healthcheck to ensure they are all ready.
4. Kick off the benchmark.
5. If successful, grab the stdout and parse it into marshal-able object(s).
6. Marshal the resulting objects and send the bytes to each configured exporter.
7. Cleanup the benchmark.
8. Cleanup each exporter.
9. Fin.

The role that a user plays in all of this is telling gobench what to do. To explore gobench's universal options and benchmark-specific options, use the `--help` flag.

From a developer's perspective, gobench as a module is structured as follows:

* Root: `Containerfile`s, `Makefile`, `main.go`
* `mappings/`: ElasticSearch index mappings for each benchmark's output.
* `exporters/`: Definition and implementation of each available exporter.
* `define/`: Definition of high-level structs used within gobench.
* `cmd/`: [Cobra](https://github.com/spf13/cobra) based, [viper](https://github.com/spf13/viper) enabled CLI.
* `benchmarks/**`: Definition and implementation of each benchmark supported by gobench.


## Getting Started

To see a list of currently supported benchmarks, run the following and check the list of available commands:

`$ gobench run --help`

Let's run through an example.

Let's say I want to run `uperf` using the default [iperf.xml](https://raw.githubusercontent.com/uperf/uperf/master/workloads/iperf.xml) workload, in a basic localhost-localhost network performance test. We'll first start by cloning gobench and building our uperf container image:

```bash
$ git clone https://github.com/learnitall/gobench
$ cd gobench
$ make uperf
```

Now let's create a quick script to run inside our container and name it `test.sh`:

```bash
#!/bin/bash
# test.sh

# start a uperf worker
uperf -s > /dev/null 2>&1 &
# download the workload
curl -s -LO https://raw.githubusercontent.com/uperf/uperf/master/workloads/iperf.xml

# set uperf env variables
export h=localhost
export proto=tcp
export nthr=3

# run the benchmark
# -p: print results in json
# -q: silence all log output
# --: start uperf args
# iperf.xml: our target workload
# -R: ask uperf to give results in raw format
gobench run uperf \
    -p \
    -q \
    -- \
    iperf.xml \
   -R
```

Finally, we can run our test benchmark and explore with jq:

```bash
podman run --rm -it -v ./test.sh:/opt/test.sh:Z gobench:uperf-latest /opt/test.sh > out.json
cat out.json | jq
```

If you'd like to experiment with exporting results to a EK stack, the `Makefile` comes included with recipes for setting up a local stack with podman. Check out the `local-es`, `local-kb` and `local-cleanup` recipes.

## Development Values

These are the values that gobench strives to maintain during development:

1. Easy to use.
2. Readable code that's easy to modify to, contribute to and debug.
3. Modular nature, enabling support for any benchmark with any export format.
