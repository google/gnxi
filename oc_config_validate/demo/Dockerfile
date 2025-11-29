FROM python:3.9-slim-bullseye

RUN apt-get update && apt-get install -y procps
RUN pip3 install oc-config-validate

WORKDIR /opt/oc_config_validate

COPY ./* demo/

RUN mkdir /opt/certs

CMD /opt/oc_config_validate/demo/run_demo.sh -L -N
