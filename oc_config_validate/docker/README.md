# Docker

`oc_config_validate` can be run in a Docker container, using 3 Volumes to share
files between 3 directories in the container and the host:

 Volume | Path in container
 ------ | ----------------
 oc_config_validate_tests | /opt/oc_config_validate/tests
 oc_config_validate_init_configs | /opt/oc_config_validate/init_configs
 oc_config_validate_results | /opt/oc_config_validate/results


## How to use

 1. Create the 3 Docker Volumes in your host:

 ```
 docker volume create oc_config_validate_tests
 docker volume create oc_config_validate_init_configs
 docker volume create oc_config_validate_results
 ```

 1. Put all your tests and initial configuration files in the volumes.

    * Your test files MUST reference the init_config files as
     `init_configs/<file>.json`.

 1. Get the Docker image and test it can reach the gNMI target:

   ```
    docker run -it \
      -v oc_config_validate_tests:/opt/oc_config_validate/tests \
      -v oc_config_validate_init_configs:/opt/oc_config_validate/init_configs \
      -v oc_config_validate_results:/opt/oc_config_validate/results \
      joseignaciotamayo/oc_config_validate:latest /bin/bash
   ```

   You will likely need to edit your host configuration to allow communication
   between the container and the gNMI Target. Use tools like `ping`, `telnet`
   and `ip route` in the container to verify connectivity.

 1. Run `oc_config_validate` with tests and resultss files, using the relevant
    options. E.g.:

   ```
    docker run -it \
      -v oc_config_validate_tests:/opt/oc_config_validate/tests \
      -v oc_config_validate_init_configs:/opt/oc_config_validate/init_configs \
      -v oc_config_validate_results:/opt/oc_config_validate/results \
      joseignaciotamayo/oc_config_validate:latest \
       python3 -m oc_config_validate --tests_file tests/system.yaml \
        --results_file results/system.json --stop_on_error --log_gnmi
   ```
