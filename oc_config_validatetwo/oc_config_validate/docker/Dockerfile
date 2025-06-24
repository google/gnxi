FROM python:3.9-slim-bullseye

RUN apt-get update && apt-get install -y iproute2 iputils-ping ncat
RUN pip3 install oc-config-validate

WORKDIR /opt/oc_config_validatetwo

VOLUME /opt/oc_config_validatetwo/tests /opt/oc_config_validatetwo/init_configs /opt/oc_config_validatetwo/tests/results

CMD python3 -m oc_config_validatetwo --version
