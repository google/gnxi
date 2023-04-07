FROM python:3.9-slim-bullseye

RUN apt-get update && apt-get install -y iproute2 iputils-ping ncat
RUN pip3 install --extra-index-url https://test.pypi.org/simple/ oc-config-validate

WORKDIR /opt/oc_config_validate

VOLUME /opt/oc_config_validate/tests /opt/oc_config_validate/init_configs /opt/oc_config_validate/tests/results

CMD python3 -m oc_config_validate --version
