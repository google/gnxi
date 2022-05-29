# Demo

1. It starts an instance of [gnmmi_target](../gnmi_target) with a sample
   configuration.

1. It runs `oc_config_validate` with a [sample test file](demo/tests.yaml).

1. It writes the results to [a file](demo/results.json).

## On Docker

Run the demo of `oc_config_validate` in a Docker container:

```
docker pull joseignaciotamayo/oc_config_validate:demo_latest

docker run -it joseignaciotamayo/oc_config_validate:demo_latest

```

The Docker container just clones this repo and runs `demo/run_demo.sh`, on top of a *Debian bullseye* image.

> This Docker image is large because it contains both the Python3 and Golang
> runtimes.

## From source

Run the script `demo/run_demo.sh` on a Linux system to see a quick demonstration of
`oc_config_validate`.

 * `gnmmi_target` needs `golang>=1.17` runtime, install it first.
 * The script assumes `oc_config_validate` and its dependencies are installed.
   If you are using a *virtualenv*, you need to activate it before running
   the demo.

Run `demo/run_demo.sh -h` for additional options to use.

> The script calls [generate.sh](../../certs/generate.sh) to create local TLS
  certificates in [gnxi/certs/](../../certs/).
